// Package task provides functionality for managing time tracking tasks and segments.
package task

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/goccy/go-yaml"
)

// AddTask adds a new task to a watch (thread-safe).
func (w *Watch) AddTask(name string, description string, tags []string, category string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Default to "work" if no category specified
	if category == "" {
		category = "work"
	}

	newTask := Task{
		Name:        name,
		Description: description,
		Tags:        tags,
		Category:    category,
		Segments:    []*Segment{},
		mu:          sync.RWMutex{},
	}

	w.Tasks = append(w.Tasks, &newTask)
}

// AddSegment adds a new segment to a task (thread-safe).
func (t *Task) AddSegment(note string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	newSeg := Segment{
		Note:   note,
		Create: time.Now(),
		Finish: time.Time{},
	}

	t.Segments = append(t.Segments, &newSeg)
}

// CloseSegment closes an open segment (thread-safe).
func (t *Task) CloseSegment() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, segment := range t.Segments {
		if segment.Finish.IsZero() {
			segment.Finish = time.Now()
		}
	}
}

// HasUnclosedSegment checks if a task has unclosed segments (thread-safe).
func (t *Task) HasUnclosedSegment() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, segment := range t.Segments {
		if segment.Finish.IsZero() {
			return true
		}
	}

	return false
}

// GetClosedSegmentsDuration calculates total duration of closed segments (thread-safe).
func (t *Task) GetClosedSegmentsDuration() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var totalDuration time.Duration

	for _, segment := range t.Segments {
		if !segment.Finish.IsZero() {
			totalDuration += segment.Finish.Sub(segment.Create)
		}
	}

	return totalDuration
}

// DefaultTasksFileName is the default filename for storing tasks.
const DefaultTasksFileName = ".ohgmas-tasks.yaml"

// GetTasksFilePath gets the path to the tasks file in user's home directory.
func GetTasksFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultTasksFileName
	}

	return filepath.Join(homeDir, DefaultTasksFileName)
}

// SaveTasksToFile saves tasks to YAML file at specified path.
func (w *Watch) SaveTasksToFile(filePath string) error {
	data, err := yaml.Marshal(w.Tasks)
	if err != nil {
		return fmt.Errorf("unable to yaml marshal: %w", err)
	}

	err = os.WriteFile(filePath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// LoadTasksFromFile loads tasks from YAML file at specified path.
func (w *Watch) LoadTasksFromFile(filePath string) error {
	data, err := os.ReadFile(filePath) //nolint:gosec // File path is provided by the caller for intended file loading
	if err != nil {
		if os.IsNotExist(err) {
			w.Tasks = []*Task{} // Set empty slice if file doesn't exist

			return nil
		}

		return fmt.Errorf("unable to read file: %w", err)
	}

	err = yaml.Unmarshal(data, &w.Tasks)
	if err != nil {
		return fmt.Errorf("unable to yaml unmarshal: %w", err)
	}

	return nil
}

// SaveTasks saves tasks to YAML file (uses default path).
func (w *Watch) SaveTasks() error {
	return w.SaveTasksToFile(GetTasksFilePath())
}

// LoadTasks loads tasks from YAML file (uses default path).
func (w *Watch) LoadTasks() error {
	return w.LoadTasksFromFile(GetTasksFilePath())
}

// SetCategory sets the category of a task (thread-safe).
func (t *Task) SetCategory(category string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Category = category
}

// GetCategory gets the category of a task (thread-safe).
func (t *Task) GetCategory() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.Category
}

// GetTasksByCategory returns tasks filtered by category, sorted by activity (thread-safe).
func (w *Watch) GetTasksByCategory(category string) []*Task {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var filteredTasks []*Task

	for _, task := range w.Tasks {
		if task.GetCategory() == category {
			filteredTasks = append(filteredTasks, task)
		}
	}

	// Sort by activity using the same logic as GetTasksSortedByActivity
	return sortTasksByActivity(filteredTasks)
}

// GetTasksSortedByActivityWithFilter returns tasks filtered by category if specified, otherwise all tasks.
func (w *Watch) GetTasksSortedByActivityWithFilter(categoryFilter string) []*Task {
	if categoryFilter == "" {
		return w.GetTasksSortedByActivity()
	}

	return w.GetTasksByCategory(categoryFilter)
}

// sortTasksByActivity sorts a slice of tasks by last activity (most recent first).
func sortTasksByActivity(tasks []*Task) []*Task {
	if len(tasks) == 0 {
		return tasks
	}

	// Create index mapping for sorting
	taskIndices := make([]int, len(tasks))
	for i := range taskIndices {
		taskIndices[i] = i
	}

	sort.Slice(taskIndices, func(i, j int) bool {
		taskA, taskB := tasks[taskIndices[i]], tasks[taskIndices[j]]
		lastActivityA := taskA.GetLastActivity()
		lastActivityB := taskB.GetLastActivity()

		// Tasks with no segments go to the bottom
		if lastActivityA.IsZero() && lastActivityB.IsZero() {
			return false // Keep original order for tasks with no segments
		}

		if lastActivityA.IsZero() {
			return false // Task A goes after task B
		}

		if lastActivityB.IsZero() {
			return true // Task A goes before task B
		}

		// Sort by most recent activity first
		return lastActivityA.After(lastActivityB)
	})

	// Return sorted tasks
	sorted := make([]*Task, len(tasks))
	for i, idx := range taskIndices {
		sorted[i] = tasks[idx]
	}

	return sorted
}
