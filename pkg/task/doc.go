/*
Package task provides time tracking functionality for the ohgmas-watch application.

The package offers three main types:
  - Watch: A collection of tasks with thread-safe operations
  - Task: An individual work item with time segments
  - Segment: A time period with start/finish times and optional notes

Basic usage:

	watch := &task.Watch{Tasks: []*task.Task{}}
	watch.AddTask("Feature Development", "Working on new feature", []string{"work", "coding"})

	// Start a work segment
	watch.Tasks[0].AddSegment("Implementing API endpoints")

	// Later, close the segment
	watch.Tasks[0].CloseSegment()

	// Save to file
	err := watch.SaveTasks()

Thread Safety:

All public methods are thread-safe and can be called concurrently.
The package uses sync.RWMutex internally for synchronization.

File Operations:

Tasks are persisted in YAML format to ~/.ohgmas-tasks.yaml by default.
Custom paths can be specified using SaveTasksToFile and LoadTasksFromFile.
*/
package task
