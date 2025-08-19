// package main
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/huckleberry-1881/ohgmas-watch/pkg/task"
)

// Function to generate summary by tagset.
func generateSummary(includeTasks bool, start, finish *time.Time) error {
	// Load tasks
	watch := &task.Watch{
		Tasks: []*task.Task{},
	}

	err := watch.LoadTasks()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	// Get summary by tagset
	summaries := watch.GetSummaryByTagset(start, finish)

	// Print summary
	for _, summary := range summaries {
		taskCount := len(summary.Tasks)
		duration := summary.Duration

		// Format duration for display
		var durationStr string
		if duration == 0 {
			durationStr = "0m"
		} else {
			durationStr = formatDuration(duration)
		}

		// Use singular/plural form correctly
		taskWord := "tasks"
		if taskCount == 1 {
			taskWord = "task"
		}

		_, _ = fmt.Fprintf(os.Stdout, "%s totaling %d %s and %s\n", summary.Tagset, taskCount, taskWord, durationStr)

		// If includeTasks is true, list individual tasks
		if includeTasks {
			for _, taskItem := range summary.Tasks {
				taskDuration := taskItem.GetFilteredClosedSegmentsDuration(start, finish)

				var taskDurationStr string
				if taskDuration == 0 {
					taskDurationStr = "0m"
				} else {
					taskDurationStr = formatDuration(taskDuration)
				}

				_, _ = fmt.Fprintf(os.Stdout, "  Task: %s %s\n", taskItem.Name, taskDurationStr)
			}
		}
	}

	return nil
}

func main() {
	// Define command line flags
	summaryFlag := flag.Bool("summary", false, "Generate a summary of work completed by tagset")
	tasksFlag := flag.Bool("tasks", false, "Include individual task details in summary (requires --summary)")
	startFlag := flag.String("start", "",
		"Filter segments to only include those closed after this datetime (RFC3339 format: 2006-01-02T15:04:05Z)")
	finishFlag := flag.String("finish", "",
		"Filter segments to only include those closed before this datetime (RFC3339 format: 2006-01-02T15:04:05Z)")

	// Parse command line flags
	flag.Parse()

	// Check if summary flag was provided
	if *summaryFlag {
		start, finish, err := parseTimeFlags(*startFlag, *finishFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		err = generateSummary(*tasksFlag, start, finish)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		return
	}

	// Check if tasks flag was provided without summary
	if *tasksFlag {
		fmt.Fprintf(os.Stderr, "Error: --tasks flag requires --summary flag\n")
		os.Exit(1)
	}

	// Start TUI application
	app := tview.NewApplication()

	// Task storage - load from file
	watch := &task.Watch{
		Tasks: []*task.Task{},
	}

	err := watch.LoadTasks()
	if err != nil {
		// If we can't load tasks, start with empty watch
		watch.Tasks = []*task.Task{}
	}

	// Create UI components
	table := tview.NewTable()
	table.SetBorder(true).SetTitle("Tasks")
	table.SetSelectable(true, false)
	table.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorBlack))
	table.SetSeparator(tview.Borders.Vertical)

	// Set up table headers
	table.SetCell(0, 0, tview.NewTableCell("Status").
		SetTextColor(tcell.ColorYellow).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	table.SetCell(0, 1, tview.NewTableCell("Task Name").
		SetTextColor(tcell.ColorYellow).
		SetSelectable(false).
		SetAlign(tview.AlignLeft))
	table.SetCell(0, 2, tview.NewTableCell("Category").
		SetTextColor(tcell.ColorYellow).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	table.SetCell(0, 3, tview.NewTableCell("Tags").
		SetTextColor(tcell.ColorYellow).
		SetSelectable(false).
		SetAlign(tview.AlignLeft))
	table.SetCell(0, 4, tview.NewTableCell("Last Activity").
		SetTextColor(tcell.ColorYellow).
		SetSelectable(false).
		SetAlign(tview.AlignCenter))
	table.SetCell(0, 5, tview.NewTableCell("This Week").
		SetTextColor(tcell.ColorYellow).
		SetSelectable(false).
		SetAlign(tview.AlignRight))
	table.SetCell(0, 6, tview.NewTableCell("Duration").
		SetTextColor(tcell.ColorYellow).
		SetSelectable(false).
		SetAlign(tview.AlignRight))

	// Description view for selected task
	descriptionView := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true)
	descriptionView.SetBorder(true).SetTitle("Description")

	// Command summary
	commandText := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[yellow]Commands:[white] ↑/↓ Navigate | [green]Enter[white] Segment Details | " +
			"[green]t[white] New Task | [green]s[white] New Segment | [green]n[white] New Segment w/ Note | " +
			"[green]e[white] End Segment | [blue]c[white] Completed | [blue]w[white] Work | [blue]b[white] Backlog | " +
			"[purple]f[white] Filter | [red]Ctrl+C[white] Exit")
	commandText.SetBorder(true).SetTitle("Commands")

	// Map from table row to original task index in watch.Tasks
	var tableRowToTaskIndex []int

	// Current category filter (empty means show all)
	var currentCategoryFilter string
	categoryFilters := []string{"", "completed", "work", "backlog"} // "" = all
	filterIndex := 0

	// Helper function to get original task index from table row
	getTaskIndex := func(tableRow int) int {
		dataRow := tableRow - 1 // -1 because row 0 is headers
		if dataRow >= 0 && dataRow < len(tableRowToTaskIndex) {
			return tableRowToTaskIndex[dataRow]
		}

		return -1 // Invalid row
	}

	// Helper function to save tasks and refresh the table
	saveAndRefresh := func() {
		err = watch.SaveTasks()
		if err != nil {
			panic(err)
		}

		// Clear existing rows (keep header row)
		rowCount := table.GetRowCount()
		for r := rowCount - 1; r > 0; r-- {
			table.RemoveRow(r)
		}

		// Get tasks sorted by last activity (with optional category filter)
		sortedTasks := watch.GetTasksSortedByActivityWithFilter(currentCategoryFilter)

		// Update table title to show current filter
		filterTitle := "Tasks"
		if currentCategoryFilter != "" {
			filterTitle = fmt.Sprintf("Tasks (%s)", currentCategoryFilter)
		}
		table.SetTitle(filterTitle)

		// Update the row-to-task mapping
		tableRowToTaskIndex = make([]int, len(sortedTasks))
		for i, t := range sortedTasks {
			tableRowToTaskIndex[i] = watch.GetTaskIndex(t)
		}

		// Add task rows using sorted order
		for i, task := range sortedTasks {
			row := i + 1 // +1 because row 0 is headers

			// Status column
			statusCell := tview.NewTableCell("").SetAlign(tview.AlignCenter)
			if task.IsActive() {
				statusCell.SetText("▶").SetTextColor(tcell.ColorRed)
			} else {
				statusCell.SetText("●").SetTextColor(tcell.ColorGray)
			}

			// Task Name column
			nameCell := tview.NewTableCell(task.Name).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft)

			// Category column
			category := task.GetCategory()
			if category == "" {
				category = "work" // Default for existing tasks
			}
			var categoryColor tcell.Color
			switch category {
			case "completed":
				categoryColor = tcell.ColorGreen
			case "work":
				categoryColor = tcell.ColorYellow
			case "backlog":
				categoryColor = tcell.ColorGray
			default:
				categoryColor = tcell.ColorWhite
			}
			categoryCell := tview.NewTableCell(category).
				SetTextColor(categoryColor).
				SetAlign(tview.AlignCenter)

			// Tags column
			tagsText := ""
			if len(task.Tags) > 0 {
				tagsText = fmt.Sprintf("(%s)", strings.Join(task.Tags, ", "))
			}
			tagsCell := tview.NewTableCell(tagsText).
				SetTextColor(tcell.ColorBlue).
				SetAlign(tview.AlignLeft)

			// Last Activity column
			lastActivityText := "-"
			lastActivityColor := tcell.ColorGray
			lastActivity := task.GetLastActivity()
			if !lastActivity.IsZero() {
				lastActivityText = lastActivity.Format("2006-01-02")
				if task.IsActive() {
					lastActivityColor = tcell.ColorGreen // Green for active
				} else {
					lastActivityColor = tcell.ColorWhite
				}
			}
			lastActivityCell := tview.NewTableCell(lastActivityText).
				SetTextColor(lastActivityColor).
				SetAlign(tview.AlignCenter)

			// This Week column
			weekStart := getLastMonday()
			thisWeekDuration := task.GetThisWeekDuration(weekStart)
			thisWeekText := "0m"
			if thisWeekDuration > 0 {
				thisWeekText = formatDuration(thisWeekDuration)
			}
			thisWeekCell := tview.NewTableCell(thisWeekText).
				SetTextColor(tcell.ColorLightBlue).
				SetAlign(tview.AlignRight)

			// Duration column
			duration := task.GetClosedSegmentsDuration()
			durationText := "0m"
			if duration > 0 {
				durationText = formatDuration(duration)
			}
			durationCell := tview.NewTableCell(durationText).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignRight)

			// Set cells in the table
			table.SetCell(row, 0, statusCell)
			table.SetCell(row, 1, nameCell)
			table.SetCell(row, 2, categoryCell)
			table.SetCell(row, 3, tagsCell)
			table.SetCell(row, 4, lastActivityCell)
			table.SetCell(row, 5, thisWeekCell)
			table.SetCell(row, 6, durationCell)
		}

		// If we have filtered tasks, select the first data row (row 1)
		if len(sortedTasks) > 0 {
			table.Select(1, 0)
		}
	}

	// Helper function to update description view for current selection
	updateDescriptionView := func() {
		row, _ := table.GetSelection()
		currentIndex := getTaskIndex(row)
		if currentIndex >= 0 && currentIndex < len(watch.Tasks) {
			task := watch.Tasks[currentIndex]

			// Build the description content with segment info
			var content strings.Builder

			// Add last segment information if any segments exist
			lastSegment := task.GetLastSegment()
			if lastSegment != nil {
				if lastSegment.Finish.IsZero() {
					// Open segment - show start date and current duration
					content.WriteString(fmt.Sprintf("[yellow]Current Segment:[white] Started %s\n",
						lastSegment.Create.Format("2006-01-02 15:04:05")))

					currentDuration := task.GetCurrentSegmentDuration()
					content.WriteString(fmt.Sprintf("[yellow]Duration:[white] %s (ongoing)\n\n",
						formatDuration(currentDuration)))
				} else {
					// Closed segment - show end date and duration
					content.WriteString(fmt.Sprintf("[green]Last Segment:[white] Ended %s\n",
						lastSegment.Finish.Format("2006-01-02 15:04:05")))

					segmentDuration := lastSegment.Finish.Sub(lastSegment.Create)
					content.WriteString(fmt.Sprintf("[green]Duration:[white] %s\n\n",
						formatDuration(segmentDuration)))
				}
			}

			// Add the task description
			content.WriteString(task.Description)

			descriptionView.SetText(content.String())
		} else {
			descriptionView.SetText("")
		}
	}

	// Update description when selection changes
	table.SetSelectionChangedFunc(func(row, _ int) {
		updateDescriptionView()
	})

	// Create main layout
	mainLayout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(table, 0, 2, true).
		AddItem(descriptionView, 0, 1, false).
		AddItem(commandText, 3, 0, false)

	// Function to show new task form
	showNewTaskForm := func() {
		form := tview.NewForm()
		form.SetBorder(true).SetTitle("New Task")
		form.SetLabelColor(tcell.ColorWhite)
		form.SetFieldBackgroundColor(tcell.ColorGray)
		form.SetFieldTextColor(tcell.ColorGreen)
		form.SetButtonTextColor(tcell.ColorWhite)

		var name, description, tags string

		form.AddInputField("Name:", "", 50, nil, func(text string) {
			name = text
		})
		form.AddTextArea("Description:", "", 50, 5, 500, func(text string) {
			description = text
		})
		form.AddInputField("Tags (comma-separated):", "", 50, nil, func(text string) {
			tags = text
		})

		form.AddButton("Create", func() {
			if name != "" {
				tagList := []string{}

				if tags != "" {
					for tag := range strings.SplitSeq(tags, ",") {
						tagList = append(tagList, strings.TrimSpace(tag))
					}
				}

				watch.AddTask(name, description, tagList, "work")

				saveAndRefresh()
				app.SetRoot(mainLayout, true)
			}
		})

		form.AddButton("Cancel", func() {
			app.SetRoot(mainLayout, true)
		})

		// Center the form
		centeredForm := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(form, 0, 2, true).
				AddItem(nil, 0, 1, false), 0, 2, true).
			AddItem(nil, 0, 1, false)

		app.SetRoot(centeredForm, true)
	}

	// Function to show new segment form with note
	showNewSegmentWithNoteForm := func() {
		row, _ := table.GetSelection()
		currentIndex := getTaskIndex(row)
		if currentIndex < 0 || currentIndex >= len(watch.Tasks) {
			return // No task selected
		}

		form := tview.NewForm()
		form.SetBorder(true).SetTitle("New Segment")
		form.SetLabelColor(tcell.ColorWhite)
		form.SetFieldBackgroundColor(tcell.ColorGray)
		form.SetFieldTextColor(tcell.ColorGreen)
		form.SetButtonTextColor(tcell.ColorWhite)

		var note string

		form.AddTextArea("Note:", "", 50, 3, 300, func(text string) {
			note = text
		})

		form.AddButton("Create", func() {
			// Check if there's already an open segment
			for _, segment := range watch.Tasks[currentIndex].Segments {
				if segment.Finish.IsZero() {
					app.SetRoot(mainLayout, true) // Already has an open segment, just return to main

					return
				}
			}

			watch.Tasks[currentIndex].AddSegment(note)

			saveAndRefresh()
			app.SetRoot(mainLayout, true)
		})

		form.AddButton("Cancel", func() {
			app.SetRoot(mainLayout, true)
		})

		// Center the form
		centeredForm := tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(form, 0, 2, true).
				AddItem(nil, 0, 1, false), 0, 2, true).
			AddItem(nil, 0, 1, false)

		app.SetRoot(centeredForm, true)
	}

	// Function to create new segment without note
	createSegmentWithoutNote := func() {
		row, _ := table.GetSelection()
		currentIndex := getTaskIndex(row)
		if currentIndex < 0 || currentIndex >= len(watch.Tasks) {
			return // No task selected
		}

		// Check if there's already an open segment
		for _, segment := range watch.Tasks[currentIndex].Segments {
			if segment.Finish.IsZero() {
				return // Already has an open segment, don't create another
			}
		}

		watch.Tasks[currentIndex].AddSegment("")

		saveAndRefresh()
	}

	// Function to end an open segment
	endSegment := func() {
		row, _ := table.GetSelection()
		currentIndex := getTaskIndex(row)
		if currentIndex < 0 || currentIndex >= len(watch.Tasks) {
			return // No task selected
		}

		watch.Tasks[currentIndex].CloseSegment()
		saveAndRefresh()
	}

	// Function to change task category
	changeTaskCategory := func(category string) {
		row, _ := table.GetSelection()
		currentIndex := getTaskIndex(row)
		if currentIndex < 0 || currentIndex >= len(watch.Tasks) {
			return // No task selected
		}

		watch.Tasks[currentIndex].SetCategory(category)
		saveAndRefresh()
	}

	// Function to cycle through category filters
	cycleCategoryFilter := func() {
		filterIndex = (filterIndex + 1) % len(categoryFilters)
		currentCategoryFilter = categoryFilters[filterIndex]
		saveAndRefresh()
	}

	// Function to show segment details for the selected task
	showSegmentDetails := func() {
		row, _ := table.GetSelection()
		currentIndex := getTaskIndex(row)
		if currentIndex < 0 || currentIndex >= len(watch.Tasks) {
			return // No task selected
		}

		selectedTask := watch.Tasks[currentIndex]

		// Create segment details text view
		segmentView := tview.NewTextView().
			SetDynamicColors(true).
			SetWordWrap(true).
			SetScrollable(true)
		segmentView.SetBorder(true).SetTitle("Segments for: " + selectedTask.Name)

		// Build segment information
		var content strings.Builder
		if len(selectedTask.Segments) == 0 {
			content.WriteString("[gray]No segments found for this task.[-]\n")
		} else {
			for i, segment := range selectedTask.Segments {
				content.WriteString(fmt.Sprintf("[white]Segment %d:[-]\n", i+1))
				content.WriteString(fmt.Sprintf("  [green]Created:[-] %s\n", segment.Create.Format("2006-01-02 15:04:05")))

				if segment.Finish.IsZero() {
					content.WriteString("  [red]Status:[-] Open\n")

					duration := time.Since(segment.Create)

					content.WriteString(fmt.Sprintf("  [yellow]Duration:[-] %s (ongoing)\n", formatDuration(duration)))
				} else {
					content.WriteString(fmt.Sprintf("  [green]Finished:[-] %s\n", segment.Finish.Format("2006-01-02 15:04:05")))

					duration := segment.Finish.Sub(segment.Create)

					content.WriteString(fmt.Sprintf("  [yellow]Duration:[-] %s\n", formatDuration(duration)))
				}

				if segment.Note != "" {
					content.WriteString(fmt.Sprintf("  [cyan]Note:[-] %s\n", segment.Note))
				} else {
					content.WriteString("  [gray]Note: (none)[-]\n")
				}

				content.WriteString("\n")
			}
		}

		segmentView.SetText(content.String())

		// Create back button
		backButton := tview.NewButton("Back to Tasks").SetSelectedFunc(func() {
			app.SetRoot(mainLayout, true)
		})
		backButton.SetBackgroundColor(tcell.ColorGray)
		backButton.SetLabelColor(tcell.ColorWhite)

		// Create layout for segment details
		segmentLayout := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(segmentView, 0, 1, true).
			AddItem(backButton, 1, 0, false)

		// Set up key bindings for the segment view
		segmentView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEscape {
				app.SetRoot(mainLayout, true)

				return nil
			}

			return event
		})

		app.SetRoot(segmentLayout, true)
	}

	// Set up key bindings
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Rune() == 't':
			showNewTaskForm()

			return nil
		case event.Rune() == 's':
			createSegmentWithoutNote()

			return nil
		case event.Rune() == 'n':
			showNewSegmentWithNoteForm()

			return nil
		case event.Rune() == 'e':
			endSegment()

			return nil
		case event.Rune() == 'c':
			changeTaskCategory("completed")

			return nil
		case event.Rune() == 'w':
			changeTaskCategory("work")

			return nil
		case event.Rune() == 'b':
			changeTaskCategory("backlog")

			return nil
		case event.Rune() == 'f':
			cycleCategoryFilter()

			return nil
		case event.Key() == tcell.KeyEnter:
			showSegmentDetails()

			return nil
		}

		return event
	})

	// Start a goroutine to update the description view every minute for ongoing segments
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			row, _ := table.GetSelection()
			currentIndex := getTaskIndex(row)
			if currentIndex >= 0 && currentIndex < len(watch.Tasks) {
				task := watch.Tasks[currentIndex]
				// Only update if there's an open segment
				if task.IsActive() {
					app.QueueUpdateDraw(func() {
						updateDescriptionView()
					})
				}
			}
		}
	}()

	// Initial setup
	saveAndRefresh()

	err = app.SetRoot(mainLayout, true).EnableMouse(false).Run()
	if err != nil {
		panic(err)
	}
}
