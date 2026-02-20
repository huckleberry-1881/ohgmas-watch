package task //nolint:testpackage // tests unexported functions

import (
	"testing"
	"time"
)

func TestGetTagsetKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tags []string
		want string
	}{
		{
			name: "empty tags",
			tags: []string{},
			want: "(no tags)",
		},
		{
			name: "nil tags",
			tags: nil,
			want: "(no tags)",
		},
		{
			name: "single tag",
			tags: []string{"frontend"},
			want: "frontend",
		},
		{
			name: "multiple tags sorted",
			tags: []string{"frontend", "backend", "api"},
			want: "api, backend, frontend",
		},
		{
			name: "already sorted",
			tags: []string{"a", "b", "c"},
			want: "a, b, c",
		},
		{
			name: "duplicate tags",
			tags: []string{"work", "work", "personal"},
			want: "personal, work, work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getTagsetKey(tt.tags)
			if got != tt.want {
				t.Errorf("getTagsetKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSortTagsetSummaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		tagsetMap map[string]*TagsetSummary
		wantOrder []string
	}{
		{
			name:      "empty map",
			tagsetMap: map[string]*TagsetSummary{},
			wantOrder: []string{},
		},
		{
			name: "sorted by duration descending",
			tagsetMap: map[string]*TagsetSummary{
				"short":  {Tagset: "short", Duration: time.Hour},
				"long":   {Tagset: "long", Duration: 3 * time.Hour},
				"medium": {Tagset: "medium", Duration: 2 * time.Hour},
			},
			wantOrder: []string{"long", "medium", "short"},
		},
		{
			name: "equal durations",
			tagsetMap: map[string]*TagsetSummary{
				"a": {Tagset: "a", Duration: time.Hour},
				"b": {Tagset: "b", Duration: time.Hour},
			},
			wantOrder: nil, // Order undefined for equal durations
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := sortTagsetSummaries(tt.tagsetMap)

			if tt.wantOrder == nil {
				// Just check length for undefined order
				if len(got) != len(tt.tagsetMap) {
					t.Errorf("sortTagsetSummaries() returned %d items, want %d", len(got), len(tt.tagsetMap))
				}

				return
			}

			if len(got) != len(tt.wantOrder) {
				t.Fatalf("sortTagsetSummaries() returned %d items, want %d", len(got), len(tt.wantOrder))
			}

			for i, wantTagset := range tt.wantOrder {
				if got[i].Tagset != wantTagset {
					t.Errorf("sortTagsetSummaries()[%d].Tagset = %q, want %q", i, got[i].Tagset, wantTagset)
				}
			}
		})
	}
}

func TestWatch_GetSummaryByTagset(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		tasks     []*Task
		start     *time.Time
		finish    *time.Time
		wantLen   int
		wantFirst string
	}{
		{
			name:    "no tasks",
			tasks:   []*Task{},
			start:   nil,
			finish:  nil,
			wantLen: 0,
		},
		{
			name: "single task with segments",
			tasks: []*Task{
				{
					Name: "Task 1",
					Tags: []string{"frontend"},
					Segments: []*Segment{
						{Create: baseTime, Finish: baseTime.Add(time.Hour)},
					},
				},
			},
			start:     nil,
			finish:    nil,
			wantLen:   1,
			wantFirst: "frontend",
		},
		{
			name: "tasks grouped by tagset",
			tasks: []*Task{
				{
					Name: "Task 1",
					Tags: []string{"frontend", "api"},
					Segments: []*Segment{
						{Create: baseTime, Finish: baseTime.Add(time.Hour)},
					},
				},
				{
					Name: "Task 2",
					Tags: []string{"api", "frontend"}, // Same tags, different order
					Segments: []*Segment{
						{Create: baseTime.Add(time.Hour), Finish: baseTime.Add(2 * time.Hour)},
					},
				},
			},
			start:     nil,
			finish:    nil,
			wantLen:   1, // Should group into one tagset
			wantFirst: "api, frontend",
		},
		{
			name: "filtered by time range",
			tasks: []*Task{
				{
					Name: "In Range",
					Tags: []string{"work"},
					Segments: []*Segment{
						{Create: baseTime, Finish: baseTime.Add(time.Hour)},
					},
				},
				{
					Name: "Out of Range",
					Tags: []string{"personal"},
					Segments: []*Segment{
						{Create: baseTime.Add(10 * time.Hour), Finish: baseTime.Add(11 * time.Hour)},
					},
				},
			},
			start:     timePtr(baseTime.Add(-time.Hour)),
			finish:    timePtr(baseTime.Add(2 * time.Hour)),
			wantLen:   1,
			wantFirst: "work",
		},
		{
			name: "task without segments excluded",
			tasks: []*Task{
				{
					Name:     "No Segments",
					Tags:     []string{"empty"},
					Segments: []*Segment{},
				},
				{
					Name: "Has Segments",
					Tags: []string{"full"},
					Segments: []*Segment{
						{Create: baseTime, Finish: baseTime.Add(time.Hour)},
					},
				},
			},
			start:     timePtr(baseTime.Add(-time.Hour)),
			finish:    timePtr(baseTime.Add(2 * time.Hour)),
			wantLen:   1,
			wantFirst: "full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			watch := &Watch{Tasks: tt.tasks}
			got := watch.GetSummaryByTagset(tt.start, tt.finish)

			if len(got) != tt.wantLen {
				t.Errorf("GetSummaryByTagset() returned %d summaries, want %d", len(got), tt.wantLen)
			}

			if tt.wantLen > 0 && got[0].Tagset != tt.wantFirst {
				t.Errorf("GetSummaryByTagset()[0].Tagset = %q, want %q", got[0].Tagset, tt.wantFirst)
			}
		})
	}
}

func TestWatch_GetSummaryByTagset_Duration(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	watch := &Watch{
		Tasks: []*Task{
			{
				Name: "Task 1",
				Tags: []string{"work"},
				Segments: []*Segment{
					{Create: baseTime, Finish: baseTime.Add(2 * time.Hour)},
				},
			},
			{
				Name: "Task 2",
				Tags: []string{"work"},
				Segments: []*Segment{
					{Create: baseTime.Add(3 * time.Hour), Finish: baseTime.Add(4 * time.Hour)},
				},
			},
		},
	}

	got := watch.GetSummaryByTagset(nil, nil)

	if len(got) != 1 {
		t.Fatalf("Expected 1 tagset, got %d", len(got))
	}

	expectedDuration := 3 * time.Hour
	if got[0].Duration != expectedDuration {
		t.Errorf("GetSummaryByTagset()[0].Duration = %v, want %v", got[0].Duration, expectedDuration)
	}

	if len(got[0].Tasks) != 2 {
		t.Errorf("GetSummaryByTagset()[0].Tasks = %d, want 2", len(got[0].Tasks))
	}
}

func TestWatch_GetTasksSortedByActivity(t *testing.T) {
	t.Parallel()

	now := time.Now()

	watch := &Watch{
		Tasks: []*Task{
			{Name: "Old", Segments: []*Segment{
				{Create: now.Add(-2 * time.Hour), Finish: now.Add(-time.Hour)},
			}},
			{Name: "Recent", Segments: []*Segment{
				{Create: now.Add(-30 * time.Minute), Finish: now},
			}},
			{Name: "No Activity"},
		},
	}

	got := watch.GetTasksSortedByActivity()

	if len(got) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(got))
	}

	if got[0].Name != "Recent" {
		t.Errorf("First task should be 'Recent', got %q", got[0].Name)
	}
	if got[1].Name != "Old" {
		t.Errorf("Second task should be 'Old', got %q", got[1].Name)
	}
	if got[2].Name != "No Activity" {
		t.Errorf("Third task should be 'No Activity', got %q", got[2].Name)
	}
}

func TestWatch_GetTaskIndex(t *testing.T) {
	t.Parallel()

	task1 := &Task{Name: "Task 1"}
	task2 := &Task{Name: "Task 2"}
	task3 := &Task{Name: "Task 3"}
	notInWatch := &Task{Name: "Not In Watch"}

	watch := &Watch{
		Tasks: []*Task{task1, task2, task3},
	}

	tests := []struct {
		name string
		task *Task
		want int
	}{
		{"first task", task1, 0},
		{"second task", task2, 1},
		{"third task", task3, 2},
		{"not in watch", notInWatch, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := watch.GetTaskIndex(tt.task)
			if got != tt.want {
				t.Errorf("GetTaskIndex() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestWatch_GetWeeklySummaryByTagset(t *testing.T) {
	t.Parallel()

	// Week 1: Jan 15 (Monday)
	week1Start := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	// Week 2: Jan 22 (Monday)
	week2Start := time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC)

	watch := &Watch{
		Tasks: []*Task{
			{
				Name: "Week 1 Task",
				Tags: []string{"work"},
				Segments: []*Segment{
					{Create: week1Start.Add(time.Hour), Finish: week1Start.Add(2 * time.Hour)},
				},
			},
			{
				Name: "Week 2 Task",
				Tags: []string{"personal"},
				Segments: []*Segment{
					{Create: week2Start.Add(time.Hour), Finish: week2Start.Add(3 * time.Hour)},
				},
			},
		},
	}

	weekStarts := []time.Time{week1Start, week2Start}
	got := watch.GetWeeklySummaryByTagset(weekStarts)

	if len(got) != 2 {
		t.Fatalf("Expected 2 weekly summaries, got %d", len(got))
	}

	if !got[0].WeekStart.Equal(week1Start) {
		t.Errorf("First week start = %v, want %v", got[0].WeekStart, week1Start)
	}

	if len(got[0].Tagsets) != 1 {
		t.Errorf("First week tagsets = %d, want 1", len(got[0].Tagsets))
	}

	if got[0].Tagsets[0].Tagset != "work" {
		t.Errorf("First week tagset = %q, want 'work'", got[0].Tagsets[0].Tagset)
	}
}

func TestWatch_GetWeeklySummaryByTagset_EmptyWeeks(t *testing.T) {
	t.Parallel()

	week1Start := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	week2Start := time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC) // No data for this week

	watch := &Watch{
		Tasks: []*Task{
			{
				Name: "Week 1 Task",
				Tags: []string{"work"},
				Segments: []*Segment{
					{Create: week1Start.Add(time.Hour), Finish: week1Start.Add(2 * time.Hour)},
				},
			},
		},
	}

	weekStarts := []time.Time{week1Start, week2Start}
	got := watch.GetWeeklySummaryByTagset(weekStarts)

	// Week 2 should be excluded because it has no data
	if len(got) != 1 {
		t.Errorf("Expected 1 weekly summary (empty week excluded), got %d", len(got))
	}
}

func TestWatch_GetEarliestAndLatestSegmentTimes(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		tasks        []*Task
		wantEarliest time.Time
		wantLatest   time.Time
	}{
		{
			name:         "no tasks",
			tasks:        []*Task{},
			wantEarliest: time.Time{},
			wantLatest:   time.Time{},
		},
		{
			name: "no segments",
			tasks: []*Task{
				{Name: "Empty", Segments: []*Segment{}},
			},
			wantEarliest: time.Time{},
			wantLatest:   time.Time{},
		},
		{
			name: "single closed segment",
			tasks: []*Task{
				{
					Name: "Task",
					Segments: []*Segment{
						{Create: baseTime, Finish: baseTime.Add(time.Hour)},
					},
				},
			},
			wantEarliest: baseTime,
			wantLatest:   baseTime.Add(time.Hour),
		},
		{
			name: "open segment uses create time for latest",
			tasks: []*Task{
				{
					Name: "Task",
					Segments: []*Segment{
						{Create: baseTime, Finish: time.Time{}},
					},
				},
			},
			wantEarliest: baseTime,
			wantLatest:   baseTime,
		},
		{
			name: "multiple tasks and segments",
			tasks: []*Task{
				{
					Name: "Task 1",
					Segments: []*Segment{
						{Create: baseTime, Finish: baseTime.Add(time.Hour)},
						{Create: baseTime.Add(2 * time.Hour), Finish: baseTime.Add(3 * time.Hour)},
					},
				},
				{
					Name: "Task 2",
					Segments: []*Segment{
						{Create: baseTime.Add(-time.Hour), Finish: baseTime.Add(-30 * time.Minute)}, // Earliest
						{Create: baseTime.Add(4 * time.Hour), Finish: baseTime.Add(5 * time.Hour)}, // Latest
					},
				},
			},
			wantEarliest: baseTime.Add(-time.Hour),
			wantLatest:   baseTime.Add(5 * time.Hour),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			watch := &Watch{Tasks: tt.tasks}
			gotEarliest, gotLatest := watch.GetEarliestAndLatestSegmentTimes()

			if !gotEarliest.Equal(tt.wantEarliest) {
				t.Errorf("GetEarliestAndLatestSegmentTimes() earliest = %v, want %v", gotEarliest, tt.wantEarliest)
			}
			if !gotLatest.Equal(tt.wantLatest) {
				t.Errorf("GetEarliestAndLatestSegmentTimes() latest = %v, want %v", gotLatest, tt.wantLatest)
			}
		})
	}
}

func TestWatch_GetWeeklySummaryByTagsetWithTasks(t *testing.T) {
	t.Parallel()

	week1Start := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	watch := &Watch{
		Tasks: []*Task{
			{
				Name: "Task 1",
				Tags: []string{"work"},
				Segments: []*Segment{
					{Create: week1Start.Add(time.Hour), Finish: week1Start.Add(3 * time.Hour)},
				},
			},
			{
				Name: "Task 2",
				Tags: []string{"work"},
				Segments: []*Segment{
					{Create: week1Start.Add(4 * time.Hour), Finish: week1Start.Add(5 * time.Hour)},
				},
			},
		},
	}

	weekStarts := []time.Time{week1Start}
	got := watch.GetWeeklySummaryByTagsetWithTasks(weekStarts)

	if len(got) != 1 {
		t.Fatalf("Expected 1 weekly summary, got %d", len(got))
	}

	if len(got[0].Tagsets) != 1 {
		t.Fatalf("Expected 1 tagset, got %d", len(got[0].Tagsets))
	}

	tagset := got[0].Tagsets[0]
	if tagset.Tagset != "work" {
		t.Errorf("Tagset = %q, want 'work'", tagset.Tagset)
	}

	if len(tagset.Tasks) != 2 {
		t.Errorf("Tasks in tagset = %d, want 2", len(tagset.Tasks))
	}

	expectedDuration := 3 * time.Hour // 2h + 1h
	if tagset.Duration != expectedDuration {
		t.Errorf("Duration = %v, want %v", tagset.Duration, expectedDuration)
	}
}

func TestWatch_GetWeeklySummaryByTagsetWithTasks_TaskFiltering(t *testing.T) {
	t.Parallel()

	week1Start := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	week2Start := time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC)

	watch := &Watch{
		Tasks: []*Task{
			{
				Name: "Week 1 Only",
				Tags: []string{"work"},
				Segments: []*Segment{
					{Create: week1Start.Add(time.Hour), Finish: week1Start.Add(2 * time.Hour)},
				},
			},
			{
				Name: "Week 2 Only",
				Tags: []string{"work"},
				Segments: []*Segment{
					{Create: week2Start.Add(time.Hour), Finish: week2Start.Add(2 * time.Hour)},
				},
			},
		},
	}

	weekStarts := []time.Time{week1Start}
	got := watch.GetWeeklySummaryByTagsetWithTasks(weekStarts)

	if len(got) != 1 {
		t.Fatalf("Expected 1 weekly summary, got %d", len(got))
	}

	// Only week 1 task should be included
	if len(got[0].Tagsets[0].Tasks) != 1 {
		t.Errorf("Expected 1 task for week 1, got %d", len(got[0].Tagsets[0].Tasks))
	}

	if got[0].Tagsets[0].Tasks[0].Name != "Week 1 Only" {
		t.Errorf("Task name = %q, want 'Week 1 Only'", got[0].Tagsets[0].Tasks[0].Name)
	}
}
