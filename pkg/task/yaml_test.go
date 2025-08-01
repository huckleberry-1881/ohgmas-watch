package task

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)


// Test helper to create a temporary test file.
func createTempTasksFile(t *testing.T, content string) string {
	t.Helper()

	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test-tasks.yaml")

	if content != "" {
		err := os.WriteFile(tempFile, []byte(content), 0600)
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
	}

	return tempFile
}

// Helper for benchmarks that need temp files.
func createTempTasksFileB(b *testing.B, content string) string {
	b.Helper()

	tempDir := b.TempDir()
	tempFile := filepath.Join(tempDir, "test-tasks.yaml")

	if content != "" {
		err := os.WriteFile(tempFile, []byte(content), 0600)
		if err != nil {
			b.Fatalf("Failed to create temp file: %v", err)
		}
	}

	return tempFile
}

// Test saveTasksToFile with empty task list.
func TestSaveTasksEmpty(t *testing.T) {
	t.Parallel()

	tempFile := createTempTasksFile(t, "")

	watch := &Watch{Tasks: []*Task{}}

	err := watch.SaveTasksToFile(tempFile)
	if err != nil {
		t.Errorf("saveTasksToFile failed: %v", err)
	}

	// Verify file was created and contains empty array
	loadedWatch := &Watch{Tasks: []*Task{}}

	err = loadedWatch.LoadTasksFromFile(tempFile)
	if err != nil {
		t.Errorf("Failed to load saved file: %v", err)
	}

	if len(loadedWatch.Tasks) != 0 {
		t.Errorf("Expected empty task list, got %d tasks", len(loadedWatch.Tasks))
	}
}

// Test saveTasksToFile with single task.
func TestSaveTasksSingle(t *testing.T) {
	tempFile := createTempTasksFile(t, "")

	now := time.Now()
	watch := &Watch{Tasks: []*Task{
		{
			Name:        "Test Task",
			Description: "A test task description",
			Tags:        []string{"work", "urgent"},
			Segments: []*Segment{
				{
					Create: now,
					Finish: time.Time{}, // Unclosed segment
					Note:   "Working on test",
				},
			},
		},
	}}

	err := watch.SaveTasksToFile(tempFile)
	if err != nil {
		t.Errorf("saveTasksToFile failed: %v", err)
	}

	// Verify the saved content
	loadedWatch := &Watch{Tasks: []*Task{}}

	err = loadedWatch.LoadTasksFromFile(tempFile)
	if err != nil {
		t.Errorf("Failed to load saved file: %v", err)
	}

	if len(loadedWatch.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(loadedWatch.Tasks))
	}

	task := loadedWatch.Tasks[0]
	if task.Name != "Test Task" {
		t.Errorf("Expected task name 'Test Task', got '%s'", task.Name)
	}

	if task.Description != "A test task description" {
		t.Errorf("Expected description 'A test task description', got '%s'", task.Description)
	}

	if len(task.Tags) != 2 || task.Tags[0] != "work" || task.Tags[1] != "urgent" {
		t.Errorf("Expected tags [work, urgent], got %v", task.Tags)
	}

	if len(task.Segments) != 1 {
		t.Errorf("Expected 1 segment, got %d", len(task.Segments))
	}

	segment := task.Segments[0]
	if segment.Note != "Working on test" {
		t.Errorf("Expected segment note 'Working on test', got '%s'", segment.Note)
	}

	// Check that timestamps are preserved (within reasonable tolerance)
	if abs(segment.Create.Sub(now)) > time.Second {
		t.Errorf("Create timestamp not preserved correctly")
	}

	if !segment.Finish.IsZero() {
		t.Errorf("Expected zero finish time for unclosed segment")
	}
}

// Test saveTasksToFile with multiple complex tasks.
func TestSaveTasksMultiple(t *testing.T) {
	tempFile := createTempTasksFile(t, "")

	now1 := time.Now()
	now2 := now1.Add(time.Hour)

	tasks := []*Task{
		&Task{
			Name:        "Task 1",
			Description: "First task",
			Tags:        []string{"work"},
			Segments: []*Segment{
				{
					Create: now1,
					Finish: now2,
					Note:   "Completed work",
				},
				{
					Create: now2,
					Finish: time.Time{},
					Note:   "In progress",
				},
			},
		},
		&Task{
			Name:        "Task 2",
			Description: "Second task",
			Tags:        []string{},
			Segments:    []*Segment{},
		},
		&Task{
			Name:        "Task 3",
			Description: "",
			Tags:        []string{"personal", "hobby", "fun"},
			Segments: []*Segment{
				{
					Create: now1,
					Finish: now1.Add(30 * time.Minute),
					Note:   "",
				},
			},
		},
	}

	err := (&Watch{Tasks: tasks}).SaveTasksToFile(tempFile)
	if err != nil {
		t.Errorf("saveTasksToFile failed: %v", err)
	}

	// Load and verify
	loadedWatch := &Watch{Tasks: []*Task{}}; err = loadedWatch.LoadTasksFromFile(tempFile)
	if err != nil {
		t.Errorf("Failed to load saved file: %v", err)
	}

	if len(loadedWatch.Tasks) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(loadedWatch.Tasks))
	}

	// Verify first task
	task1 := loadedWatch.Tasks[0]
	if task1.Name != "Task 1" || len(task1.Segments) != 2 {
		t.Errorf("Task 1 not saved correctly")
	}

	// Verify second task (empty)
	task2 := loadedWatch.Tasks[1]
	if task2.Name != "Task 2" || len(task2.Tags) != 0 || len(task2.Segments) != 0 {
		t.Errorf("Task 2 not saved correctly")
	}

	// Verify third task
	task3 := loadedWatch.Tasks[2]
	if task3.Name != "Task 3" || len(task3.Tags) != 3 {
		t.Errorf("Task 3 not saved correctly")
	}
}

// Test loadTasksFromFile with non-existent file.
func TestLoadTasksNonExistent(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	nonExistentFile := filepath.Join(tempDir, "non-existent.yaml")

	watch := &Watch{Tasks: []*Task{}}; err := watch.LoadTasksFromFile(nonExistentFile)
	if err != nil {
		t.Errorf("loadTasksFromFile should not error on non-existent file, got: %v", err)
	}

	if len(watch.Tasks) != 0 {
		t.Errorf("Expected empty task list for non-existent file, got %d tasks", len(watch.Tasks))
	}
}

// Test loadTasksFromFile with empty file.
func TestLoadTasksEmpty(t *testing.T) {
	t.Parallel()

	tempFile := createTempTasksFile(t, "[]")

	watch := &Watch{Tasks: []*Task{}}; err := watch.LoadTasksFromFile(tempFile)
	if err != nil {
		t.Errorf("loadTasksFromFile failed: %v", err)
	}

	if len(watch.Tasks) != 0 {
		t.Errorf("Expected empty task list, got %d tasks", len(watch.Tasks))
	}
}

// Test loadTasksFromFile with valid YAML.
func TestLoadTasksValid(t *testing.T) {
	t.Parallel()

	yamlContent := `- name: Test Task
  description: A test description
  tags:
    - work
    - urgent
  segments:
    - create: 2024-01-01T10:00:00Z
      finish: 2024-01-01T11:00:00Z
      note: Completed segment
    - create: 2024-01-01T11:00:00Z
      finish: 0001-01-01T00:00:00Z
      note: Open segment
- name: Second Task
  description: ""
  tags: []
  segments: []`

	tempFile := createTempTasksFile(t, yamlContent)

	watch := &Watch{Tasks: []*Task{}}; err := watch.LoadTasksFromFile(tempFile)
	if err != nil {
		t.Errorf("loadTasksFromFile failed: %v", err)
	}

	if len(watch.Tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(watch.Tasks))
	}

	// Verify first task
	task1 := watch.Tasks[0]
	if task1.Name != "Test Task" {
		t.Errorf("Expected task name 'Test Task', got '%s'", task1.Name)
	}

	if len(task1.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(task1.Tags))
	}

	if len(task1.Segments) != 2 {
		t.Errorf("Expected 2 segments, got %d", len(task1.Segments))
	}

	// Check segment states
	seg1 := task1.Segments[0]
	if seg1.Finish.IsZero() {
		t.Errorf("First segment should be closed")
	}

	seg2 := task1.Segments[1]
	if !seg2.Finish.IsZero() {
		t.Errorf("Second segment should be open")
	}

	// Verify second task
	task2 := watch.Tasks[1]
	if task2.Name != "Second Task" {
		t.Errorf("Expected task name 'Second Task', got '%s'", task2.Name)
	}

	if len(task2.Segments) != 0 {
		t.Errorf("Expected 0 segments for second task, got %d", len(task2.Segments))
	}
}

// Test loadTasksFromFile with invalid YAML.
func TestLoadTasksInvalidYAML(t *testing.T) {
	t.Parallel()

	invalidYAML := `- name: Test Task
  invalid_yaml: [
  missing_closing_bracket`

	tempFile := createTempTasksFile(t, invalidYAML)

	watch := &Watch{Tasks: []*Task{}}; err := watch.LoadTasksFromFile(tempFile)
	if err == nil {
		t.Errorf("Expected error for invalid YAML, but got none")
	}

	// Tasks should remain unchanged (empty slice) when unmarshal fails
	if len(watch.Tasks) != 0 {
		t.Errorf("Expected empty tasks slice for invalid YAML, got %v", watch.Tasks)
	}
}

// Test round-trip: save then load.
func TestSaveLoadRoundTrip(t *testing.T) {
	t.Parallel()

	tempFile := createTempTasksFile(t, "")

	now := time.Now().Truncate(time.Second) // Truncate to avoid precision issues

	originalTasks := []*Task{
		&Task{
			Name:        "Round Trip Task",
			Description: "Testing save/load cycle",
			Tags:        []string{"test", "roundtrip"},
			Segments: []*Segment{
				{
					Create: now,
					Finish: now.Add(time.Hour),
					Note:   "First segment",
				},
				{
					Create: now.Add(time.Hour),
					Finish: time.Time{},
					Note:   "Second segment - open",
				},
			},
		},
	}

	// Save
	err := (&Watch{Tasks: originalTasks}).SaveTasksToFile(tempFile)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load
	loadedWatch := &Watch{Tasks: []*Task{}}; err = loadedWatch.LoadTasksFromFile(tempFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Compare
	if len(loadedWatch.Tasks) != len(originalTasks) {
		t.Fatalf("Task count mismatch: original %d, loaded %d", len(originalTasks), len(loadedWatch.Tasks))
	}

	original := originalTasks[0]
	loaded := loadedWatch.Tasks[0]

	if loaded.Name != original.Name {
		t.Errorf("Name mismatch: original '%s', loaded '%s'", original.Name, loaded.Name)
	}

	if loaded.Description != original.Description {
		t.Errorf("Description mismatch: original '%s', loaded '%s'", original.Description, loaded.Description)
	}

	if len(loaded.Tags) != len(original.Tags) {
		t.Errorf("Tags length mismatch: original %d, loaded %d", len(original.Tags), len(loaded.Tags))
	}

	for i, tag := range original.Tags {
		if loaded.Tags[i] != tag {
			t.Errorf("Tag %d mismatch: original '%s', loaded '%s'", i, tag, loaded.Tags[i])
		}
	}

	if len(loaded.Segments) != len(original.Segments) {
		t.Errorf("Segments length mismatch: original %d, loaded %d", len(original.Segments), len(loaded.Segments))
	}

	for index, originalSeg := range original.Segments {
		loadedSeg := loaded.Segments[index]

		// Time comparison with small tolerance
		if abs(loadedSeg.Create.Sub(originalSeg.Create)) > time.Second {
			t.Errorf("Segment %d Create time mismatch", index)
		}

		if originalSeg.Finish.IsZero() != loadedSeg.Finish.IsZero() {
			t.Errorf("Segment %d Finish zero state mismatch", index)
		}

		if !originalSeg.Finish.IsZero() && abs(loadedSeg.Finish.Sub(originalSeg.Finish)) > time.Second {
			t.Errorf("Segment %d Finish time mismatch", index)
		}

		if loadedSeg.Note != originalSeg.Note {
			t.Errorf("Segment %d Note mismatch: original '%s', loaded '%s'", index, originalSeg.Note, loadedSeg.Note)
		}
	}
}

// Test saving tasks with special characters.
func TestSaveTasksSpecialCharacters(t *testing.T) {
	t.Parallel()

	tempFile := createTempTasksFile(t, "")

	tasks := []*Task{
		&Task{
			Name:        "Task with special chars: 你好, émojis 🎉, & symbols!",
			Description: "Multi-line\ndescription with\ttabs and \"quotes\"",
			Tags:        []string{"unicode: 测试", "emoji: 🚀"},
			Segments: []*Segment{
				{
					Create: time.Now(),
					Finish: time.Time{},
					Note:   "Note with\nnewlines and 'quotes'",
				},
			},
		},
	}

	err := (&Watch{Tasks: tasks}).SaveTasksToFile(tempFile)
	if err != nil {
		t.Errorf("Failed to save tasks with special characters: %v", err)
	}

	loadedWatch := &Watch{Tasks: []*Task{}}; err = loadedWatch.LoadTasksFromFile(tempFile)
	if err != nil {
		t.Errorf("Failed to load tasks with special characters: %v", err)
	}

	if len(loadedWatch.Tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(loadedWatch.Tasks))
	}

	loaded := loadedWatch.Tasks[0]
	original := tasks[0]

	if loaded.Name != original.Name {
		t.Errorf("Special character name not preserved: expected '%s', got '%s'", original.Name, loaded.Name)
	}

	if loaded.Description != original.Description {
		t.Errorf("Special character description not preserved: expected '%s', got '%s'",
			original.Description, loaded.Description)
	}

	if len(loaded.Tags) != len(original.Tags) || loaded.Tags[0] != original.Tags[0] || loaded.Tags[1] != original.Tags[1] {
		t.Errorf("Special character tags not preserved: expected %v, got %v", original.Tags, loaded.Tags)
	}
}

// Test edge case: very large number of tasks.
func TestSaveLoadLargeTasks(t *testing.T) {
	t.Parallel()

	tempFile := createTempTasksFile(t, "")

	// Create 1000 tasks
	tasks := make([]*Task, 1000)
	for taskIndex := range 1000 {
		tasks[taskIndex] = &Task{
			Name:        fmt.Sprintf("Task %d", taskIndex),
			Description: fmt.Sprintf("Description for task %d", taskIndex),
			Tags:        []string{"tag1", "tag2"},
			Segments: []*Segment{
				{
					Create: time.Now().Add(time.Duration(taskIndex) * time.Minute),
					Finish: time.Now().Add(time.Duration(taskIndex+1) * time.Minute),
					Note:   fmt.Sprintf("Segment for task %d", taskIndex),
				},
			},
		}
	}

	err := (&Watch{Tasks: tasks}).SaveTasksToFile(tempFile)
	if err != nil {
		t.Errorf("Failed to save large task set: %v", err)
	}

	loadedWatch := &Watch{Tasks: []*Task{}}; err = loadedWatch.LoadTasksFromFile(tempFile)
	if err != nil {
		t.Errorf("Failed to load large task set: %v", err)
	}

	if len(loadedWatch.Tasks) != 1000 {
		t.Errorf("Expected 1000 tasks, got %d", len(loadedWatch.Tasks))
	}

	// Spot check a few tasks
	if loadedWatch.Tasks[0].Name != "Task 0" {
		t.Errorf("First task name incorrect")
	}

	if loadedWatch.Tasks[999].Name != "Task 999" {
		t.Errorf("Last task name incorrect")
	}
}

// Helper function for absolute duration.
func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}

	return d
}

// Benchmark saveTasksToFile.
func BenchmarkSaveTasks(b *testing.B) {
	tempFile := createTempTasksFileB(b, "")

	// Create a realistic task set
	tasks := make([]*Task, 100)
	now := time.Now()

	for taskIndex := range 100 {
		tasks[taskIndex] = &Task{
			Name:        "Benchmark Task",
			Description: "A task for benchmarking",
			Tags:        []string{"benchmark", "test"},
			Segments: []*Segment{
				{
					Create: now.Add(time.Duration(taskIndex) * time.Minute),
					Finish: now.Add(time.Duration(taskIndex+1) * time.Minute),
					Note:   "Benchmark segment",
				},
			},
		}
	}

	b.ResetTimer()

	for range b.N {
		_ = (&Watch{Tasks: tasks}).SaveTasksToFile(tempFile)
	}
}

// Benchmark loadTasksFromFile.
func BenchmarkLoadTasks(b *testing.B) {
	// First create a file with tasks
	tempFile := createTempTasksFileB(b, "")

	// Create and save tasks
	tasks := make([]*Task, 100)
	now := time.Now()

	for taskIndex := range 100 {
		tasks[taskIndex] = &Task{
			Name:        "Benchmark Task",
			Description: "A task for benchmarking",
			Tags:        []string{"benchmark", "test"},
			Segments: []*Segment{
				{
					Create: now.Add(time.Duration(taskIndex) * time.Minute),
					Finish: now.Add(time.Duration(taskIndex+1) * time.Minute),
					Note:   "Benchmark segment",
				},
			},
		}
	}

	_ = (&Watch{Tasks: tasks}).SaveTasksToFile(tempFile) // Save once

	b.ResetTimer()

	for range b.N {
		w := &Watch{Tasks: []*Task{}}; _ = w.LoadTasksFromFile(tempFile)
	}
}
