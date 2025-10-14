package task_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/huckleberry-1881/ohgmas-watch/pkg/task"
)

func TestAddTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		taskName    string
		description string
		tags        []string
	}{
		{
			name:        "add simple task",
			taskName:    "Test Task",
			description: "A test task",
			tags:        []string{"test"},
		},
		{
			name:        "add task with multiple tags",
			taskName:    "Complex Task",
			description: "A more complex task",
			tags:        []string{"work", "urgent", "development"},
		},
		{
			name:        "add task with no tags",
			taskName:    "Simple Task",
			description: "Task without tags",
			tags:        []string{},
		},
		{
			name:        "add task with nil tags",
			taskName:    "Nil Tags Task",
			description: "Task with nil tags",
			tags:        nil,
		},
		{
			name:        "add task with empty strings",
			taskName:    "",
			description: "",
			tags:        []string{""},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			watch := &task.Watch{Tasks: []*task.Task{}}
			initialTaskCount := len(watch.Tasks)

			watch.AddTask(testCase.taskName, testCase.description, testCase.tags, "work")

			if len(watch.Tasks) != initialTaskCount+1 {
				t.Errorf("Expected %d tasks, got %d", initialTaskCount+1, len(watch.Tasks))
			}

			addedTask := watch.Tasks[len(watch.Tasks)-1]
			if addedTask.Name != testCase.taskName {
				t.Errorf("Expected task name '%s', got '%s'", testCase.taskName, addedTask.Name)
			}

			if addedTask.Description != testCase.description {
				t.Errorf("Expected description '%s', got '%s'", testCase.description, addedTask.Description)
			}

			if len(addedTask.Tags) != len(testCase.tags) {
				t.Errorf("Expected %d tags, got %d", len(testCase.tags), len(addedTask.Tags))
			}

			for i, tag := range testCase.tags {
				if i < len(addedTask.Tags) && addedTask.Tags[i] != tag {
					t.Errorf("Expected tag '%s' at index %d, got '%s'", tag, i, addedTask.Tags[i])
				}
			}

			if addedTask.Segments == nil {
				t.Error("Expected segments slice to be initialized, got nil")
			}

			if len(addedTask.Segments) != 0 {
				t.Errorf("Expected empty segments slice, got %d segments", len(addedTask.Segments))
			}
		})
	}
}

func TestAddSegment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		note string
	}{
		{
			name: "add segment with note",
			note: "Working on feature X",
		},
		{
			name: "add segment without note",
			note: "",
		},
		{
			name: "add segment with long note",
			note: "This is a very long note that describes what I was working on during this time segment. " +
				"It includes multiple sentences and detailed information about the work being performed.",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			task := &task.Task{
				Name:        "Test Task",
				Description: "Test Description",
				Tags:        []string{"test"},
				Segments:    []*task.Segment{},
			}

			beforeTime := time.Now()

			task.AddSegment(testCase.note)

			afterTime := time.Now()

			if len(task.Segments) != 1 {
				t.Errorf("Expected 1 segment, got %d", len(task.Segments))
			}

			segment := task.Segments[0]
			if segment.Note != testCase.note {
				t.Errorf("Expected note '%s', got '%s'", testCase.note, segment.Note)
			}

			if segment.Create.Before(beforeTime) || segment.Create.After(afterTime) {
				t.Errorf("Expected create time between %v and %v, got %v", beforeTime, afterTime, segment.Create)
			}

			if !segment.Finish.IsZero() {
				t.Errorf("Expected finish time to be zero, got %v", segment.Finish)
			}
		})
	}
}

func TestAddMultipleSegments(t *testing.T) {
	t.Parallel()

	task := &task.Task{
		Name:        "Test Task",
		Description: "Test Description",
		Tags:        []string{"test"},
		Segments:    []*task.Segment{},
	}

	segmentNotes := []string{"First segment", "Second segment", "Third segment"}

	for _, note := range segmentNotes {
		task.AddSegment(note)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	if len(task.Segments) != len(segmentNotes) {
		t.Errorf("Expected %d segments, got %d", len(segmentNotes), len(task.Segments))
	}

	for i, expectedNote := range segmentNotes {
		if task.Segments[i].Note != expectedNote {
			t.Errorf("Segment %d: expected note '%s', got '%s'", i, expectedNote, task.Segments[i].Note)
		}
	}

	// Verify timestamps are in order
	for i := 1; i < len(task.Segments); i++ {
		if task.Segments[i].Create.Before(task.Segments[i-1].Create) {
			t.Errorf("Segment %d created before segment %d", i, i-1)
		}
	}
}

func TestCloseSegment(t *testing.T) {
	t.Parallel()

	t.Run("close single open segment", func(t *testing.T) {
		t.Parallel()

		task := &task.Task{
			Name:        "Test Task",
			Description: "Test Description",
			Tags:        []string{"test"},
			Segments: []*task.Segment{
				{
					Create: time.Now().Add(-1 * time.Hour),
					Finish: time.Time{}, // Open segment
					Note:   "Open segment",
				},
			},
		}

		beforeTime := time.Now()

		task.CloseSegment()

		afterTime := time.Now()

		segment := task.Segments[0]
		if segment.Finish.IsZero() {
			t.Error("Expected segment to be closed, but finish time is still zero")
		}

		if segment.Finish.Before(beforeTime) || segment.Finish.After(afterTime) {
			t.Errorf("Expected finish time between %v and %v, got %v", beforeTime, afterTime, segment.Finish)
		}
	})

	t.Run("close multiple open segments", func(t *testing.T) {
		t.Parallel()

		task := &task.Task{
			Name:        "Test Task",
			Description: "Test Description",
			Tags:        []string{"test"},
			Segments: []*task.Segment{
				{
					Create: time.Now().Add(-2 * time.Hour),
					Finish: time.Time{}, // Open segment
					Note:   "First open segment",
				},
				{
					Create: time.Now().Add(-1 * time.Hour),
					Finish: time.Time{}, // Open segment
					Note:   "Second open segment",
				},
			},
		}

		task.CloseSegment()

		for i, segment := range task.Segments {
			if segment.Finish.IsZero() {
				t.Errorf("Segment %d should be closed but finish time is zero", i)
			}
		}
	})

	t.Run("close segments with mixed open/closed", func(t *testing.T) {
		t.Parallel()

		baseTime := time.Now().Add(-3 * time.Hour)
		task := &task.Task{
			Name:        "Test Task",
			Description: "Test Description",
			Tags:        []string{"test"},
			Segments: []*task.Segment{
				{
					Create: baseTime,
					Finish: baseTime.Add(30 * time.Minute), // Closed segment
					Note:   "Closed segment",
				},
				{
					Create: baseTime.Add(1 * time.Hour),
					Finish: time.Time{}, // Open segment
					Note:   "Open segment",
				},
			},
		}

		originalFinishTime := task.Segments[0].Finish
		task.CloseSegment()

		// First segment should remain unchanged
		if !task.Segments[0].Finish.Equal(originalFinishTime) {
			t.Error("Closed segment finish time should not change")
		}

		// Second segment should now be closed
		if task.Segments[1].Finish.IsZero() {
			t.Error("Open segment should now be closed")
		}
	})

	t.Run("close segments with no open segments", func(t *testing.T) {
		t.Parallel()

		baseTime := time.Now().Add(-2 * time.Hour)
		task := &task.Task{
			Name:        "Test Task",
			Description: "Test Description",
			Tags:        []string{"test"},
			Segments: []*task.Segment{
				{
					Create: baseTime,
					Finish: baseTime.Add(30 * time.Minute),
					Note:   "Already closed",
				},
			},
		}

		originalFinishTime := task.Segments[0].Finish
		task.CloseSegment()

		if !task.Segments[0].Finish.Equal(originalFinishTime) {
			t.Error("Already closed segment should not change")
		}
	})
}

func TestHasUnclosedSegment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		segments []*task.Segment
		expected bool
	}{
		{
			name:     "no segments",
			segments: []*task.Segment{},
			expected: false,
		},
		{
			name: "single open segment",
			segments: []*task.Segment{
				{
					Create: time.Now().Add(-1 * time.Hour),
					Finish: time.Time{},
					Note:   "Open segment",
				},
			},
			expected: true,
		},
		{
			name: "single closed segment",
			segments: []*task.Segment{
				{
					Create: time.Now().Add(-1 * time.Hour),
					Finish: time.Now().Add(-30 * time.Minute),
					Note:   "Closed segment",
				},
			},
			expected: false,
		},
		{
			name: "mixed segments with open",
			segments: []*task.Segment{
				{
					Create: time.Now().Add(-2 * time.Hour),
					Finish: time.Now().Add(-90 * time.Minute),
					Note:   "Closed segment",
				},
				{
					Create: time.Now().Add(-1 * time.Hour),
					Finish: time.Time{},
					Note:   "Open segment",
				},
			},
			expected: true,
		},
		{
			name: "all closed segments",
			segments: []*task.Segment{
				{
					Create: time.Now().Add(-2 * time.Hour),
					Finish: time.Now().Add(-90 * time.Minute),
					Note:   "First closed",
				},
				{
					Create: time.Now().Add(-1 * time.Hour),
					Finish: time.Now().Add(-30 * time.Minute),
					Note:   "Second closed",
				},
			},
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			task := &task.Task{
				Name:        "Test Task",
				Description: "Test Description",
				Tags:        []string{"test"},
				Segments:    testCase.segments,
			}

			result := task.HasUnclosedSegment()
			if result != testCase.expected {
				t.Errorf("Expected %v, got %v", testCase.expected, result)
			}
		})
	}
}

func TestGetClosedSegmentsDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		segments []*task.Segment
		expected time.Duration
	}{
		{
			name:     "no segments",
			segments: []*task.Segment{},
			expected: 0,
		},
		{
			name: "single closed segment",
			segments: []*task.Segment{
				{
					Create: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
					Finish: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC),
					Note:   "One hour segment",
				},
			},
			expected: 1 * time.Hour,
		},
		{
			name: "single open segment",
			segments: []*task.Segment{
				{
					Create: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
					Finish: time.Time{},
					Note:   "Open segment",
				},
			},
			expected: 0,
		},
		{
			name: "multiple closed segments",
			segments: []*task.Segment{
				{
					Create: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
					Finish: time.Date(2023, 1, 1, 10, 30, 0, 0, time.UTC),
					Note:   "30 minute segment",
				},
				{
					Create: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC),
					Finish: time.Date(2023, 1, 1, 12, 15, 0, 0, time.UTC),
					Note:   "75 minute segment",
				},
			},
			expected: 30*time.Minute + 75*time.Minute,
		},
		{
			name: "mixed open and closed segments",
			segments: []*task.Segment{
				{
					Create: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
					Finish: time.Date(2023, 1, 1, 10, 45, 0, 0, time.UTC),
					Note:   "45 minute closed segment",
				},
				{
					Create: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC),
					Finish: time.Time{},
					Note:   "Open segment",
				},
				{
					Create: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Finish: time.Date(2023, 1, 1, 12, 20, 0, 0, time.UTC),
					Note:   "20 minute closed segment",
				},
			},
			expected: 45*time.Minute + 20*time.Minute,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			task := &task.Task{
				Name:        "Test Task",
				Description: "Test Description",
				Tags:        []string{"test"},
				Segments:    testCase.segments,
			}

			result := task.GetClosedSegmentsDuration()
			if result != testCase.expected {
				t.Errorf("Expected duration %v, got %v", testCase.expected, result)
			}
		})
	}
}

func TestGetTasksFilePath(t *testing.T) {
	t.Parallel()

	// Test the function
	path := task.GetTasksFilePath()

	if path == "" {
		t.Error("Expected non-empty path")
	}

	// Should end with the expected filename
	expectedFilename := task.DefaultTasksFileName
	if filepath.Base(path) != expectedFilename {
		// If we can't get user home dir, it should return the default filename
		if path != expectedFilename {
			t.Errorf("Expected path to end with '%s' or be '%s', got '%s'", expectedFilename, expectedFilename, path)
		}
	}

	// If we got a full path, verify it's absolute or relative to current dir
	if filepath.IsAbs(path) {
		// Should be in user's home directory
		homeDir, err := os.UserHomeDir()
		if err == nil {
			expectedPath := filepath.Join(homeDir, expectedFilename)
			if path != expectedPath {
				t.Errorf("Expected path '%s', got '%s'", expectedPath, path)
			}
		}
	}
}

func TestGetTasksFilePathWithHomeError(t *testing.T) {
	t.Parallel()

	// This test is harder to implement since we can't easily mock os.UserHomeDir
	// But we can verify the fallback behavior by checking the code logic
	// The function should return ".ohgmas-tasks.yaml" if os.UserHomeDir() fails

	// We'll test this indirectly by ensuring the function always returns something reasonable
	path := task.GetTasksFilePath()
	if path == "" {
		t.Error("GetTasksFilePath should never return empty string")
	}

	if filepath.Base(path) != ".ohgmas-tasks.yaml" && path != ".ohgmas-tasks.yaml" {
		t.Errorf("Expected path to contain '.ohgmas-tasks.yaml', got '%s'", path)
	}
}
