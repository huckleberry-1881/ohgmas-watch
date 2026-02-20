package main

import (
	"fmt"
	"os"
	"time"

	"github.com/huckleberry-1881/ohgmas-watch/pkg/task"
)

// generateSummary generates and prints a weekly summary grouped by tagset.
func generateSummary(includeTasks bool, start, finish *time.Time, filePath string) error {
	watch, err := loadWatchForSummary(filePath)
	if err != nil {
		return err
	}

	earliest, latest := watch.GetEarliestAndLatestSegmentTimes()
	if earliest.IsZero() {
		_, _ = fmt.Fprintf(os.Stdout, "No segments found\n")

		return nil
	}

	filterStart, filterFinish := getTimeFilters(start, finish, earliest, latest)
	weekStarts := getWeekStarts(filterStart, filterFinish)
	weeklySummaries := getWeeklySummaries(watch, weekStarts, includeTasks)

	printWeeklySummaries(weeklySummaries, includeTasks)

	return nil
}

// loadWatchForSummary loads the watch from the specified file or default location.
func loadWatchForSummary(filePath string) (*task.Watch, error) {
	watch := &task.Watch{
		Tasks: []*task.Task{},
	}

	var err error

	if filePath != "" {
		err = watch.LoadTasksFromFile(filePath)
	} else {
		err = watch.LoadTasks()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	return watch, nil
}

// getTimeFilters returns the start and finish times for filtering, using defaults if not provided.
func getTimeFilters(start, finish *time.Time, earliest, latest time.Time) (time.Time, time.Time) {
	filterStart := earliest
	if start != nil {
		filterStart = *start
	}

	filterFinish := latest
	if finish != nil {
		filterFinish = *finish
	}

	return filterStart, filterFinish
}

// getWeeklySummaries retrieves weekly summaries based on whether tasks should be included.
func getWeeklySummaries(watch *task.Watch, weekStarts []time.Time, includeTasks bool) []task.WeeklySummary {
	if includeTasks {
		return watch.GetWeeklySummaryByTagsetWithTasks(weekStarts)
	}

	return watch.GetWeeklySummaryByTagset(weekStarts)
}

// printWeeklySummaries prints the weekly summaries to stdout.
func printWeeklySummaries(weeklySummaries []task.WeeklySummary, includeTasks bool) {
	for _, weeklySummary := range weeklySummaries {
		weekStartStr := weeklySummary.WeekStart.Format("01/02/2006")
		_, _ = fmt.Fprintf(os.Stdout, "Week starting %s\n", weekStartStr)

		for _, tagsetSummary := range weeklySummary.Tagsets {
			durationStr := formatDuration(tagsetSummary.Duration)
			_, _ = fmt.Fprintf(os.Stdout, "- %s [%s]\n", tagsetSummary.Tagset, durationStr)

			if includeTasks {
				printTasksForTagset(weeklySummary.WeekStart, tagsetSummary.Tasks)
			}
		}

		_, _ = fmt.Fprintf(os.Stdout, "\n")
	}
}

// printTasksForTagset prints the individual tasks for a tagset.
func printTasksForTagset(weekStart time.Time, tasks []*task.Task) {
	weekEnd := weekStart.AddDate(0, 0, 7)

	for _, taskItem := range tasks {
		taskDuration := taskItem.GetFilteredClosedSegmentsDuration(&weekStart, &weekEnd)
		taskDurationStr := formatDuration(taskDuration)
		_, _ = fmt.Fprintf(os.Stdout, "-- %s [%s]\n", taskItem.Name, taskDurationStr)
	}
}
