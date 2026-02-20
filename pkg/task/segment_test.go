package task //nolint:testpackage // tests unexported functions

import (
	"testing"
	"time"
)

func TestIsSegmentInRange(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		segment *Segment
		start   *time.Time
		finish  *time.Time
		want    bool
	}{
		{
			name: "open segment is never in range",
			segment: &Segment{
				Create: baseTime,
				Finish: time.Time{},
			},
			start:  nil,
			finish: nil,
			want:   false,
		},
		{
			name: "closed segment with nil bounds",
			segment: &Segment{
				Create: baseTime,
				Finish: baseTime.Add(time.Hour),
			},
			start:  nil,
			finish: nil,
			want:   true,
		},
		{
			name: "segment finishes before start",
			segment: &Segment{
				Create: baseTime,
				Finish: baseTime.Add(time.Hour),
			},
			start:  timePtr(baseTime.Add(2 * time.Hour)),
			finish: nil,
			want:   false,
		},
		{
			name: "segment finishes exactly at start (exclusive, not included)",
			segment: &Segment{
				Create: baseTime,
				Finish: baseTime.Add(time.Hour),
			},
			start:  timePtr(baseTime.Add(time.Hour)),
			finish: nil,
			want:   false,
		},
		{
			name: "segment finishes after start",
			segment: &Segment{
				Create: baseTime,
				Finish: baseTime.Add(2 * time.Hour),
			},
			start:  timePtr(baseTime.Add(time.Hour)),
			finish: nil,
			want:   true,
		},
		{
			name: "segment finishes after finish (exclusive)",
			segment: &Segment{
				Create: baseTime,
				Finish: baseTime.Add(2 * time.Hour),
			},
			start:  nil,
			finish: timePtr(baseTime.Add(time.Hour)),
			want:   false,
		},
		{
			name: "segment finishes exactly at finish (inclusive)",
			segment: &Segment{
				Create: baseTime,
				Finish: baseTime.Add(time.Hour),
			},
			start:  nil,
			finish: timePtr(baseTime.Add(time.Hour)),
			want:   true,
		},
		{
			name: "segment finishes before finish",
			segment: &Segment{
				Create: baseTime,
				Finish: baseTime.Add(time.Hour),
			},
			start:  nil,
			finish: timePtr(baseTime.Add(2 * time.Hour)),
			want:   true,
		},
		{
			name: "segment within both bounds",
			segment: &Segment{
				Create: baseTime,
				Finish: baseTime.Add(time.Hour),
			},
			start:  timePtr(baseTime.Add(-time.Hour)),
			finish: timePtr(baseTime.Add(2 * time.Hour)),
			want:   true,
		},
		{
			name: "segment outside both bounds (before)",
			segment: &Segment{
				Create: baseTime,
				Finish: baseTime.Add(time.Hour),
			},
			start:  timePtr(baseTime.Add(2 * time.Hour)),
			finish: timePtr(baseTime.Add(3 * time.Hour)),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isSegmentInRange(tt.segment, tt.start, tt.finish)
			if got != tt.want {
				t.Errorf("isSegmentInRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_HasSegmentsInRange(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		segments []*Segment
		start    *time.Time
		finish   *time.Time
		want     bool
	}{
		{
			name:     "no segments",
			segments: []*Segment{},
			start:    nil,
			finish:   nil,
			want:     false,
		},
		{
			name: "has matching segment",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
			},
			start:  timePtr(baseTime.Add(-time.Hour)),
			finish: timePtr(baseTime.Add(2 * time.Hour)),
			want:   true,
		},
		{
			name: "all segments outside range",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
			},
			start:  timePtr(baseTime.Add(2 * time.Hour)),
			finish: timePtr(baseTime.Add(3 * time.Hour)),
			want:   false,
		},
		{
			name: "only open segments",
			segments: []*Segment{
				{Create: baseTime, Finish: time.Time{}},
			},
			start:  nil,
			finish: nil,
			want:   false,
		},
		{
			name: "mixed open and closed, one matching",
			segments: []*Segment{
				{Create: baseTime, Finish: time.Time{}}, // Open
				{Create: baseTime.Add(time.Hour), Finish: baseTime.Add(2 * time.Hour)}, // Closed in range
			},
			start:  timePtr(baseTime),
			finish: timePtr(baseTime.Add(3 * time.Hour)),
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task := &Task{
				Name:     "Test Task",
				Segments: tt.segments,
			}

			got := task.HasSegmentsInRange(tt.start, tt.finish)
			if got != tt.want {
				t.Errorf("HasSegmentsInRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_GetFilteredClosedSegmentsDuration(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		segments []*Segment
		start    *time.Time
		finish   *time.Time
		want     time.Duration
	}{
		{
			name:     "no segments",
			segments: []*Segment{},
			start:    nil,
			finish:   nil,
			want:     0,
		},
		{
			name: "single segment in range",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
			},
			start:  nil,
			finish: nil,
			want:   time.Hour,
		},
		{
			name: "segment outside range not counted",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
			},
			start:  timePtr(baseTime.Add(2 * time.Hour)),
			finish: timePtr(baseTime.Add(3 * time.Hour)),
			want:   0,
		},
		{
			name: "multiple segments some in range",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},                    // In range: 1h
				{Create: baseTime.Add(2 * time.Hour), Finish: baseTime.Add(3 * time.Hour)}, // In range: 1h
				{Create: baseTime.Add(5 * time.Hour), Finish: baseTime.Add(6 * time.Hour)}, // Out of range
			},
			start:  timePtr(baseTime.Add(-time.Hour)),
			finish: timePtr(baseTime.Add(4 * time.Hour)),
			want:   2 * time.Hour,
		},
		{
			name: "open segments ignored",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)}, // Closed: 1h
				{Create: baseTime.Add(2 * time.Hour), Finish: time.Time{}}, // Open: not counted
			},
			start:  nil,
			finish: nil,
			want:   time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task := &Task{
				Name:     "Test Task",
				Segments: tt.segments,
			}

			got := task.GetFilteredClosedSegmentsDuration(tt.start, tt.finish)
			if got != tt.want {
				t.Errorf("GetFilteredClosedSegmentsDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_GetLastActivity(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		segments []*Segment
		want     time.Time
	}{
		{
			name:     "no segments returns zero time",
			segments: []*Segment{},
			want:     time.Time{},
		},
		{
			name: "closed segment returns finish time",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
			},
			want: baseTime.Add(time.Hour),
		},
		{
			name: "open segment returns create time",
			segments: []*Segment{
				{Create: baseTime, Finish: time.Time{}},
			},
			want: baseTime,
		},
		{
			name: "multiple segments returns last segment time",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
				{Create: baseTime.Add(2 * time.Hour), Finish: baseTime.Add(3 * time.Hour)},
			},
			want: baseTime.Add(3 * time.Hour),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task := &Task{
				Name:     "Test Task",
				Segments: tt.segments,
			}

			got := task.GetLastActivity()
			if !got.Equal(tt.want) {
				t.Errorf("GetLastActivity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_IsActive(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		segments []*Segment
		want     bool
	}{
		{
			name:     "no segments",
			segments: []*Segment{},
			want:     false,
		},
		{
			name: "all closed",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
			},
			want: false,
		},
		{
			name: "has open segment",
			segments: []*Segment{
				{Create: baseTime, Finish: time.Time{}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task := &Task{
				Name:     "Test Task",
				Segments: tt.segments,
			}

			got := task.IsActive()
			if got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_GetCurrentSegmentDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		segments      []*Segment
		expectNonZero bool
	}{
		{
			name:          "no segments",
			segments:      []*Segment{},
			expectNonZero: false,
		},
		{
			name: "last segment closed",
			segments: []*Segment{
				{Create: time.Now().Add(-time.Hour), Finish: time.Now().Add(-30 * time.Minute)},
			},
			expectNonZero: false,
		},
		{
			name: "last segment open",
			segments: []*Segment{
				{Create: time.Now().Add(-time.Hour), Finish: time.Time{}},
			},
			expectNonZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task := &Task{
				Name:     "Test Task",
				Segments: tt.segments,
			}

			got := task.GetCurrentSegmentDuration()
			if tt.expectNonZero && got == 0 {
				t.Error("GetCurrentSegmentDuration() = 0, want non-zero")
			}
			if !tt.expectNonZero && got != 0 {
				t.Errorf("GetCurrentSegmentDuration() = %v, want 0", got)
			}
		})
	}
}

func TestTask_GetLastSegment(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		segments   []*Segment
		wantNil    bool
		wantCreate time.Time
	}{
		{
			name:     "no segments returns nil",
			segments: []*Segment{},
			wantNil:  true,
		},
		{
			name: "single segment",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
			},
			wantNil:    false,
			wantCreate: baseTime,
		},
		{
			name: "multiple segments returns last",
			segments: []*Segment{
				{Create: baseTime, Finish: baseTime.Add(time.Hour)},
				{Create: baseTime.Add(2 * time.Hour), Finish: baseTime.Add(3 * time.Hour)},
			},
			wantNil:    false,
			wantCreate: baseTime.Add(2 * time.Hour),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task := &Task{
				Name:     "Test Task",
				Segments: tt.segments,
			}

			got := task.GetLastSegment()
			if tt.wantNil {
				if got != nil {
					t.Error("GetLastSegment() should return nil")
				}
			} else {
				if got == nil {
					t.Error("GetLastSegment() should not return nil")
				} else if !got.Create.Equal(tt.wantCreate) {
					t.Errorf("GetLastSegment().Create = %v, want %v", got.Create, tt.wantCreate)
				}
			}
		})
	}
}

func TestTask_GetThisWeekDuration(t *testing.T) {
	t.Parallel()

	// Use a fixed Monday as week start
	weekStart := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC) // Monday

	tests := []struct {
		name      string
		segments  []*Segment
		weekStart time.Time
		want      time.Duration
	}{
		{
			name:      "no segments",
			segments:  []*Segment{},
			weekStart: weekStart,
			want:      0,
		},
		{
			name: "segment finished after week start",
			segments: []*Segment{
				{Create: weekStart.Add(time.Hour), Finish: weekStart.Add(2 * time.Hour)},
			},
			weekStart: weekStart,
			want:      time.Hour,
		},
		{
			name: "segment finished before week start",
			segments: []*Segment{
				{Create: weekStart.Add(-2 * time.Hour), Finish: weekStart.Add(-time.Hour)},
			},
			weekStart: weekStart,
			want:      0,
		},
		{
			name: "segment finished exactly at week start (not included)",
			segments: []*Segment{
				{Create: weekStart.Add(-time.Hour), Finish: weekStart},
			},
			weekStart: weekStart,
			want:      0,
		},
		{
			name: "open segment not included",
			segments: []*Segment{
				{Create: weekStart.Add(time.Hour), Finish: time.Time{}},
			},
			weekStart: weekStart,
			want:      0,
		},
		{
			name: "multiple segments mixed",
			segments: []*Segment{
				{Create: weekStart.Add(-2 * time.Hour), Finish: weekStart.Add(-time.Hour)}, // Before: not counted
				{Create: weekStart.Add(time.Hour), Finish: weekStart.Add(3 * time.Hour)},   // After: 2h
				{Create: weekStart.Add(4 * time.Hour), Finish: weekStart.Add(5 * time.Hour)}, // After: 1h
			},
			weekStart: weekStart,
			want:      3 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task := &Task{
				Name:     "Test Task",
				Segments: tt.segments,
			}

			got := task.GetThisWeekDuration(tt.weekStart)
			if got != tt.want {
				t.Errorf("GetThisWeekDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to create a time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
