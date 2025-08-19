package task

import "time"

// Manager defines the interface for task management operations.
type Manager interface {
	AddTask(name, description string, tags []string)
	GetTasksSortedByActivity() []*Task
	GetSummaryByTagset(start, finish *time.Time) []TagsetSummary
	SaveTasks() error
	LoadTasks() error
}

// TimeTracker defines the interface for time tracking operations.
type TimeTracker interface {
	AddSegment(note string)
	CloseSegment()
	HasUnclosedSegment() bool
	GetClosedSegmentsDuration() time.Duration
	GetLastActivity() time.Time
	IsActive() bool
}

// Persister defines the interface for persistence operations.
type Persister interface {
	SaveTasksToFile(filePath string) error
	LoadTasksFromFile(filePath string) error
}

// Ensure our types implement the interfaces (compile-time check).
var (
	_ Manager     = (*Watch)(nil)
	_ TimeTracker = (*Task)(nil)
	_ Persister   = (*Watch)(nil)
)