package main

import (
	"fmt"
	"time"
)

// formatDuration formats a duration into a human-readable string.
// Returns "0m" for zero durations.
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0m"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%02dm", hours, minutes)
	}

	return fmt.Sprintf("%dm", minutes)
}

// parseTimeFlags parses start and finish time flags.
func parseTimeFlags(startFlag, finishFlag string) (*time.Time, *time.Time, error) {
	var start, finish *time.Time

	if startFlag != "" {
		parsedStart, err := time.Parse(time.RFC3339, startFlag)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing start time: %w", err)
		}

		start = &parsedStart
	}

	if finishFlag != "" {
		parsedFinish, err := time.Parse(time.RFC3339, finishFlag)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing finish time: %w", err)
		}

		finish = &parsedFinish
	}

	return start, finish, nil
}

// getMondayOfWeek returns the Monday of the week containing the given time at 00:00:00.
func getMondayOfWeek(t time.Time) time.Time {
	// Get the current weekday (0 = Sunday, 1 = Monday, etc.)
	currentWeekday := int(t.Weekday())

	// Calculate days to subtract to get to Monday
	var daysBack int
	if currentWeekday == 0 { // Sunday
		daysBack = 6
	} else { // Monday (1) through Saturday (6)
		daysBack = currentWeekday - 1
	}

	// Get Monday's date
	monday := t.AddDate(0, 0, -daysBack)

	// Set to beginning of day (00:00:00)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
}

// getLastMonday returns the time of the most recent Monday at 00:00:00.
func getLastMonday() time.Time {
	return getMondayOfWeek(time.Now())
}

// getWeekStarts returns all Monday dates from earliest to latest covering the time range.
// If start is nil, uses the earliest segment date. If finish is nil, uses now.
func getWeekStarts(earliestSegment, latestSegment time.Time) []time.Time {
	// Get the Monday of the week containing the earliest segment
	weekStart := getMondayOfWeek(earliestSegment)
	weekEnd := getMondayOfWeek(latestSegment)

	var weeks []time.Time
	for current := weekStart; !current.After(weekEnd); current = current.AddDate(0, 0, 7) {
		weeks = append(weeks, current)
	}

	return weeks
}
