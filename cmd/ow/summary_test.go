package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/huckleberry-1881/ohgmas-watch/pkg/task"
)

func TestGetTimeFilters(t *testing.T) {
	t.Parallel()

	earliest := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	latest := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	customStart := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	customFinish := time.Date(2024, 6, 30, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name       string
		start      *time.Time
		finish     *time.Time
		earliest   time.Time
		latest     time.Time
		wantStart  time.Time
		wantFinish time.Time
	}{
		{
			name:       "nil start and finish uses defaults",
			start:      nil,
			finish:     nil,
			earliest:   earliest,
			latest:     latest,
			wantStart:  earliest,
			wantFinish: latest,
		},
		{
			name:       "custom start with nil finish",
			start:      &customStart,
			finish:     nil,
			earliest:   earliest,
			latest:     latest,
			wantStart:  customStart,
			wantFinish: latest,
		},
		{
			name:       "nil start with custom finish",
			start:      nil,
			finish:     &customFinish,
			earliest:   earliest,
			latest:     latest,
			wantStart:  earliest,
			wantFinish: customFinish,
		},
		{
			name:       "both custom start and finish",
			start:      &customStart,
			finish:     &customFinish,
			earliest:   earliest,
			latest:     latest,
			wantStart:  customStart,
			wantFinish: customFinish,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotStart, gotFinish := getTimeFilters(tt.start, tt.finish, tt.earliest, tt.latest)

			if !gotStart.Equal(tt.wantStart) {
				t.Errorf("getTimeFilters() start = %v, want %v", gotStart, tt.wantStart)
			}

			if !gotFinish.Equal(tt.wantFinish) {
				t.Errorf("getTimeFilters() finish = %v, want %v", gotFinish, tt.wantFinish)
			}
		})
	}
}

func TestGetWeeklySummaries(t *testing.T) {
	t.Parallel()

	now := time.Now()
	weekStart := getMondayOfWeek(now)

	watch := &task.Watch{
		Tasks: []*task.Task{
			{
				Name:     "Test Task",
				Tags:     []string{"tag1"},
				Category: "work",
				Segments: []*task.Segment{
					{Create: weekStart.Add(time.Hour), Finish: weekStart.Add(2 * time.Hour)},
				},
			},
		},
	}

	weekStarts := []time.Time{weekStart}

	// Test without tasks.
	summaries := getWeeklySummaries(watch, weekStarts, false)
	if len(summaries) == 0 {
		t.Error("getWeeklySummaries() returned empty summaries")
	}

	// Test with tasks.
	summariesWithTasks := getWeeklySummaries(watch, weekStarts, true)
	if len(summariesWithTasks) == 0 {
		t.Error("getWeeklySummaries() with tasks returned empty summaries")
	}
}

func TestLoadWatchForSummary(t *testing.T) {
	t.Parallel()

	t.Run("loads from custom file path", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test-tasks.yaml")

		// Create a valid watch and save it.
		originalWatch := &task.Watch{
			Tasks: []*task.Task{
				{Name: "Test Task", Category: "work"},
			},
		}

		err := originalWatch.SaveTasksToFile(filePath)
		if err != nil {
			t.Fatalf("Failed to save test file: %v", err)
		}

		// Load it back.
		watch, err := loadWatchForSummary(filePath)
		if err != nil {
			t.Errorf("loadWatchForSummary() error = %v", err)
		}

		if len(watch.Tasks) != 1 {
			t.Errorf("loadWatchForSummary() loaded %d tasks, want 1", len(watch.Tasks))
		}
	})

	t.Run("returns error for invalid file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "invalid.yaml")

		// Write invalid YAML.
		err := os.WriteFile(filePath, []byte("invalid: yaml: ["), 0600)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		_, err = loadWatchForSummary(filePath)
		if err == nil {
			t.Error("loadWatchForSummary() should return error for invalid YAML")
		}
	})

	t.Run("loads from non-existent file creates empty watch", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "nonexistent.yaml")

		watch, err := loadWatchForSummary(filePath)
		if err != nil {
			t.Errorf("loadWatchForSummary() unexpected error = %v", err)
		}

		if watch == nil {
			t.Fatal("loadWatchForSummary() returned nil watch")
		}

		if len(watch.Tasks) != 0 {
			t.Errorf("loadWatchForSummary() should have 0 tasks for non-existent file, got %d", len(watch.Tasks))
		}
	})
}

// captureStdout captures stdout output from a function call.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close() //nolint:errcheck // test helper, pipe close won't fail

	os.Stdout = oldStdout

	var buf bytes.Buffer

	_, _ = io.Copy(&buf, r)

	return buf.String()
}

func TestPrintWeeklySummaries(t *testing.T) { //nolint:paralleltest // stdout capture
	// Cannot run in parallel due to stdout capture.
	weekStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	summaries := []task.WeeklySummary{
		{
			WeekStart: weekStart,
			Tagsets: []task.TagsetSummary{
				{
					Tagset:   "tag1, tag2",
					Duration: 2*time.Hour + 30*time.Minute,
					Tasks: []*task.Task{
						{
							Name: "Test Task",
							Segments: []*task.Segment{
								{Create: weekStart.Add(time.Hour), Finish: weekStart.Add(3*time.Hour + 30*time.Minute)},
							},
						},
					},
				},
			},
		},
	}

	output := captureStdout(t, func() {
		printWeeklySummaries(summaries, false)
	})

	// Verify output contains expected content.
	if output == "" {
		t.Error("printWeeklySummaries() produced no output")
	}

	expectedContent := "Week starting 01/15/2024"
	if !strings.Contains(output, expectedContent) {
		t.Errorf("printWeeklySummaries() output missing expected content %q", expectedContent)
	}
}

func TestPrintWeeklySummaries_WithTasks(t *testing.T) { //nolint:paralleltest // stdout capture
	// Cannot run in parallel due to stdout capture.
	weekStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	summaries := []task.WeeklySummary{
		{
			WeekStart: weekStart,
			Tagsets: []task.TagsetSummary{
				{
					Tagset:   "development",
					Duration: time.Hour,
					Tasks: []*task.Task{
						{
							Name: "Feature Implementation",
							Segments: []*task.Segment{
								{Create: weekStart.Add(time.Hour), Finish: weekStart.Add(2 * time.Hour)},
							},
						},
					},
				},
			},
		},
	}

	output := captureStdout(t, func() {
		printWeeklySummaries(summaries, true)
	})

	// Verify output contains task details.
	if !strings.Contains(output, "Feature Implementation") {
		t.Error("printWeeklySummaries() with tasks should include task names")
	}
}

func TestPrintTasksForTagset(t *testing.T) { //nolint:paralleltest // stdout capture
	// Cannot run in parallel due to stdout capture.
	weekStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	tasks := []*task.Task{
		{
			Name: "Task A",
			Segments: []*task.Segment{
				{Create: weekStart.Add(time.Hour), Finish: weekStart.Add(2 * time.Hour)},
			},
		},
		{
			Name: "Task B",
			Segments: []*task.Segment{
				{Create: weekStart.Add(3 * time.Hour), Finish: weekStart.Add(4*time.Hour + 30*time.Minute)},
			},
		},
	}

	output := captureStdout(t, func() {
		printTasksForTagset(weekStart, tasks)
	})

	// Verify output contains both tasks.
	if !strings.Contains(output, "Task A") {
		t.Error("printTasksForTagset() output should contain Task A")
	}

	if !strings.Contains(output, "Task B") {
		t.Error("printTasksForTagset() output should contain Task B")
	}

	// Verify duration format.
	if !strings.Contains(output, "[1h00m]") {
		t.Error("printTasksForTagset() output should contain formatted duration for Task A")
	}
}

func TestGenerateSummary_NoSegments(t *testing.T) { //nolint:paralleltest // stdout capture
	// Cannot run in parallel due to stdout capture.
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "empty-tasks.yaml")

	// Create an empty watch.
	watch := &task.Watch{Tasks: []*task.Task{}}

	err := watch.SaveTasksToFile(filePath)
	if err != nil {
		t.Fatalf("Failed to save test file: %v", err)
	}

	var genErr error

	output := captureStdout(t, func() {
		genErr = generateSummary(false, nil, nil, filePath)
	})

	if genErr != nil {
		t.Errorf("generateSummary() error = %v", genErr)
	}

	if !strings.Contains(output, "No segments found") {
		t.Error("generateSummary() should print 'No segments found' for empty tasks")
	}
}

// generateSummaryTestHelper creates a watch with a single task, saves it, and runs generateSummary.
func generateSummaryTestHelper(t *testing.T, taskName, tag string, includeTasks bool) string {
	t.Helper()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "tasks.yaml")

	baseTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	watch := &task.Watch{
		Tasks: []*task.Task{
			{
				Name:     taskName,
				Tags:     []string{tag},
				Category: "work",
				Segments: []*task.Segment{
					{Create: baseTime, Finish: baseTime.Add(2 * time.Hour)},
				},
			},
		},
	}

	err := watch.SaveTasksToFile(filePath)
	if err != nil {
		t.Fatalf("Failed to save test file: %v", err)
	}

	var genErr error

	output := captureStdout(t, func() {
		genErr = generateSummary(includeTasks, nil, nil, filePath)
	})

	if genErr != nil {
		t.Errorf("generateSummary() error = %v", genErr)
	}

	return output
}

func TestGenerateSummary_WithSegments(t *testing.T) { //nolint:paralleltest // stdout capture
	output := generateSummaryTestHelper(t, "Test Task", "feature", false)

	if !strings.Contains(output, "Week starting") {
		t.Error("generateSummary() should print week information")
	}
}

func TestGenerateSummary_WithTimeFilter(t *testing.T) { //nolint:paralleltest // stdout capture
	// Cannot run in parallel due to stdout capture.
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "tasks-filtered.yaml")

	jan15 := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	feb15 := time.Date(2024, 2, 15, 10, 0, 0, 0, time.UTC)

	watch := &task.Watch{
		Tasks: []*task.Task{
			{
				Name:     "January Task",
				Tags:     []string{"january"},
				Category: "work",
				Segments: []*task.Segment{
					{Create: jan15, Finish: jan15.Add(time.Hour)},
				},
			},
			{
				Name:     "February Task",
				Tags:     []string{"february"},
				Category: "work",
				Segments: []*task.Segment{
					{Create: feb15, Finish: feb15.Add(time.Hour)},
				},
			},
		},
	}

	err := watch.SaveTasksToFile(filePath)
	if err != nil {
		t.Fatalf("Failed to save test file: %v", err)
	}

	// Filter to only include January.
	filterStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	filterFinish := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	var genErr error

	output := captureStdout(t, func() {
		genErr = generateSummary(false, &filterStart, &filterFinish, filePath)
	})

	if genErr != nil {
		t.Errorf("generateSummary() error = %v", genErr)
	}

	// Should include January week.
	if !strings.Contains(output, "01/15/2024") {
		t.Error("generateSummary() with filter should include January week")
	}
}

func TestGenerateSummary_InvalidFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML.
	err := os.WriteFile(filePath, []byte("invalid: yaml: ["), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err = generateSummary(false, nil, nil, filePath)
	if err == nil {
		t.Error("generateSummary() should return error for invalid file")
	}
}

func TestGenerateSummary_WithTasks(t *testing.T) { //nolint:paralleltest // stdout capture
	output := generateSummaryTestHelper(t, "Detailed Task", "feature", true)

	if !strings.Contains(output, "Detailed Task") {
		t.Error("generateSummary() with includeTasks should include task names")
	}
}
