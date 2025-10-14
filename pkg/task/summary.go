package task

import (
	"sort"
	"strings"
	"time"
)

// TagsetSummary represents a summary of tasks grouped by tagset.
type TagsetSummary struct {
	Tagset   string
	Tasks    []*Task
	Duration time.Duration
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

		// Create a sorted tagset key
		tagset := make([]string, len(currentTask.Tags))
		copy(tagset, currentTask.Tags)
		sort.Strings(tagset)

		tagsetKey := strings.Join(tagset, ", ")
		if tagsetKey == "" {
			tagsetKey = "(no tags)"
		}

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

	// Convert map to slice and sort by total duration (descending)
	summaries := make([]TagsetSummary, 0, len(tagsetMap))
	for _, summary := range tagsetMap {
		summaries = append(summaries, *summary)
	}

	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Duration > summaries[j].Duration
	})

	return summaries
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
