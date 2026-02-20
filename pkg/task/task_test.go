package task

import (
	"sync"
	"testing"
	"time"
)

func TestWatch_AddTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		taskName    string
		description string
		tags        []string
		category    string
		wantLen     int
	}{
		{
			name:        "adds task with all fields",
			taskName:    "Test Task",
			description: "A test description",
			tags:        []string{"tag1", "tag2"},
			category:    "work",
			wantLen:     1,
		},
		{
			name:        "adds task with empty category defaults to work",
			taskName:    "Default Category Task",
			description: "Description",
			tags:        []string{},
			category:    "",
			wantLen:     1,
		},
		{
			name:        "adds task with nil tags",
			taskName:    "Nil Tags Task",
			description: "",
			tags:        nil,
			category:    "backlog",
			wantLen:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			watch := &Watch{Tasks: []*Task{}}
			watch.AddTask(tt.taskName, tt.description, tt.tags, tt.category)

			if len(watch.Tasks) != tt.wantLen {
				t.Errorf("AddTask() resulted in %d tasks, want %d", len(watch.Tasks), tt.wantLen)
			}

			task := watch.Tasks[0]
			if task.Name != tt.taskName {
				t.Errorf("AddTask() name = %q, want %q", task.Name, tt.taskName)
			}
			if task.Description != tt.description {
				t.Errorf("AddTask() description = %q, want %q", task.Description, tt.description)
			}

			expectedCategory := tt.category
			if expectedCategory == "" {
				expectedCategory = "work"
			}
			if task.Category != expectedCategory {
				t.Errorf("AddTask() category = %q, want %q", task.Category, expectedCategory)
			}
		})
	}
}

func TestWatch_AddTask_Multiple(t *testing.T) {
	t.Parallel()

	watch := &Watch{Tasks: []*Task{}}
	watch.AddTask("Task 1", "Desc 1", []string{"tag1"}, "work")
	watch.AddTask("Task 2", "Desc 2", []string{"tag2"}, "completed")
	watch.AddTask("Task 3", "Desc 3", []string{"tag3"}, "backlog")

	if len(watch.Tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(watch.Tasks))
	}
}

func TestWatch_AddTask_Concurrent(t *testing.T) {
	t.Parallel()

	watch := &Watch{Tasks: []*Task{}}
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			watch.AddTask("Task", "Description", []string{"tag"}, "work")
		}(i)
	}

	wg.Wait()

	if len(watch.Tasks) != 100 {
		t.Errorf("Expected 100 tasks after concurrent adds, got %d", len(watch.Tasks))
	}
}

func TestTask_AddSegment(t *testing.T) {
	t.Parallel()

	task := &Task{
		Name:     "Test Task",
		Segments: []*Segment{},
	}

	task.AddSegment("Test note")

	if len(task.Segments) != 1 {
		t.Errorf("AddSegment() resulted in %d segments, want 1", len(task.Segments))
	}

	segment := task.Segments[0]
	if segment.Note != "Test note" {
		t.Errorf("AddSegment() note = %q, want %q", segment.Note, "Test note")
	}
	if segment.Create.IsZero() {
		t.Error("AddSegment() should set Create time")
	}
	if !segment.Finish.IsZero() {
		t.Error("AddSegment() should leave Finish as zero")
	}
}

func TestTask_AddSegment_EmptyNote(t *testing.T) {
	t.Parallel()

	task := &Task{
		Name:     "Test Task",
		Segments: []*Segment{},
	}

	task.AddSegment("")

	if len(task.Segments) != 1 {
		t.Errorf("AddSegment() resulted in %d segments, want 1", len(task.Segments))
	}

	if task.Segments[0].Note != "" {
		t.Errorf("AddSegment() note = %q, want empty", task.Segments[0].Note)
	}
}

func TestTask_AddSegment_Concurrent(t *testing.T) {
	t.Parallel()

	task := &Task{
		Name:     "Test Task",
		Segments: []*Segment{},
	}

	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			task.AddSegment("Note")
		}(i)
	}

	wg.Wait()

	if len(task.Segments) != 50 {
		t.Errorf("Expected 50 segments after concurrent adds, got %d", len(task.Segments))
	}
}

func TestTask_CloseSegment(t *testing.T) {
	t.Parallel()

	task := &Task{
		Name: "Test Task",
		Segments: []*Segment{
			{Create: time.Now().Add(-time.Hour), Finish: time.Time{}, Note: "Open segment"},
		},
	}

	task.CloseSegment()

	if task.Segments[0].Finish.IsZero() {
		t.Error("CloseSegment() should set Finish time")
	}
}

func TestTask_CloseSegment_NoOpenSegments(t *testing.T) {
	t.Parallel()

	closedTime := time.Now().Add(-30 * time.Minute)
	task := &Task{
		Name: "Test Task",
		Segments: []*Segment{
			{Create: time.Now().Add(-time.Hour), Finish: closedTime, Note: "Already closed"},
		},
	}

	task.CloseSegment()

	// Should not change the finish time of already closed segments
	if !task.Segments[0].Finish.Equal(closedTime) {
		t.Error("CloseSegment() should not modify already closed segments")
	}
}

func TestTask_CloseSegment_MultipleOpenSegments(t *testing.T) {
	t.Parallel()

	task := &Task{
		Name: "Test Task",
		Segments: []*Segment{
			{Create: time.Now().Add(-2 * time.Hour), Finish: time.Time{}, Note: "Open 1"},
			{Create: time.Now().Add(-time.Hour), Finish: time.Time{}, Note: "Open 2"},
		},
	}

	task.CloseSegment()

	for i, seg := range task.Segments {
		if seg.Finish.IsZero() {
			t.Errorf("CloseSegment() segment %d should be closed", i)
		}
	}
}

func TestTask_CloseSegment_EmptySegments(t *testing.T) {
	t.Parallel()

	task := &Task{
		Name:     "Test Task",
		Segments: []*Segment{},
	}

	// Should not panic
	task.CloseSegment()

	if len(task.Segments) != 0 {
		t.Error("CloseSegment() should not add segments")
	}
}

func TestTask_HasUnclosedSegment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		segments []*Segment
		want     bool
	}{
		{
			name:     "no segments",
			segments: []*Segment{},
			want:     false,
		},
		{
			name: "all segments closed",
			segments: []*Segment{
				{Create: time.Now().Add(-time.Hour), Finish: time.Now().Add(-30 * time.Minute)},
			},
			want: false,
		},
		{
			name: "one open segment",
			segments: []*Segment{
				{Create: time.Now().Add(-time.Hour), Finish: time.Time{}},
			},
			want: true,
		},
		{
			name: "mixed segments with open",
			segments: []*Segment{
				{Create: time.Now().Add(-2 * time.Hour), Finish: time.Now().Add(-time.Hour)},
				{Create: time.Now().Add(-30 * time.Minute), Finish: time.Time{}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task := &Task{
				Name:     "Test Task",
				Segments: tt.segments,
			}

			if got := task.HasUnclosedSegment(); got != tt.want {
				t.Errorf("HasUnclosedSegment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_GetClosedSegmentsDuration(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name     string
		segments []*Segment
		want     time.Duration
	}{
		{
			name:     "no segments",
			segments: []*Segment{},
			want:     0,
		},
		{
			name: "single closed segment",
			segments: []*Segment{
				{Create: now.Add(-time.Hour), Finish: now},
			},
			want: time.Hour,
		},
		{
			name: "multiple closed segments",
			segments: []*Segment{
				{Create: now.Add(-2 * time.Hour), Finish: now.Add(-time.Hour)},
				{Create: now.Add(-30 * time.Minute), Finish: now},
			},
			want: time.Hour + 30*time.Minute,
		},
		{
			name: "ignores open segments",
			segments: []*Segment{
				{Create: now.Add(-2 * time.Hour), Finish: now.Add(-time.Hour)},
				{Create: now.Add(-30 * time.Minute), Finish: time.Time{}}, // Open
			},
			want: time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task := &Task{
				Name:     "Test Task",
				Segments: tt.segments,
			}

			got := task.GetClosedSegmentsDuration()
			if got != tt.want {
				t.Errorf("GetClosedSegmentsDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_SetCategory(t *testing.T) {
	t.Parallel()

	task := &Task{Name: "Test", Category: "work"}

	task.SetCategory("completed")
	if task.Category != "completed" {
		t.Errorf("SetCategory() = %q, want %q", task.Category, "completed")
	}

	task.SetCategory("backlog")
	if task.Category != "backlog" {
		t.Errorf("SetCategory() = %q, want %q", task.Category, "backlog")
	}
}

func TestTask_GetCategory(t *testing.T) {
	t.Parallel()

	task := &Task{Name: "Test", Category: "work"}

	if got := task.GetCategory(); got != "work" {
		t.Errorf("GetCategory() = %q, want %q", got, "work")
	}

	task.Category = "completed"
	if got := task.GetCategory(); got != "completed" {
		t.Errorf("GetCategory() = %q, want %q", got, "completed")
	}
}

func TestWatch_GetTasksByCategory(t *testing.T) {
	t.Parallel()

	now := time.Now()
	watch := &Watch{
		Tasks: []*Task{
			{Name: "Work 1", Category: "work", Segments: []*Segment{
				{Create: now.Add(-time.Hour), Finish: now},
			}},
			{Name: "Completed 1", Category: "completed"},
			{Name: "Work 2", Category: "work"},
			{Name: "Backlog 1", Category: "backlog"},
		},
	}

	workTasks := watch.GetTasksByCategory("work")
	if len(workTasks) != 2 {
		t.Errorf("GetTasksByCategory('work') returned %d tasks, want 2", len(workTasks))
	}

	completedTasks := watch.GetTasksByCategory("completed")
	if len(completedTasks) != 1 {
		t.Errorf("GetTasksByCategory('completed') returned %d tasks, want 1", len(completedTasks))
	}

	nonExistent := watch.GetTasksByCategory("nonexistent")
	if len(nonExistent) != 0 {
		t.Errorf("GetTasksByCategory('nonexistent') returned %d tasks, want 0", len(nonExistent))
	}
}

func TestWatch_GetTasksSortedByActivityWithFilter(t *testing.T) {
	t.Parallel()

	now := time.Now()
	watch := &Watch{
		Tasks: []*Task{
			{Name: "Work 1", Category: "work", Segments: []*Segment{
				{Create: now.Add(-time.Hour), Finish: now},
			}},
			{Name: "Completed 1", Category: "completed"},
			{Name: "Work 2", Category: "work", Segments: []*Segment{
				{Create: now.Add(-2 * time.Hour), Finish: now.Add(-time.Hour)},
			}},
		},
	}

	// With filter
	workTasks := watch.GetTasksSortedByActivityWithFilter("work")
	if len(workTasks) != 2 {
		t.Errorf("GetTasksSortedByActivityWithFilter('work') returned %d tasks, want 2", len(workTasks))
	}

	// Without filter (empty string)
	allTasks := watch.GetTasksSortedByActivityWithFilter("")
	if len(allTasks) != 3 {
		t.Errorf("GetTasksSortedByActivityWithFilter('') returned %d tasks, want 3", len(allTasks))
	}
}

func TestSortTasksByActivity(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		tasks     []*Task
		wantOrder []string
	}{
		{
			name:      "empty tasks",
			tasks:     []*Task{},
			wantOrder: []string{},
		},
		{
			name: "tasks sorted by activity",
			tasks: []*Task{
				{Name: "Old", Segments: []*Segment{
					{Create: now.Add(-2 * time.Hour), Finish: now.Add(-time.Hour)},
				}},
				{Name: "Recent", Segments: []*Segment{
					{Create: now.Add(-30 * time.Minute), Finish: now},
				}},
			},
			wantOrder: []string{"Recent", "Old"},
		},
		{
			name: "tasks with no segments go to bottom",
			tasks: []*Task{
				{Name: "No Segments"},
				{Name: "Has Segments", Segments: []*Segment{
					{Create: now.Add(-time.Hour), Finish: now},
				}},
			},
			wantOrder: []string{"Has Segments", "No Segments"},
		},
		{
			name: "multiple tasks without segments maintain order",
			tasks: []*Task{
				{Name: "No Seg 1"},
				{Name: "No Seg 2"},
			},
			wantOrder: []string{"No Seg 1", "No Seg 2"},
		},
		{
			name: "open segment uses create time",
			tasks: []*Task{
				{Name: "Closed Old", Segments: []*Segment{
					{Create: now.Add(-2 * time.Hour), Finish: now.Add(-time.Hour)},
				}},
				{Name: "Open Recent", Segments: []*Segment{
					{Create: now.Add(-10 * time.Minute), Finish: time.Time{}},
				}},
			},
			wantOrder: []string{"Open Recent", "Closed Old"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sorted := sortTasksByActivity(tt.tasks)

			if len(sorted) != len(tt.wantOrder) {
				t.Fatalf("sortTasksByActivity() returned %d tasks, want %d", len(sorted), len(tt.wantOrder))
			}

			for i, name := range tt.wantOrder {
				if sorted[i].Name != name {
					t.Errorf("sortTasksByActivity()[%d].Name = %q, want %q", i, sorted[i].Name, name)
				}
			}
		})
	}
}

func TestGetTasksFilePath(t *testing.T) {
	t.Parallel()

	path := GetTasksFilePath()

	// Should return a non-empty path
	if path == "" {
		t.Error("GetTasksFilePath() returned empty string")
	}

	// Should contain the default filename
	if path != DefaultTasksFileName && len(path) < len(DefaultTasksFileName) {
		t.Errorf("GetTasksFilePath() = %q, should contain %q", path, DefaultTasksFileName)
	}
}
