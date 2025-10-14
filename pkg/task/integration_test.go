package task_test

import (
	"testing"
	"time"

	"github.com/huckleberry-1881/ohgmas-watch/pkg/task"
)

// TestTaskWorkflow tests a complete workflow of task operations.
func TestTaskWorkflow(t *testing.T) {
	t.Parallel()

	// Create a new watch
	watch := &task.Watch{Tasks: []*task.Task{}}

	// Add a task
	watch.AddTask("Feature Development", "Implement new user authentication", []string{"development", "security"}, "work")

	if len(watch.Tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(watch.Tasks))
	}

	task := watch.Tasks[0]

	// Verify task was created correctly
	if task.Name != "Feature Development" {
		t.Errorf("Expected task name 'Feature Development', got '%s'", task.Name)
	}

	if len(task.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(task.Tags))
	}

	// Task should have no unclosed segments initially
	if task.HasUnclosedSegment() {
		t.Error("New task should not have unclosed segments")
	}

	// Duration should be zero initially
	if duration := task.GetClosedSegmentsDuration(); duration != 0 {
		t.Errorf("Expected zero duration, got %v", duration)
	}

	// Start working on the task
	task.AddSegment("Started implementing login form")

	// Now should have an unclosed segment
	if !task.HasUnclosedSegment() {
		t.Error("Task should have unclosed segment after adding segment")
	}

	// Duration should still be zero (segment is open)
	if duration := task.GetClosedSegmentsDuration(); duration != 0 {
		t.Errorf("Expected zero duration with open segment, got %v", duration)
	}

	// Wait a bit and add another segment
	time.Sleep(1 * time.Millisecond)
	task.AddSegment("Working on password validation")

	// Should still have unclosed segments
	if !task.HasUnclosedSegment() {
		t.Error("Task should still have unclosed segments")
	}

	if len(task.Segments) != 2 {
		t.Errorf("Expected 2 segments, got %d", len(task.Segments))
	}

	// Close all segments
	task.CloseSegment()

	// Should no longer have unclosed segments
	if task.HasUnclosedSegment() {
		t.Error("Task should not have unclosed segments after closing")
	}

	// Duration should now be greater than zero
	if duration := task.GetClosedSegmentsDuration(); duration <= 0 {
		t.Errorf("Expected positive duration after closing segments, got %v", duration)
	}

	// Verify all segments are closed
	for segmentIndex, segment := range task.Segments {
		if segment.Finish.IsZero() {
			t.Errorf("Segment %d should be closed but has zero finish time", segmentIndex)
		}

		if segment.Finish.Before(segment.Create) {
			t.Errorf("Segment %d finish time (%v) is before create time (%v)", segmentIndex, segment.Finish, segment.Create)
		}
	}
}

// TestMultipleTasksWorkflow tests working with multiple tasks.
func TestMultipleTasksWorkflow(t *testing.T) {
	t.Parallel()

	watch := &task.Watch{Tasks: []*task.Task{}}

	// Add multiple tasks
	taskData := []struct {
		name        string
		description string
		tags        []string
	}{
		{"Task 1", "First task", []string{"tag1"}},
		{"Task 2", "Second task", []string{"tag2", "urgent"}},
		{"Task 3", "Third task", []string{}},
	}

	for _, td := range taskData {
		watch.AddTask(td.name, td.description, td.tags, "work")
	}

	if len(watch.Tasks) != len(taskData) {
		t.Fatalf("Expected %d tasks, got %d", len(taskData), len(watch.Tasks))
	}

	// Start work on different tasks
	watch.Tasks[0].AddSegment("Working on task 1")
	watch.Tasks[1].AddSegment("Working on task 2")
	// Leave task 3 without segments

	// Verify states
	if !watch.Tasks[0].HasUnclosedSegment() {
		t.Error("Task 1 should have unclosed segment")
	}

	if !watch.Tasks[1].HasUnclosedSegment() {
		t.Error("Task 2 should have unclosed segment")
	}

	if watch.Tasks[2].HasUnclosedSegment() {
		t.Error("Task 3 should not have unclosed segment")
	}

	// Close segments on task 1
	watch.Tasks[0].CloseSegment()

	if watch.Tasks[0].HasUnclosedSegment() {
		t.Error("Task 1 should not have unclosed segment after closing")
	}

	if !watch.Tasks[1].HasUnclosedSegment() {
		t.Error("Task 2 should still have unclosed segment")
	}

	// Verify durations
	if watch.Tasks[0].GetClosedSegmentsDuration() <= 0 {
		t.Error("Task 1 should have positive duration")
	}

	if watch.Tasks[1].GetClosedSegmentsDuration() != 0 {
		t.Error("Task 2 should have zero duration (open segment)")
	}

	if watch.Tasks[2].GetClosedSegmentsDuration() != 0 {
		t.Error("Task 3 should have zero duration (no segments)")
	}
}

// TestSegmentTimingAccuracy tests the accuracy of segment timing.
func TestSegmentTimingAccuracy(t *testing.T) {
	t.Parallel()

	testTask := &task.Task{
		Name:        "Timing Test",
		Description: "Test task for timing accuracy",
		Tags:        []string{},
		Segments:    []*task.Segment{},
	}

	// Record start time
	startTime := time.Now()

	// Add a segment
	testTask.AddSegment("Test segment")

	// Wait a measurable amount of time
	time.Sleep(10 * time.Millisecond)

	// Close the segment
	testTask.CloseSegment()

	endTime := time.Now()

	if len(testTask.Segments) != 1 {
		t.Fatalf("Expected 1 segment, got %d", len(testTask.Segments))
	}

	segment := testTask.Segments[0]

	// Verify timing is reasonable
	if segment.Create.Before(startTime) {
		t.Error("Segment create time is before test start time")
	}

	if segment.Finish.After(endTime) {
		t.Error("Segment finish time is after test end time")
	}

	if segment.Finish.Before(segment.Create) {
		t.Error("Segment finish time is before create time")
	}

	// Duration should be at least what we waited
	duration := segment.Finish.Sub(segment.Create)
	if duration < 5*time.Millisecond {
		t.Errorf("Expected duration >= 5ms, got %v", duration)
	}

	// Verify task reports correct duration
	taskDuration := testTask.GetClosedSegmentsDuration()
	if taskDuration != duration {
		t.Errorf("Task duration (%v) doesn't match segment duration (%v)", taskDuration, duration)
	}
}

// TestConcurrentSegmentOperations tests concurrent operations on segments.
func TestConcurrentSegmentOperations(t *testing.T) {
	t.Parallel()

	testTask := &task.Task{
		Name:        "Concurrent Test",
		Description: "Test task for concurrent operations",
		Tags:        []string{},
		Segments:    []*task.Segment{},
	}

	// Add multiple segments concurrently
	const numSegments = 10

	done := make(chan bool, numSegments)

	for segmentIndex := range numSegments {
		go func(_ int) {
			testTask.AddSegment("Concurrent segment")

			done <- true
		}(segmentIndex)
	}

	// Wait for all goroutines to complete
	for range numSegments {
		<-done
	}

	if len(testTask.Segments) != numSegments {
		t.Errorf("Expected %d segments, got %d", numSegments, len(testTask.Segments))
	}

	// All segments should be open
	if !testTask.HasUnclosedSegment() {
		t.Error("Task should have unclosed segments")
	}

	// Close all segments
	testTask.CloseSegment()

	// No segments should be open
	if testTask.HasUnclosedSegment() {
		t.Error("Task should not have unclosed segments after closing")
	}

	// All segments should have valid timing
	for segmentIndex, segment := range testTask.Segments {
		if segment.Finish.IsZero() {
			t.Errorf("Segment %d should be closed", segmentIndex)
		}

		if segment.Finish.Before(segment.Create) {
			t.Errorf("Segment %d finish time is before create time", segmentIndex)
		}
	}
}

// TestEdgeCaseEmptyOperations tests operations on empty or nil structures.
func TestEdgeCaseEmptyOperations(t *testing.T) {
	t.Parallel()

	t.Run("operations on task with nil segments", func(t *testing.T) {
		t.Parallel()

		testTask := &task.Task{
			Name:        "Test Task",
			Description: "Test task with nil segments",
			Tags:        []string{},
			Segments:    nil, // nil segments slice
		}

		// These operations should not panic
		if testTask.HasUnclosedSegment() {
			t.Error("Task with nil segments should not have unclosed segments")
		}

		if duration := testTask.GetClosedSegmentsDuration(); duration != 0 {
			t.Errorf("Task with nil segments should have zero duration, got %v", duration)
		}

		// CloseSegment should not panic
		testTask.CloseSegment()

		// AddSegment should work and initialize the slice
		testTask.AddSegment("First segment")

		if len(testTask.Segments) != 1 {
			t.Errorf("Expected 1 segment after AddSegment, got %d", len(testTask.Segments))
		}
	})

	t.Run("operations on empty watch", func(t *testing.T) {
		t.Parallel()

		watch := &task.Watch{Tasks: []*task.Task{}}

		// Should be able to add tasks
		watch.AddTask("Test", "Description", []string{"tag"}, "work")

		if len(watch.Tasks) != 1 {
			t.Errorf("Expected 1 task, got %d", len(watch.Tasks))
		}
	})

	t.Run("operations on watch with nil tasks", func(t *testing.T) {
		t.Parallel()

		watch := &task.Watch{Tasks: nil}

		// AddTask should work and initialize the slice
		watch.AddTask("Test", "Description", []string{"tag"}, "work")

		if len(watch.Tasks) != 1 {
			t.Errorf("Expected 1 task after AddTask, got %d", len(watch.Tasks))
		}
	})
}

// TestLargeNumberOfOperations tests performance with many operations.
func TestLargeNumberOfOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	t.Parallel()

	watch := &task.Watch{Tasks: []*task.Task{}}

	const numTasks = 100

	const segmentsPerTask = 50

	// Add many tasks

	for range numTasks {
		watch.AddTask("Task", "Description", []string{"tag"}, "work")
	}

	if len(watch.Tasks) != numTasks {
		t.Fatalf("Expected %d tasks, got %d", numTasks, len(watch.Tasks))
	}

	// Add many segments to each task
	for _, task := range watch.Tasks {
		for range segmentsPerTask {
			task.AddSegment("Segment")
		}
	}

	// Verify all tasks have the expected number of segments
	for taskIndex, task := range watch.Tasks {
		if len(task.Segments) != segmentsPerTask {
			t.Errorf("Task %d: expected %d segments, got %d", taskIndex, segmentsPerTask, len(task.Segments))
		}

		if !task.HasUnclosedSegment() {
			t.Errorf("Task %d should have unclosed segments", taskIndex)
		}
	}

	// Close all segments on all tasks
	for _, task := range watch.Tasks {
		task.CloseSegment()
	}

	// Verify all segments are closed and calculate total duration
	var totalDuration time.Duration

	for taskIndex, task := range watch.Tasks {
		if task.HasUnclosedSegment() {
			t.Errorf("Task %d should not have unclosed segments", taskIndex)
		}

		duration := task.GetClosedSegmentsDuration()
		if duration <= 0 {
			t.Errorf("Task %d should have positive duration", taskIndex)
		}

		totalDuration += duration
	}

	if totalDuration <= 0 {
		t.Error("Total duration should be positive")
	}
}
