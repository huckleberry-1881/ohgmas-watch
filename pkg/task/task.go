// Package task provides functionality for managing time tracking tasks and segments.
package task

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/goccy/go-yaml"
)

// AddTask adds a new task to a watch.
func (w *Watch) AddTask(name string, description string, tags []string) {
	newTask := Task{
		Name:        name,
		Description: description,
		Tags:        tags,
		Segments:    []*Segment{},
	}

	w.Tasks = append(w.Tasks, &newTask)
}

// AddSegment adds a new segment to a task.
func (t *Task) AddSegment(note string) {
	newSeg := Segment{
		Note:   note,
		Create: time.Now(),
		Finish: time.Time{},
	}

	t.Segments = append(t.Segments, &newSeg)
}

// CloseSegment closes an open segment.
func (t *Task) CloseSegment() {
	for _, segment := range t.Segments {
		if segment.Finish.IsZero() {
			segment.Finish = time.Now()
		}
	}
}

// HasUnclosedSegment checks if a task has unclosed segments.
func (t *Task) HasUnclosedSegment() bool {
	for _, segment := range t.Segments {
		if segment.Finish.IsZero() {
			return true
		}
	}

	return false
}

// GetClosedSegmentsDuration calculates total duration of closed segments.
func (t *Task) GetClosedSegmentsDuration() time.Duration {
	var totalDuration time.Duration

	for _, segment := range t.Segments {
		if !segment.Finish.IsZero() {
			totalDuration += segment.Finish.Sub(segment.Create)
		}
	}

	return totalDuration
}

const defaultTasksFileName = ".ohgmas-tasks.yaml"

// GetTasksFilePath gets the path to the tasks file in user's home directory.
func GetTasksFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return defaultTasksFileName
	}

	return filepath.Join(homeDir, defaultTasksFileName)
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
	data, err := os.ReadFile(filePath)
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
