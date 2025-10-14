package task_test

import (
	"testing"
	"time"

	"github.com/huckleberry-1881/ohgmas-watch/pkg/task"
)

func TestGetSummaryByTagset(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		tasks         []*task.Task
		start         *time.Time
		finish        *time.Time
		expectedCount int
		expectedFirst string // Expected first tagset name
	}{
		{
			name:          "no tasks",
			tasks:         []*task.Task{},
			start:         nil,
			finish:        nil,
			expectedCount: 0,
		},
		{
			name: "tasks with different tags",
			tasks: []*task.Task{
				{
					Name: "Task1",
					Tags: []string{"work", "urgent"},
					Segments: []*task.Segment{
						{Create: baseTime, Finish: baseTime.Add(2 * time.Hour)},
					},
				},
				{
					Name: "Task2",
					Tags: []string{"personal"},
					Segments: []*task.Segment{
						{Create: baseTime, Finish: baseTime.Add(time.Hour)},
					},
				},
				{
					Name: "Task3",
					Tags: []string{},
					Segments: []*task.Segment{
						{Create: baseTime, Finish: baseTime.Add(30 * time.Minute)},
					},
				},
			},
			start:         nil,
			finish:        nil,
			expectedCount: 3,
			expectedFirst: "urgent, work", // Longest duration (2 hours)
		},
		{
			name: "filter by time range",
			tasks: []*task.Task{
				{
					Name: "Task1",
					Tags: []string{"work"},
					Segments: []*task.Segment{
						{Create: baseTime, Finish: baseTime.Add(time.Hour)},
					},
				},
				{
					Name: "Task2",
					Tags: []string{"work"},
					Segments: []*task.Segment{
						{Create: baseTime.Add(2 * time.Hour), Finish: baseTime.Add(3 * time.Hour)},
					},
				},
			},
			start: func() *time.Time {
				t := baseTime.Add(90 * time.Minute)

				return &t
			}(),
			finish: nil,
			expectedCount: 1,
			expectedFirst: "work",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			watch := &task.Watch{
				Tasks: tc.tasks,
			}

			summaries := watch.GetSummaryByTagset(tc.start, tc.finish)

			if len(summaries) != tc.expectedCount {
				t.Errorf("Expected %d summaries, got %d", tc.expectedCount, len(summaries))
			}

			if tc.expectedCount > 0 && len(summaries) > 0 {
				if summaries[0].Tagset != tc.expectedFirst {
					t.Errorf("Expected first tagset to be '%s', got '%s'", tc.expectedFirst, summaries[0].Tagset)
				}
			}
		})
	}
}

func TestGetTasksSortedByActivity(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		tasks         []*task.Task
		expectedOrder []string // Expected task names in order
	}{
		{
			name:          "no tasks",
			tasks:         []*task.Task{},
			expectedOrder: []string{},
		},
		{
			name: "tasks with different activity times",
			tasks: []*task.Task{
				{
					Name: "Old Task",
					Segments: []*task.Segment{
						{Create: baseTime, Finish: baseTime.Add(time.Hour)},
					},
				},
				{
					Name: "Recent Task",
					Segments: []*task.Segment{
						{Create: baseTime.Add(2 * time.Hour), Finish: baseTime.Add(3 * time.Hour)},
					},
				},
				{
					Name: "Active Task",
					Segments: []*task.Segment{
						{Create: baseTime.Add(4 * time.Hour), Finish: time.Time{}},
					},
				},
				{
					Name:     "No Segments",
					Segments: []*task.Segment{},
				},
			},
			expectedOrder: []string{"Active Task", "Recent Task", "Old Task", "No Segments"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			watch := &task.Watch{
				Tasks: tc.tasks,
			}

			sorted := watch.GetTasksSortedByActivity()

			if len(sorted) != len(tc.expectedOrder) {
				t.Errorf("Expected %d tasks, got %d", len(tc.expectedOrder), len(sorted))
			}

			for i, expectedName := range tc.expectedOrder {
				if i < len(sorted) && sorted[i].Name != expectedName {
					t.Errorf("At position %d: expected '%s', got '%s'", i, expectedName, sorted[i].Name)
				}
			}
		})
	}
}

func TestGetTaskIndex(t *testing.T) {
	t.Parallel()

	task1 := &task.Task{Name: "Task1"}
	task2 := &task.Task{Name: "Task2"}
	task3 := &task.Task{Name: "Task3"}
	taskNotInList := &task.Task{Name: "Not In List"}

	watch := &task.Watch{
		Tasks: []*task.Task{task1, task2, task3},
	}

	tests := []struct {
		name     string
		task     *task.Task
		expected int
	}{
		{
			name:     "first task",
			task:     task1,
			expected: 0,
		},
		{
			name:     "middle task",
			task:     task2,
			expected: 1,
		},
		{
			name:     "last task",
			task:     task3,
			expected: 2,
		},
		{
			name:     "task not in list",
			task:     taskNotInList,
			expected: -1,
		},
		{
			name:     "nil task",
			task:     nil,
			expected: -1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := watch.GetTaskIndex(tc.task)
			if result != tc.expected {
				t.Errorf("Expected index %d, got %d", tc.expected, result)
			}
		})
	}
}
