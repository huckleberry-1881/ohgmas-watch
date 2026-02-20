package main

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "zero duration",
			duration: 0,
			want:     "0m",
		},
		{
			name:     "one minute",
			duration: time.Minute,
			want:     "1m",
		},
		{
			name:     "30 minutes",
			duration: 30 * time.Minute,
			want:     "30m",
		},
		{
			name:     "59 minutes",
			duration: 59 * time.Minute,
			want:     "59m",
		},
		{
			name:     "one hour exactly",
			duration: time.Hour,
			want:     "1h00m",
		},
		{
			name:     "one hour 30 minutes",
			duration: time.Hour + 30*time.Minute,
			want:     "1h30m",
		},
		{
			name:     "two hours 5 minutes",
			duration: 2*time.Hour + 5*time.Minute,
			want:     "2h05m",
		},
		{
			name:     "10 hours 45 minutes",
			duration: 10*time.Hour + 45*time.Minute,
			want:     "10h45m",
		},
		{
			name:     "100 hours",
			duration: 100 * time.Hour,
			want:     "100h00m",
		},
		{
			name:     "seconds ignored",
			duration: time.Hour + 30*time.Minute + 45*time.Second,
			want:     "1h30m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestParseTimeFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		startFlag  string
		finishFlag string
		wantStart  bool
		wantFinish bool
		wantErr    bool
	}{
		{
			name:       "empty flags",
			startFlag:  "",
			finishFlag: "",
			wantStart:  false,
			wantFinish: false,
			wantErr:    false,
		},
		{
			name:       "valid start only",
			startFlag:  "2024-01-15T12:00:00Z",
			finishFlag: "",
			wantStart:  true,
			wantFinish: false,
			wantErr:    false,
		},
		{
			name:       "valid finish only",
			startFlag:  "",
			finishFlag: "2024-01-20T18:00:00Z",
			wantStart:  false,
			wantFinish: true,
			wantErr:    false,
		},
		{
			name:       "both valid",
			startFlag:  "2024-01-15T12:00:00Z",
			finishFlag: "2024-01-20T18:00:00Z",
			wantStart:  true,
			wantFinish: true,
			wantErr:    false,
		},
		{
			name:       "invalid start",
			startFlag:  "not-a-date",
			finishFlag: "",
			wantStart:  false,
			wantFinish: false,
			wantErr:    true,
		},
		{
			name:       "invalid finish",
			startFlag:  "",
			finishFlag: "invalid",
			wantStart:  false,
			wantFinish: false,
			wantErr:    true,
		},
		{
			name:       "valid start invalid finish",
			startFlag:  "2024-01-15T12:00:00Z",
			finishFlag: "bad-date",
			wantStart:  false,
			wantFinish: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			start, finish, err := parseTimeFlags(tt.startFlag, tt.finishFlag)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimeFlags() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr {
				return
			}

			if (start != nil) != tt.wantStart {
				t.Errorf("parseTimeFlags() start = %v, wantStart %v", start, tt.wantStart)
			}

			if (finish != nil) != tt.wantFinish {
				t.Errorf("parseTimeFlags() finish = %v, wantFinish %v", finish, tt.wantFinish)
			}
		})
	}
}

func TestParseTimeFlags_Values(t *testing.T) {
	t.Parallel()

	start, finish, err := parseTimeFlags("2024-01-15T12:30:00Z", "2024-01-20T18:45:00Z")
	if err != nil {
		t.Fatalf("parseTimeFlags() unexpected error: %v", err)
	}

	expectedStart := time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC)
	if !start.Equal(expectedStart) {
		t.Errorf("parseTimeFlags() start = %v, want %v", start, expectedStart)
	}

	expectedFinish := time.Date(2024, 1, 20, 18, 45, 0, 0, time.UTC)
	if !finish.Equal(expectedFinish) {
		t.Errorf("parseTimeFlags() finish = %v, want %v", finish, expectedFinish)
	}
}

func TestGetMondayOfWeek(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		when time.Time
		want time.Time
	}{
		{
			name: "monday returns same monday at midnight",
			when: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC), // Monday
			want: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "tuesday returns previous monday",
			when: time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC), // Tuesday
			want: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "wednesday returns previous monday",
			when: time.Date(2024, 1, 17, 12, 0, 0, 0, time.UTC), // Wednesday
			want: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "thursday returns previous monday",
			when: time.Date(2024, 1, 18, 9, 0, 0, 0, time.UTC), // Thursday
			want: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "friday returns previous monday",
			when: time.Date(2024, 1, 19, 17, 0, 0, 0, time.UTC), // Friday
			want: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "saturday returns previous monday",
			when: time.Date(2024, 1, 20, 8, 0, 0, 0, time.UTC), // Saturday
			want: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "sunday returns monday from 6 days ago",
			when: time.Date(2024, 1, 21, 11, 0, 0, 0, time.UTC), // Sunday
			want: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "midnight monday stays same",
			when: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			want: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "end of monday stays same day",
			when: time.Date(2024, 1, 15, 23, 59, 59, 0, time.UTC),
			want: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "crossing month boundary",
			when: time.Date(2024, 2, 1, 12, 0, 0, 0, time.UTC), // Thursday Feb 1
			want: time.Date(2024, 1, 29, 0, 0, 0, 0, time.UTC), // Monday Jan 29
		},
		{
			name: "crossing year boundary",
			when: time.Date(2024, 1, 3, 12, 0, 0, 0, time.UTC), // Wednesday Jan 3
			want: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),  // Monday Jan 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getMondayOfWeek(tt.when)
			if !got.Equal(tt.want) {
				t.Errorf("getMondayOfWeek(%v) = %v, want %v", tt.when, got, tt.want)
			}
		})
	}
}

func TestGetLastMonday(t *testing.T) {
	t.Parallel()

	// getLastMonday uses time.Now(), so we can only verify properties
	monday := getLastMonday()

	// Should be a Monday
	if monday.Weekday() != time.Monday {
		t.Errorf("getLastMonday() weekday = %v, want Monday", monday.Weekday())
	}

	// Should be at midnight
	if monday.Hour() != 0 || monday.Minute() != 0 || monday.Second() != 0 {
		t.Errorf("getLastMonday() time = %02d:%02d:%02d, want 00:00:00",
			monday.Hour(), monday.Minute(), monday.Second())
	}

	// Should not be in the future
	if monday.After(time.Now()) {
		t.Errorf("getLastMonday() = %v, should not be in the future", monday)
	}

	// Should be within the last 7 days
	weekAgo := time.Now().AddDate(0, 0, -7)
	if monday.Before(weekAgo) {
		t.Errorf("getLastMonday() = %v, should be within last 7 days", monday)
	}
}

func TestGetWeekStarts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		earliest time.Time
		latest   time.Time
		wantLen  int
	}{
		{
			name:     "same week",
			earliest: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), // Monday
			latest:   time.Date(2024, 1, 17, 15, 0, 0, 0, time.UTC), // Wednesday same week
			wantLen:  1,
		},
		{
			name:     "two consecutive weeks",
			earliest: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), // Monday week 1
			latest:   time.Date(2024, 1, 22, 15, 0, 0, 0, time.UTC), // Monday week 2
			wantLen:  2,
		},
		{
			name:     "three weeks",
			earliest: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC), // Monday week 1
			latest:   time.Date(2024, 1, 29, 15, 0, 0, 0, time.UTC), // Monday week 3
			wantLen:  3,
		},
		{
			name:     "partial weeks",
			earliest: time.Date(2024, 1, 17, 10, 0, 0, 0, time.UTC), // Wednesday week 1
			latest:   time.Date(2024, 1, 26, 15, 0, 0, 0, time.UTC), // Friday week 2
			wantLen:  2,
		},
		{
			name:     "same day",
			earliest: time.Date(2024, 1, 17, 10, 0, 0, 0, time.UTC),
			latest:   time.Date(2024, 1, 17, 15, 0, 0, 0, time.UTC),
			wantLen:  1,
		},
		{
			name:     "four weeks spanning month",
			earliest: time.Date(2024, 1, 22, 10, 0, 0, 0, time.UTC), // Monday Jan 22
			latest:   time.Date(2024, 2, 14, 15, 0, 0, 0, time.UTC), // Wednesday Feb 14
			wantLen:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getWeekStarts(tt.earliest, tt.latest)
			if len(got) != tt.wantLen {
				t.Errorf("getWeekStarts() returned %d weeks, want %d", len(got), tt.wantLen)
			}

			// Verify all returned times are Mondays at midnight
			for i, monday := range got {
				if monday.Weekday() != time.Monday {
					t.Errorf("getWeekStarts()[%d] weekday = %v, want Monday", i, monday.Weekday())
				}

				if monday.Hour() != 0 || monday.Minute() != 0 || monday.Second() != 0 {
					t.Errorf("getWeekStarts()[%d] time = %02d:%02d:%02d, want 00:00:00",
						i, monday.Hour(), monday.Minute(), monday.Second())
				}
			}

			// Verify weeks are in order and 7 days apart
			for i := 1; i < len(got); i++ {
				diff := got[i].Sub(got[i-1])
				if diff != 7*24*time.Hour {
					t.Errorf("getWeekStarts()[%d] - [%d] = %v, want 168h", i, i-1, diff)
				}
			}
		})
	}
}

func TestGetWeekStarts_Values(t *testing.T) {
	t.Parallel()

	earliest := time.Date(2024, 1, 17, 10, 0, 0, 0, time.UTC) // Wednesday Jan 17
	latest := time.Date(2024, 1, 30, 15, 0, 0, 0, time.UTC)   // Tuesday Jan 30

	weeks := getWeekStarts(earliest, latest)

	expected := []time.Time{
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), // Monday Jan 15
		time.Date(2024, 1, 22, 0, 0, 0, 0, time.UTC), // Monday Jan 22
		time.Date(2024, 1, 29, 0, 0, 0, 0, time.UTC), // Monday Jan 29
	}

	if len(weeks) != len(expected) {
		t.Fatalf("getWeekStarts() returned %d weeks, want %d", len(weeks), len(expected))
	}

	for i, want := range expected {
		if !weeks[i].Equal(want) {
			t.Errorf("getWeekStarts()[%d] = %v, want %v", i, weeks[i], want)
		}
	}
}
