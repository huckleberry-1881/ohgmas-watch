package main

import (
	"fmt"
	"time"
)

// formatDuration formats a duration into a human-readable string.
func formatDuration(d time.Duration) string {
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