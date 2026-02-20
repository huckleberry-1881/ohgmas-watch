package task

import "time"

// isSegmentInRange checks if a closed segment falls within the specified time range.
// Returns true if the segment is closed and its finish time is within the range.
// Uses exclusive lower bound (segment.Finish > start) to match GetThisWeekDuration logic.
func isSegmentInRange(segment *Segment, start, finish *time.Time) bool {
	// Only consider closed segments
	if segment.Finish.IsZero() {
		return false
	}

	// Check if segment finished at or before the start time (exclusive lower bound)
	if start != nil && !segment.Finish.After(*start) {
		return false
	}

	// Check if segment finished after the finish time (inclusive upper bound)
	// A segment is included if: start < segment.Finish <= finish
	if finish != nil && segment.Finish.After(*finish) {
		return false
	}

	return true
}

// HasSegmentsInRange checks if a task has any closed segments within the time range.
func (t *Task) HasSegmentsInRange(start, finish *time.Time) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, segment := range t.Segments {
		if isSegmentInRange(segment, start, finish) {
			return true
		}
	}

	return false
}

// GetFilteredClosedSegmentsDuration gets filtered closed segments duration within a time range.
func (t *Task) GetFilteredClosedSegmentsDuration(start, finish *time.Time) time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var totalDuration time.Duration

	for _, segment := range t.Segments {
		if isSegmentInRange(segment, start, finish) {
			totalDuration += segment.Finish.Sub(segment.Create)
		}
	}

	return totalDuration
}

// GetLastActivity returns the last activity time for a task.
func (t *Task) GetLastActivity() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.Segments) == 0 {
		return time.Time{}
	}

	lastSegment := t.Segments[len(t.Segments)-1]
	if lastSegment.Finish.IsZero() {
		return lastSegment.Create // Use start time for open segments
	}

	return lastSegment.Finish // Use end time for closed segments
}

// IsActive returns true if the task has an unclosed segment.
func (t *Task) IsActive() bool {
	return t.HasUnclosedSegment()
}

// GetCurrentSegmentDuration returns the duration of the current open segment.
func (t *Task) GetCurrentSegmentDuration() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.Segments) == 0 {
		return 0
	}

	lastSegment := t.Segments[len(t.Segments)-1]
	if lastSegment.Finish.IsZero() {
		return time.Since(lastSegment.Create)
	}

	return 0
}

// GetLastSegment returns the last segment of the task, or nil if no segments.
func (t *Task) GetLastSegment() *Segment {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.Segments) == 0 {
		return nil
	}

	return t.Segments[len(t.Segments)-1]
}

// GetThisWeekDuration calculates total duration of closed segments completed since the given start time.
func (t *Task) GetThisWeekDuration(weekStart time.Time) time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var totalDuration time.Duration

	for _, segment := range t.Segments {
		// Only include closed segments that finished after the week start
		if !segment.Finish.IsZero() && segment.Finish.After(weekStart) {
			totalDuration += segment.Finish.Sub(segment.Create)
		}
	}

	return totalDuration
}
