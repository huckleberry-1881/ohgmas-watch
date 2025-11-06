package task

import (
	"sort"
	"strings"
	"time"
)

// getTagsetKey creates a sorted, comma-separated key from a slice of tags.
func getTagsetKey(tags []string) string {
	if len(tags) == 0 {
		return "(no tags)"
	}

	tagset := make([]string, len(tags))
	copy(tagset, tags)
	sort.Strings(tagset)

	return strings.Join(tagset, ", ")
}

// sortTagsetSummaries converts a tagset map to a slice and sorts by duration (descending).
func sortTagsetSummaries(tagsetMap map[string]*TagsetSummary) []TagsetSummary {
	summaries := make([]TagsetSummary, 0, len(tagsetMap))
	for _, summary := range tagsetMap {
		summaries = append(summaries, *summary)
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Duration > summaries[j].Duration
	})

	return summaries
}

// TagsetSummary represents a summary of tasks grouped by tagset.
type TagsetSummary struct {
	Tagset   string
	Tasks    []*Task
	Duration time.Duration
}

// WeeklySummary represents a summary for a specific week.
type WeeklySummary struct {
	WeekStart time.Time
	Tagsets   []TagsetSummary
}

// GetSummaryByTagset generates a summary of tasks grouped by tagset.
func (w *Watch) GetSummaryByTagset(start, finish *time.Time) []TagsetSummary {
	// Group tasks by tagset (combination of tags)
	tagsetMap := make(map[string]*TagsetSummary)

	for _, currentTask := range w.Tasks {
		// Skip tasks that have no segments in the specified time range
		if (start != nil || finish != nil) && !currentTask.HasSegmentsInRange(start, finish) {
			continue
		}

		tagsetKey := getTagsetKey(currentTask.Tags)

		if tagsetMap[tagsetKey] == nil {
			tagsetMap[tagsetKey] = &TagsetSummary{
				Tagset:   tagsetKey,
				Tasks:    []*Task{},
				Duration: 0,
			}
		}

		tagsetMap[tagsetKey].Tasks = append(tagsetMap[tagsetKey].Tasks, currentTask)
		tagsetMap[tagsetKey].Duration += currentTask.GetFilteredClosedSegmentsDuration(start, finish)
	}

	return sortTagsetSummaries(tagsetMap)
}

// GetTasksSortedByActivity returns tasks sorted by last activity (most recent first).
func (w *Watch) GetTasksSortedByActivity() []*Task {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return sortTasksByActivity(w.Tasks)
}

// GetTaskIndex returns the original index of a task in the Watch.Tasks slice.
func (w *Watch) GetTaskIndex(task *Task) int {
	for i, t := range w.Tasks {
		if t == task {
			return i
		}
	}

	return -1
}

// GetWeeklySummaryByTagset generates weekly summaries grouped by tagset.
func (w *Watch) GetWeeklySummaryByTagset(weekStarts []time.Time) []WeeklySummary {
	var weeklySummaries []WeeklySummary

	for _, weekStart := range weekStarts {
		// Calculate the end of the week (start of next week)
		weekEnd := weekStart.AddDate(0, 0, 7)

		// Get summary for this week
		tagsetSummaries := w.GetSummaryByTagset(&weekStart, &weekEnd)

		// Only include weeks that have data
		if len(tagsetSummaries) > 0 {
			weeklySummaries = append(weeklySummaries, WeeklySummary{
				WeekStart: weekStart,
				Tagsets:   tagsetSummaries,
			})
		}
	}

	return weeklySummaries
}

// GetEarliestAndLatestSegmentTimes returns the earliest and latest segment times across all tasks.
// If no segments exist, returns zero times.
func (w *Watch) GetEarliestAndLatestSegmentTimes() (time.Time, time.Time) {
	var earliest, latest time.Time

	for _, task := range w.Tasks {
		task.mu.RLock()
		for _, segment := range task.Segments {
			// Check earliest
			if earliest.IsZero() || segment.Create.Before(earliest) {
				earliest = segment.Create
			}

			// Check latest (use finish time if closed, otherwise create time)
			segmentEnd := segment.Finish
			if segmentEnd.IsZero() {
				segmentEnd = segment.Create
			}

			if latest.IsZero() || segmentEnd.After(latest) {
				latest = segmentEnd
			}
		}
		task.mu.RUnlock()
	}

	return earliest, latest
}

// GetWeeklySummaryByTagsetWithTasks generates weekly summaries grouped by tagset with individual task breakdowns.
func (w *Watch) GetWeeklySummaryByTagsetWithTasks(weekStarts []time.Time) []WeeklySummary {
	var weeklySummaries []WeeklySummary

	for _, weekStart := range weekStarts {
		// Calculate the end of the week (start of next week)
		weekEnd := weekStart.AddDate(0, 0, 7)

		// Get summary for this week with tasks
		tagsetMap := make(map[string]*TagsetSummary)

		for _, currentTask := range w.Tasks {
			// Check if task has segments in this week
			if !currentTask.HasSegmentsInRange(&weekStart, &weekEnd) {
				continue
			}

			tagsetKey := getTagsetKey(currentTask.Tags)

			if tagsetMap[tagsetKey] == nil {
				tagsetMap[tagsetKey] = &TagsetSummary{
					Tagset:   tagsetKey,
					Tasks:    []*Task{},
					Duration: 0,
				}
			}

			taskDuration := currentTask.GetFilteredClosedSegmentsDuration(&weekStart, &weekEnd)
			tagsetMap[tagsetKey].Tasks = append(tagsetMap[tagsetKey].Tasks, currentTask)
			tagsetMap[tagsetKey].Duration += taskDuration
		}

		tagsetSummaries := sortTagsetSummaries(tagsetMap)

		// Only include weeks that have data
		if len(tagsetSummaries) > 0 {
			weeklySummaries = append(weeklySummaries, WeeklySummary{
				WeekStart: weekStart,
				Tagsets:   tagsetSummaries,
			})
		}
	}

	return weeklySummaries
}
