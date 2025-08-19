package task

import (
	"sync"
	"time"
)

// Watch represents a collection of tasks being tracked.
type Watch struct {
	Tasks []*Task      `yaml:"tasks"`
	mu    sync.RWMutex `yaml:"-"` // mutex for thread-safe operations, not serialized
}

// Task represents a work task with time tracking segments.
type Task struct {
	Name        string       `yaml:"name"`
	Description string       `yaml:"description"`
	Tags        []string     `yaml:"tags"`
	Segments    []*Segment   `yaml:"segments"`
	mu          sync.RWMutex `yaml:"-"` // mutex for thread-safe segment operations
}

// Segment represents a time tracking period for a task.
type Segment struct {
	Create time.Time `yaml:"create"`
	Finish time.Time `yaml:"finish"`
	Note   string    `yaml:"note"`
}
