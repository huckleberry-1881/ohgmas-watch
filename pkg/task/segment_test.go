package task

import (
	"testing"
	"time"
)

func TestHasSegmentsInRange(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		segments []*Segment
		start    *time.Time
		finish   *time.Time
		expected bool
	}{
		{
			name:     "no segments",
			segments: []*Segment{},
			start:    nil,
			finish:   nil,
			expected: false,
		},
		{
			name: "open segment with no time range",
			segments: []*Segment{
				{Create: baseTime, Finish: time.Time{}},
			},
			start:    nil,
			finish:   nil,
			expected: false,
		},
		{
			name: "closed segment with no time range",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
			},
			start:    nil,
			finish:   nil,
			expected: true,
		},
		{
			name: "segment before start time",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
			},
			start:    func() *time.Time { t := baseTime.Add(2 * time.Hour); return &t }(),
			finish:   nil,
			expected: false,
		},
		{
			name: "segment after finish time",
			segments: []*Segment{
				{Create: baseTime.Add(3 * time.Hour), Finish: baseTime.Add(4 * time.Hour)},
			},
			start:    nil,
			finish:   func() *time.Time { t := baseTime.Add(2 * time.Hour); return &t }(),
			expected: false,
		},
		{
			name: "segment within range",
			segments: []*Segment{
				{Create: baseTime.Add(time.Hour), Finish: baseTime.Add(2 * time.Hour)},
			},
			start:    func() *time.Time { t := baseTime; return &t }(),
			finish:   func() *time.Time { t := baseTime.Add(3 * time.Hour); return &t }(),
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			task := &Task{
				Name:     "Test Task",
				Segments: tc.segments,
			}

			result := task.HasSegmentsInRange(tc.start, tc.finish)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestGetFilteredClosedSegmentsDuration(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		segments []*Segment
		start    *time.Time
		finish   *time.Time
		expected time.Duration
	}{
		{
			name:     "no segments",
			segments: []*Segment{},
			start:    nil,
			finish:   nil,
			expected: 0,
		},
		{
			name: "all segments in range",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
				{Create: baseTime.Add(2 * time.Hour), Finish: baseTime.Add(3 * time.Hour)},
			},
			start:    nil,
			finish:   nil,
			expected: 2 * time.Hour,
		},
		{
			name: "filter by start time",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
				{Create: baseTime.Add(2 * time.Hour), Finish: baseTime.Add(3 * time.Hour)},
			},
			start:    func() *time.Time { t := baseTime.Add(90 * time.Minute); return &t }(),
			finish:   nil,
			expected: time.Hour,
		},
		{
			name: "filter by finish time",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
				{Create: baseTime.Add(2 * time.Hour), Finish: baseTime.Add(3 * time.Hour)},
			},
			start:    nil,
			finish:   func() *time.Time { t := baseTime.Add(90 * time.Minute); return &t }(),
			expected: time.Hour,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			task := &Task{
				Name:     "Test Task",
				Segments: tc.segments,
			}

			result := task.GetFilteredClosedSegmentsDuration(tc.start, tc.finish)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestGetLastActivity(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		segments []*Segment
		expected time.Time
	}{
		{
			name:     "no segments",
			segments: []*Segment{},
			expected: time.Time{},
		},
		{
			name: "open segment",
			segments: []*Segment{
				{Create: baseTime, Finish: time.Time{}},
			},
			expected: baseTime,
		},
		{
			name: "closed segment",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
			},
			expected: baseTime.Add(time.Hour),
		},
		{
			name: "multiple segments with last open",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
				{Create: baseTime.Add(2 * time.Hour), Finish: time.Time{}},
			},
			expected: baseTime.Add(2 * time.Hour),
		},
		{
			name: "multiple segments with last closed",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
				{Create: baseTime.Add(2 * time.Hour), Finish: baseTime.Add(3 * time.Hour)},
			},
			expected: baseTime.Add(3 * time.Hour),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			task := &Task{
				Name:     "Test Task",
				Segments: tc.segments,
			}

			result := task.GetLastActivity()
			if !result.Equal(tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestIsActive(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		segments []*Segment
		expected bool
	}{
		{
			name:     "no segments",
			segments: []*Segment{},
			expected: false,
		},
		{
			name: "open segment",
			segments: []*Segment{
				{Create: baseTime, Finish: time.Time{}},
			},
			expected: true,
		},
		{
			name: "closed segment",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			task := &Task{
				Name:     "Test Task",
				Segments: tc.segments,
			}

			result := task.IsActive()
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestGetCurrentSegmentDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		segments []*Segment
		expected bool // true if duration should be > 0
	}{
		{
			name:     "no segments",
			segments: []*Segment{},
			expected: false,
		},
		{
			name: "open segment",
			segments: []*Segment{
				{Create: time.Now().Add(-time.Hour), Finish: time.Time{}},
			},
			expected: true,
		},
		{
			name: "closed segment",
			segments: []*Segment{
				{Create: time.Now().Add(-2 * time.Hour), Finish: time.Now().Add(-time.Hour)},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			task := &Task{
				Name:     "Test Task",
				Segments: tc.segments,
			}

			result := task.GetCurrentSegmentDuration()
			if tc.expected && result == 0 {
				t.Error("Expected duration > 0, got 0")
			}
			if !tc.expected && result != 0 {
				t.Errorf("Expected duration 0, got %v", result)
			}
		})
	}
}

func TestGetLastSegment(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		segments []*Segment
		expected *Segment
	}{
		{
			name:     "no segments",
			segments: []*Segment{},
			expected: nil,
		},
		{
			name: "single segment",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour), Note: "first"},
			},
			expected: &Segment{Create: baseTime, Finish: baseTime.Add(time.Hour), Note: "first"},
		},
		{
			name: "multiple segments",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour), Note: "first"},
				{Create: baseTime.Add(2 * time.Hour), Finish: time.Time{}, Note: "last"},
			},
			expected: &Segment{Create: baseTime.Add(2 * time.Hour), Finish: time.Time{}, Note: "last"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			task := &Task{
				Name:     "Test Task",
				Segments: tc.segments,
			}

			result := task.GetLastSegment()
			if tc.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Error("Expected segment, got nil")
				} else if result.Note != tc.expected.Note {
					t.Errorf("Expected note %s, got %s", tc.expected.Note, result.Note)
				}
			}
		})
	}
}
