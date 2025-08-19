package task

import "time"

// HasSegmentsInRange checks if a task has any closed segments within the time range.
func (t *Task) HasSegmentsInRange(start, finish *time.Time) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	for _, segment := range t.Segments {
		if !segment.Finish.IsZero() {
			// Check if segment finished within the time range
			if start != nil && segment.Finish.Before(*start) {
				continue
			}

			if finish != nil && segment.Finish.After(*finish) {
				continue
			}
			// Found at least one segment in range
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
		if !segment.Finish.IsZero() {
			// Skip segments that finished before the start time
			if start != nil && segment.Finish.Before(*start) {
				continue
			}
			// Skip segments that finished after the finish time
			if finish != nil && segment.Finish.After(*finish) {
				continue
			}
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