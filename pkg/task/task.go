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
		category = "work" //nolint:goconst // simple default, not worth a constant
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

// DefaultTasksDir is the directory under the user's home directory where quarter files are stored.
const DefaultTasksDir = "ohgmas"

// GetCurrentFiscalQuarter returns the fiscal year and quarter for the current time.
// The fiscal year starts in October: Q1 = Oct-Dec, Q2 = Jan-Mar, Q3 = Apr-Jun, Q4 = Jul-Sep.
// The fiscal year for Oct-Dec is the following calendar year (e.g., Oct 2025 is FY2026 Q1).
func GetCurrentFiscalQuarter() (year, quarter int) { //nolint:nonamedreturns // names aid readability
	return GetFiscalQuarterForTime(time.Now())
}

// GetFiscalQuarterForTime returns the fiscal year and quarter for the given time.
func GetFiscalQuarterForTime(t time.Time) (year, quarter int) { //nolint:nonamedreturns // names aid readability
	month := int(t.Month())
	calYear := t.Year()

	switch {
	case month >= 10:
		return calYear + 1, 1
	case month <= 3:
		return calYear, 2
	case month <= 6:
		return calYear, 3
	default:
		return calYear, 4
	}
}

// GetTasksDir returns the directory where quarter files are stored.
func GetTasksDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultTasksDir
	}

	return filepath.Join(homeDir, DefaultTasksDir)
}

// GetQuarterFileName returns the filename for the given fiscal year and quarter (e.g., "2026-F3.yaml").
func GetQuarterFileName(year, quarter int) string {
	return fmt.Sprintf("%d-F%d.yaml", year, quarter)
}

// GetTasksFilePath returns the path to the current fiscal quarter's tasks file.
func GetTasksFilePath() string {
	year, quarter := GetCurrentFiscalQuarter()

	return filepath.Join(GetTasksDir(), GetQuarterFileName(year, quarter))
}

// GetAllQuarterFilePaths returns the paths of all *.yaml files in the tasks directory.
// Returns an empty slice if the directory does not exist.
func GetAllQuarterFilePaths() ([]string, error) {
	dir := GetTasksDir()

	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return []string{}, nil
	}

	matches, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob quarter files: %w", err)
	}

	return matches, nil
}

// SaveTasksToFile saves tasks to YAML file at specified path. Creates parent directory if needed.
func (w *Watch) SaveTasksToFile(filePath string) error {
	err := os.MkdirAll(filepath.Dir(filePath), 0700)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

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

// LoadTasksFromFiles loads and merges tasks from multiple YAML files.
func (w *Watch) LoadTasksFromFiles(filePaths []string) error {
	w.Tasks = []*Task{}

	for _, filePath := range filePaths {
		var loaded []*Task

		data, err := os.ReadFile(filePath) //nolint:gosec // File path is provided by the caller for intended file loading
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return fmt.Errorf("unable to read file %s: %w", filePath, err)
		}

		err = yaml.Unmarshal(data, &loaded)
		if err != nil {
			return fmt.Errorf("unable to yaml unmarshal %s: %w", filePath, err)
		}

		w.Tasks = append(w.Tasks, loaded...)
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
