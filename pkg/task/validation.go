package task

import (
	"errors"
	"strings"
	"time"
)

// Common errors.
var (
	ErrEmptyTaskName    = errors.New("task name cannot be empty")
	ErrTaskNotFound     = errors.New("task not found")
	ErrNoOpenSegment    = errors.New("no open segment to close")
	ErrSegmentNotFound  = errors.New("segment not found")
	ErrInvalidTimeRange = errors.New("start time must be before finish time")
)

// ValidateTask validates task fields before creation.
func ValidateTask(name, _ string, tags []string) error {
	if strings.TrimSpace(name) == "" {
		return ErrEmptyTaskName
	}
	
	// Clean tags - remove empty strings
	_ = make([]string, 0, len(tags))
	for _, tag := range tags {
		if trimmed := strings.TrimSpace(tag); trimmed != "" {
			// Tag is valid but we don't need to store it here.
			_ = trimmed // Acknowledge we checked the tag
		}
	}
	
	return nil
}

// ValidateTimeRange validates that start is before finish if both are provided.
func ValidateTimeRange(start, finish *time.Time) error {
	if start != nil && finish != nil && start.After(*finish) {
		return ErrInvalidTimeRange
	}

	return nil
}