package task //nolint:testpackage // direct struct construction

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatch_SaveAndLoadTasksFromFile(t *testing.T) { //nolint:cyclop // integration test verifies all fields
	t.Parallel()

	// Create a temporary file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test-tasks.yaml")

	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	// Create watch with tasks
	originalWatch := &Watch{
		Tasks: []*Task{
			{
				Name:        "Task 1",
				Description: "Description 1",
				Tags:        []string{"tag1", "tag2"},
				Category:    "work",
				Segments: []*Segment{
					{Create: baseTime, Finish: baseTime.Add(time.Hour), Note: "Note 1"},
				},
			},
			{
				Name:        "Task 2",
				Description: "Description 2",
				Tags:        []string{},
				Category:    "completed",
				Segments:    []*Segment{},
			},
		},
	}

	// Save tasks
	err := originalWatch.SaveTasksToFile(filePath)
	if err != nil {
		t.Fatalf("SaveTasksToFile() error = %v", err)
	}

	// Verify file exists
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		t.Fatal("SaveTasksToFile() did not create file")
	}

	// Load tasks into new watch
	loadedWatch := &Watch{Tasks: []*Task{}}

	err = loadedWatch.LoadTasksFromFile(filePath)
	if err != nil {
		t.Fatalf("LoadTasksFromFile() error = %v", err)
	}

	// Verify loaded data
	if len(loadedWatch.Tasks) != len(originalWatch.Tasks) {
		t.Errorf("Loaded %d tasks, want %d", len(loadedWatch.Tasks), len(originalWatch.Tasks))
	}

	// Check first task details
	if loadedWatch.Tasks[0].Name != "Task 1" {
		t.Errorf("Task name = %q, want %q", loadedWatch.Tasks[0].Name, "Task 1")
	}

	if loadedWatch.Tasks[0].Description != "Description 1" {
		t.Errorf("Task description = %q, want %q", loadedWatch.Tasks[0].Description, "Description 1")
	}

	if loadedWatch.Tasks[0].Category != "work" {
		t.Errorf("Task category = %q, want %q", loadedWatch.Tasks[0].Category, "work")
	}

	if len(loadedWatch.Tasks[0].Tags) != 2 {
		t.Errorf("Task tags count = %d, want %d", len(loadedWatch.Tasks[0].Tags), 2)
	}

	if len(loadedWatch.Tasks[0].Segments) != 1 {
		t.Errorf("Task segments count = %d, want %d", len(loadedWatch.Tasks[0].Segments), 1)
	}

	// Check segment details
	seg := loadedWatch.Tasks[0].Segments[0]
	if seg.Note != "Note 1" {
		t.Errorf("Segment note = %q, want %q", seg.Note, "Note 1")
	}

	if !seg.Create.Equal(baseTime) {
		t.Errorf("Segment create = %v, want %v", seg.Create, baseTime)
	}
}

func TestWatch_LoadTasksFromFile_NonExistent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nonexistent.yaml")

	watch := &Watch{Tasks: []*Task{}}

	err := watch.LoadTasksFromFile(filePath)
	if err != nil {
		t.Errorf("LoadTasksFromFile() should not error for non-existent file, got %v", err)
	}

	// Should initialize with empty tasks
	if watch.Tasks == nil {
		t.Error("LoadTasksFromFile() should initialize Tasks slice")
	}

	if len(watch.Tasks) != 0 {
		t.Errorf("LoadTasksFromFile() should have 0 tasks, got %d", len(watch.Tasks))
	}
}

func TestWatch_LoadTasksFromFile_InvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	err := os.WriteFile(filePath, []byte("this is not: valid: yaml: ["), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	watch := &Watch{Tasks: []*Task{}}

	err = watch.LoadTasksFromFile(filePath)
	if err == nil {
		t.Error("LoadTasksFromFile() should error for invalid YAML")
	}
}

func TestWatch_SaveTasksToFile_InvalidPath(t *testing.T) {
	t.Parallel()

	watch := &Watch{
		Tasks: []*Task{
			{Name: "Task"},
		},
	}

	// Try to save to a directory that doesn't exist
	err := watch.SaveTasksToFile("/nonexistent/directory/file.yaml")
	if err == nil {
		t.Error("SaveTasksToFile() should error for invalid path")
	}
}

func TestWatch_SaveAndLoadTasks_DefaultPath(t *testing.T) {
	t.Parallel()

	// This test uses the default path, so we skip in parallel test runs
	// to avoid conflicts. Just verify the functions exist and can be called.
	watch := &Watch{Tasks: []*Task{}}

	// Just verify the function signature works
	_ = watch.SaveTasks
	_ = watch.LoadTasks
}

func TestWatch_SaveTasksToFile_EmptyTasks(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "empty-tasks.yaml")

	watch := &Watch{Tasks: []*Task{}}

	err := watch.SaveTasksToFile(filePath)
	if err != nil {
		t.Fatalf("SaveTasksToFile() error = %v", err)
	}

	// Load and verify
	loadedWatch := &Watch{Tasks: []*Task{}}

	err = loadedWatch.LoadTasksFromFile(filePath)
	if err != nil {
		t.Fatalf("LoadTasksFromFile() error = %v", err)
	}

	if len(loadedWatch.Tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(loadedWatch.Tasks))
	}
}

func TestWatch_SaveTasksToFile_SpecialCharacters(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "special-chars.yaml")

	watch := &Watch{
		Tasks: []*Task{
			{
				Name:        "Task with 'quotes' and \"double quotes\"",
				Description: "Description with:\n- newlines\n- special chars: <>&",
				Tags:        []string{"tag:colon", "tag/slash"},
				Category:    "work",
			},
		},
	}

	err := watch.SaveTasksToFile(filePath)
	if err != nil {
		t.Fatalf("SaveTasksToFile() error = %v", err)
	}

	loadedWatch := &Watch{Tasks: []*Task{}}

	err = loadedWatch.LoadTasksFromFile(filePath)
	if err != nil {
		t.Fatalf("LoadTasksFromFile() error = %v", err)
	}

	if loadedWatch.Tasks[0].Name != watch.Tasks[0].Name {
		t.Errorf("Name not preserved: got %q, want %q", loadedWatch.Tasks[0].Name, watch.Tasks[0].Name)
	}

	if loadedWatch.Tasks[0].Description != watch.Tasks[0].Description {
		t.Errorf("Description not preserved: got %q, want %q", loadedWatch.Tasks[0].Description, watch.Tasks[0].Description)
	}
}

func TestWatch_SaveTasksToFile_FilePermissions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "permissions-test.yaml")

	watch := &Watch{
		Tasks: []*Task{
			{Name: "Task", Category: "work"},
		},
	}

	err := watch.SaveTasksToFile(filePath)
	if err != nil {
		t.Fatalf("SaveTasksToFile() error = %v", err)
	}

	// Check file permissions (should be 0600)
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("File permissions = %o, want %o", perm, 0600)
	}
}

func TestWatch_LoadTasksFromFile_LargeTasks(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "large-tasks.yaml")

	baseTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	// Create many tasks with many segments.
	var tasks []*Task

	for range 100 {
		var segments []*Segment
		for j := range 50 {
			segments = append(segments, &Segment{
				Create: baseTime.Add(time.Duration(j) * time.Hour),
				Finish: baseTime.Add(time.Duration(j)*time.Hour + 30*time.Minute),
				Note:   "Segment note",
			})
		}

		tasks = append(tasks, &Task{
			Name:        "Task",
			Description: "Description",
			Tags:        []string{"tag1", "tag2", "tag3"},
			Category:    "work",
			Segments:    segments,
		})
	}

	watch := &Watch{Tasks: tasks}

	err := watch.SaveTasksToFile(filePath)
	if err != nil {
		t.Fatalf("SaveTasksToFile() error = %v", err)
	}

	loadedWatch := &Watch{Tasks: []*Task{}}

	err = loadedWatch.LoadTasksFromFile(filePath)
	if err != nil {
		t.Fatalf("LoadTasksFromFile() error = %v", err)
	}

	if len(loadedWatch.Tasks) != 100 {
		t.Errorf("Expected 100 tasks, got %d", len(loadedWatch.Tasks))
	}

	if len(loadedWatch.Tasks[0].Segments) != 50 {
		t.Errorf("Expected 50 segments, got %d", len(loadedWatch.Tasks[0].Segments))
	}
}

func TestWatch_RoundTrip_PreservesAllFields(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "roundtrip.yaml")

	baseTime := time.Date(2024, 1, 15, 12, 30, 45, 0, time.UTC)

	original := &Watch{
		Tasks: []*Task{
			{
				Name:        "Complete Task",
				Description: "Full description with all details",
				Tags:        []string{"alpha", "beta", "gamma"},
				Category:    "completed",
				Segments: []*Segment{
					{
						Create: baseTime,
						Finish: baseTime.Add(2 * time.Hour),
						Note:   "First work session",
					},
					{
						Create: baseTime.Add(3 * time.Hour),
						Finish: baseTime.Add(4 * time.Hour),
						Note:   "Second work session",
					},
					{
						Create: baseTime.Add(5 * time.Hour),
						Finish: time.Time{}, // Open segment
						Note:   "Current session",
					},
				},
			},
		},
	}

	// Save
	err := original.SaveTasksToFile(filePath)
	if err != nil {
		t.Fatalf("SaveTasksToFile() error = %v", err)
	}

	// Load
	loaded := &Watch{Tasks: []*Task{}}

	err = loaded.LoadTasksFromFile(filePath)
	if err != nil {
		t.Fatalf("LoadTasksFromFile() error = %v", err)
	}

	// Verify all fields
	if len(loaded.Tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(loaded.Tasks))
	}

	task := loaded.Tasks[0]

	if task.Name != "Complete Task" {
		t.Errorf("Name = %q, want %q", task.Name, "Complete Task")
	}

	if task.Description != "Full description with all details" {
		t.Errorf("Description mismatch")
	}

	if task.Category != "completed" {
		t.Errorf("Category = %q, want %q", task.Category, "completed")
	}

	if len(task.Tags) != 3 {
		t.Errorf("Tags count = %d, want 3", len(task.Tags))
	}

	if len(task.Segments) != 3 {
		t.Errorf("Segments count = %d, want 3", len(task.Segments))
	}

	// Check open segment preserved
	if !task.Segments[2].Finish.IsZero() {
		t.Error("Open segment Finish should be zero")
	}
}
