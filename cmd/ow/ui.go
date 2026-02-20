package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/huckleberry-1881/ohgmas-watch/pkg/task"
)

// App holds all the application state and UI components.
type App struct {
	tviewApp      *tview.Application
	watch         *task.Watch
	tasksFilePath string

	// UI Components
	table           *tview.Table
	descriptionView *tview.TextView
	commandBar      *tview.TextView
	mainLayout      *tview.Flex

	// State
	rowToTaskIndex  []int
	categoryFilter  string
	filterIndex     int
	categoryFilters []string
}

// NewApp creates a new App instance with all UI components initialized.
func NewApp(tasksFilePath string) *App {
	app := &App{
		tviewApp:        tview.NewApplication(),
		tasksFilePath:   tasksFilePath,
		categoryFilters: []string{"", "completed", "work", "backlog"},
		filterIndex:     0,
		categoryFilter:  "",
		rowToTaskIndex:  []int{},
		table:           nil,
		descriptionView: nil,
		commandBar:      nil,
		mainLayout:      nil,
		watch: &task.Watch{
			Tasks: []*task.Task{},
		},
	}

	// Load tasks
	err := app.watch.LoadTasksFromFile(tasksFilePath)
	if err != nil {
		// If we can't load tasks, start with empty watch
		app.watch.Tasks = []*task.Task{}
	}

	// Initialize UI components
	app.initTable()
	app.initDescriptionView()
	app.initCommandBar()
	app.initMainLayout()
	app.setupKeyBindings()
	app.setupSelectionHandler()
	app.startBackgroundUpdater()

	return app
}

// Run starts the TUI application.
func (a *App) Run() error {
	// Initial table population
	a.saveAndRefresh()

	err := a.tviewApp.SetRoot(a.mainLayout, true).EnableMouse(false).Run()
	if err != nil {
		return fmt.Errorf("running TUI application: %w", err)
	}

	return nil
}

// initTable creates and configures the task table.
func (a *App) initTable() {
	a.table = tview.NewTable()
	a.table.SetBorder(true).SetTitle("Tasks")
	a.table.SetSelectable(true, false)
	a.table.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorBlack))
	a.table.SetSeparator(tview.Borders.Vertical)

	// Set up table headers
	headers := []struct {
		text  string
		align int
	}{
		{"Status", tview.AlignCenter},
		{"Task Name", tview.AlignLeft},
		{"Category", tview.AlignCenter},
		{"Tags", tview.AlignLeft},
		{"Last Activity", tview.AlignCenter},
		{"This Week", tview.AlignRight},
		{"Duration", tview.AlignRight},
	}

	for col, header := range headers {
		a.table.SetCell(0, col, tview.NewTableCell(header.text).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetAlign(header.align))
	}
}

// initDescriptionView creates the description pane.
func (a *App) initDescriptionView() {
	a.descriptionView = tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true)
	a.descriptionView.SetBorder(true).SetTitle("Description")
}

// initCommandBar creates the command help bar.
func (a *App) initCommandBar() {
	commandText := "[yellow]Commands:[white] ↑/↓ Navigate | [green]Enter[white] Details | " +
		"[green]t[white] New | [green]m[white] Modify | [green]s[white] Start | [green]n[white] Start+Note | " +
		"[green]e[white] End | [red]d[white] Delete | [blue]c/w/b[white] Category | [purple]f[white] Filter"

	a.commandBar = tview.NewTextView().
		SetDynamicColors(true).
		SetText(commandText)
	a.commandBar.SetBorder(true).SetTitle("Commands")
}

// initMainLayout creates the main layout structure.
func (a *App) initMainLayout() {
	a.mainLayout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.table, 0, 2, true).
		AddItem(a.descriptionView, 0, 1, false).
		AddItem(a.commandBar, 3, 0, false)
}

// setupSelectionHandler sets up the table selection change handler.
func (a *App) setupSelectionHandler() {
	a.table.SetSelectionChangedFunc(func(_, _ int) {
		a.updateDescriptionView()
	})
}

// setupKeyBindings configures all keyboard shortcuts.
func (a *App) setupKeyBindings() {
	a.table.SetInputCapture(a.handleKeyEvent)
}

// handleKeyEvent processes keyboard input for the main table.
func (a *App) handleKeyEvent(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEnter {
		a.showSegmentDetails()

		return nil
	}

	return a.handleRuneKey(event)
}

// handleRuneKey processes character key input.
func (a *App) handleRuneKey(event *tcell.EventKey) *tcell.EventKey {
	handlers := map[rune]func(){
		't': a.showNewTaskForm,
		'm': a.showModifyTaskForm,
		's': a.createSegmentWithoutNote,
		'n': a.showNewSegmentWithNoteForm,
		'e': a.endSegment,
		'd': a.showDeleteConfirmation,
		'c': func() { a.changeTaskCategory("completed") },
		'w': func() { a.changeTaskCategory("work") },
		'b': func() { a.changeTaskCategory("backlog") },
		'f': a.cycleCategoryFilter,
	}

	if handler, ok := handlers[event.Rune()]; ok {
		handler()

		return nil
	}

	return event
}

// startBackgroundUpdater starts a goroutine to update the description view for active segments.
func (a *App) startBackgroundUpdater() {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			row, _ := a.table.GetSelection()
			currentIndex := a.getTaskIndex(row)

			if currentIndex >= 0 && currentIndex < len(a.watch.Tasks) {
				t := a.watch.Tasks[currentIndex]
				if t.IsActive() {
					a.tviewApp.QueueUpdateDraw(func() {
						a.updateDescriptionView()
					})
				}
			}
		}
	}()
}

// getLastActivityDisplay returns the display text and color for a task's last activity.
func (a *App) getLastActivityDisplay(taskItem *task.Task) (string, tcell.Color) {
	lastActivity := taskItem.GetLastActivity()

	if lastActivity.IsZero() {
		return "-", tcell.ColorGray
	}

	text := lastActivity.Format("2006-01-02")

	if taskItem.IsActive() {
		return text, tcell.ColorGreen
	}

	return text, tcell.ColorWhite
}

// getTaskIndex returns the original task index from a table row.
func (a *App) getTaskIndex(tableRow int) int {
	dataRow := tableRow - 1 // -1 because row 0 is headers
	if dataRow >= 0 && dataRow < len(a.rowToTaskIndex) {
		return a.rowToTaskIndex[dataRow]
	}

	return -1
}

// getSelectedTask returns the currently selected task.
func (a *App) getSelectedTask() (*task.Task, bool) {
	row, _ := a.table.GetSelection()
	currentIndex := a.getTaskIndex(row)

	if currentIndex < 0 || currentIndex >= len(a.watch.Tasks) {
		return nil, false
	}

	return a.watch.Tasks[currentIndex], true
}

// saveAndRefresh saves tasks to file and refreshes the table display.
func (a *App) saveAndRefresh() {
	err := a.watch.SaveTasksToFile(a.tasksFilePath)
	if err != nil {
		a.showErrorDialog(err)

		return
	}

	// Clear existing rows (keep header row)
	rowCount := a.table.GetRowCount()
	for r := rowCount - 1; r > 0; r-- {
		a.table.RemoveRow(r)
	}

	// Get tasks sorted by last activity (with optional category filter)
	sortedTasks := a.watch.GetTasksSortedByActivityWithFilter(a.categoryFilter)

	// Update table title to show current filter
	filterTitle := "Tasks"
	if a.categoryFilter != "" {
		filterTitle = fmt.Sprintf("Tasks (%s)", a.categoryFilter)
	}

	a.table.SetTitle(filterTitle)

	// Update the row-to-task mapping
	a.rowToTaskIndex = make([]int, len(sortedTasks))
	for i, t := range sortedTasks {
		a.rowToTaskIndex[i] = a.watch.GetTaskIndex(t)
	}

	// Add task rows using sorted order
	for i, t := range sortedTasks {
		a.renderTaskRow(i+1, t) // +1 because row 0 is headers
	}

	// If we have filtered tasks, select the first data row (row 1)
	if len(sortedTasks) > 0 {
		a.table.Select(1, 0)
	}
}

// renderTaskRow renders a single task row in the table.
func (a *App) renderTaskRow(row int, taskItem *task.Task) {
	cells := a.buildTaskRowCells(taskItem)

	for col, cell := range cells {
		a.table.SetCell(row, col, cell)
	}
}

// buildTaskRowCells creates all cells for a task row.
func (a *App) buildTaskRowCells(taskItem *task.Task) []*tview.TableCell {
	return []*tview.TableCell{
		a.createStatusCell(taskItem),
		a.createNameCell(taskItem),
		a.createCategoryCell(taskItem),
		a.createTagsCell(taskItem),
		a.createLastActivityCell(taskItem),
		a.createThisWeekCell(taskItem),
		a.createDurationCell(taskItem),
	}
}

// createStatusCell creates the status indicator cell.
func (a *App) createStatusCell(taskItem *task.Task) *tview.TableCell {
	cell := tview.NewTableCell("").SetAlign(tview.AlignCenter)

	if taskItem.IsActive() {
		cell.SetText("▶").SetTextColor(tcell.ColorRed)
	} else {
		cell.SetText("●").SetTextColor(tcell.ColorGray)
	}

	return cell
}

// createNameCell creates the task name cell.
func (a *App) createNameCell(taskItem *task.Task) *tview.TableCell {
	return tview.NewTableCell(taskItem.Name).
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft)
}

// createCategoryCell creates the category cell with appropriate coloring.
func (a *App) createCategoryCell(taskItem *task.Task) *tview.TableCell {
	category := taskItem.GetCategory()
	if category == "" {
		category = "work"
	}

	colorMap := map[string]tcell.Color{
		"completed": tcell.ColorGreen,
		"work":      tcell.ColorYellow,
		"backlog":   tcell.ColorGray,
	}

	color := tcell.ColorWhite
	if c, ok := colorMap[category]; ok {
		color = c
	}

	return tview.NewTableCell(category).
		SetTextColor(color).
		SetAlign(tview.AlignCenter)
}

// createTagsCell creates the tags cell.
func (a *App) createTagsCell(taskItem *task.Task) *tview.TableCell {
	tagsText := ""
	if len(taskItem.Tags) > 0 {
		tagsText = "(" + strings.Join(taskItem.Tags, ", ") + ")"
	}

	return tview.NewTableCell(tagsText).
		SetTextColor(tcell.ColorBlue).
		SetAlign(tview.AlignLeft)
}

// createLastActivityCell creates the last activity cell.
func (a *App) createLastActivityCell(taskItem *task.Task) *tview.TableCell {
	text, color := a.getLastActivityDisplay(taskItem)

	return tview.NewTableCell(text).
		SetTextColor(color).
		SetAlign(tview.AlignCenter)
}

// createThisWeekCell creates the this week duration cell.
func (a *App) createThisWeekCell(taskItem *task.Task) *tview.TableCell {
	weekStart := getLastMonday()
	thisWeekDuration := taskItem.GetThisWeekDuration(weekStart)

	return tview.NewTableCell(formatDuration(thisWeekDuration)).
		SetTextColor(tcell.ColorLightBlue).
		SetAlign(tview.AlignRight)
}

// createDurationCell creates the total duration cell.
func (a *App) createDurationCell(taskItem *task.Task) *tview.TableCell {
	duration := taskItem.GetClosedSegmentsDuration()

	return tview.NewTableCell(formatDuration(duration)).
		SetTextColor(tcell.ColorYellow).
		SetAlign(tview.AlignRight)
}

// updateDescriptionView updates the description pane for the current selection.
func (a *App) updateDescriptionView() {
	row, _ := a.table.GetSelection()
	currentIndex := a.getTaskIndex(row)

	if currentIndex < 0 || currentIndex >= len(a.watch.Tasks) {
		a.descriptionView.SetText("")

		return
	}

	selectedTask := a.watch.Tasks[currentIndex]
	content := a.buildDescriptionContent(selectedTask)
	a.descriptionView.SetText(content)
}

// buildDescriptionContent builds the description pane content for a task.
func (a *App) buildDescriptionContent(selectedTask *task.Task) string {
	var content strings.Builder

	lastSegment := selectedTask.GetLastSegment()
	if lastSegment != nil {
		a.writeSegmentInfo(&content, selectedTask, lastSegment)
	}

	content.WriteString(selectedTask.Description)

	return content.String()
}

// writeSegmentInfo writes segment information to the content builder.
func (a *App) writeSegmentInfo(content *strings.Builder, selectedTask *task.Task, seg *task.Segment) {
	if seg.Finish.IsZero() {
		_, _ = fmt.Fprintf(content, "[yellow]Current Segment:[white] Started %s\n",
			seg.Create.Format("2006-01-02 15:04:05"))

		currentDuration := selectedTask.GetCurrentSegmentDuration()

		_, _ = fmt.Fprintf(content, "[yellow]Duration:[white] %s (ongoing)\n\n",
			formatDuration(currentDuration))

		return
	}

	_, _ = fmt.Fprintf(content, "[green]Last Segment:[white] Ended %s\n",
		seg.Finish.Format("2006-01-02 15:04:05"))

	segmentDuration := seg.Finish.Sub(seg.Create)

	_, _ = fmt.Fprintf(content, "[green]Duration:[white] %s\n\n",
		formatDuration(segmentDuration))
}

// styleForm applies consistent styling to a form.
func styleForm(form *tview.Form) {
	form.SetLabelColor(tcell.ColorWhite)
	form.SetFieldBackgroundColor(tcell.ColorGray)
	form.SetFieldTextColor(tcell.ColorGreen)
	form.SetButtonTextColor(tcell.ColorWhite)
}

// centerForm creates a centered layout for a form.
func centerForm(form *tview.Form) *tview.Flex {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 0, 2, true).
			AddItem(nil, 0, 1, false), 0, 2, true).
		AddItem(nil, 0, 1, false)
}

// showErrorDialog displays an error message in a modal dialog.
func (a *App) showErrorDialog(err error) {
	modal := tview.NewModal().
		SetText("Error: " + err.Error()).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(_ int, _ string) {
			a.tviewApp.SetRoot(a.mainLayout, true)
		})
	modal.SetBackgroundColor(tcell.ColorDarkRed)
	a.tviewApp.SetRoot(modal, true)
}

// showNewTaskForm displays the form for creating a new task.
func (a *App) showNewTaskForm() {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle("New Task")
	styleForm(form)

	var name, description, tags string

	form.AddInputField("Name:", "", 70, nil, func(text string) {
		name = text
	})
	form.AddTextArea("Description:", "", 70, 4, 1000, func(text string) {
		description = text
	})
	form.AddInputField("Tags (comma-separated):", "", 70, nil, func(text string) {
		tags = text
	})

	form.AddButton("Create", func() {
		if name == "" {
			return
		}

		tagList := parseTagsFromString(tags)
		a.watch.AddTask(name, description, tagList, "work")
		a.saveAndRefresh()
		a.tviewApp.SetRoot(a.mainLayout, true)
	})

	form.AddButton("Cancel", func() {
		a.tviewApp.SetRoot(a.mainLayout, true)
	})

	a.tviewApp.SetRoot(centerForm(form), true)
}

// showModifyTaskForm displays the form for modifying an existing task.
func (a *App) showModifyTaskForm() {
	selectedTask, ok := a.getSelectedTask()
	if !ok {
		return
	}

	form := tview.NewForm()
	form.SetBorder(true).SetTitle("Modify Task")
	styleForm(form)

	name := selectedTask.Name
	description := selectedTask.Description
	tags := strings.Join(selectedTask.Tags, ", ")

	form.AddInputField("Name:", name, 70, nil, func(text string) {
		name = text
	})
	form.AddTextArea("Description:", description, 70, 4, 1000, func(text string) {
		description = text
	})
	form.AddInputField("Tags (comma-separated):", tags, 70, nil, func(text string) {
		tags = text
	})

	form.AddButton("OK", func() {
		if name == "" {
			return
		}

		tagList := parseTagsFromString(tags)
		selectedTask.Name = name
		selectedTask.Description = description
		selectedTask.Tags = tagList

		a.saveAndRefresh()
		a.tviewApp.SetRoot(a.mainLayout, true)
	})

	form.AddButton("Cancel", func() {
		a.tviewApp.SetRoot(a.mainLayout, true)
	})

	a.tviewApp.SetRoot(centerForm(form), true)
}

// showNewSegmentWithNoteForm displays the form for creating a segment with a note.
func (a *App) showNewSegmentWithNoteForm() {
	selectedTask, ok := a.getSelectedTask()
	if !ok {
		return
	}

	form := tview.NewForm()
	form.SetBorder(true).SetTitle("New Segment")
	styleForm(form)

	var note string

	form.AddTextArea("Note:", "", 50, 3, 300, func(text string) {
		note = text
	})

	form.AddButton("Create", func() {
		if selectedTask.HasUnclosedSegment() {
			a.tviewApp.SetRoot(a.mainLayout, true)

			return
		}

		selectedTask.AddSegment(note)
		a.saveAndRefresh()
		a.tviewApp.SetRoot(a.mainLayout, true)
	})

	form.AddButton("Cancel", func() {
		a.tviewApp.SetRoot(a.mainLayout, true)
	})

	a.tviewApp.SetRoot(centerForm(form), true)
}

// createSegmentWithoutNote creates a new segment without a note.
func (a *App) createSegmentWithoutNote() {
	selectedTask, ok := a.getSelectedTask()
	if !ok {
		return
	}

	if selectedTask.HasUnclosedSegment() {
		return
	}

	selectedTask.AddSegment("")
	a.saveAndRefresh()
}

// endSegment closes the current open segment.
func (a *App) endSegment() {
	selectedTask, ok := a.getSelectedTask()
	if !ok {
		return
	}

	selectedTask.CloseSegment()
	a.saveAndRefresh()
}

// changeTaskCategory changes the category of the selected task.
func (a *App) changeTaskCategory(category string) {
	selectedTask, ok := a.getSelectedTask()
	if !ok {
		return
	}

	selectedTask.SetCategory(category)
	a.saveAndRefresh()
}

// cycleCategoryFilter cycles through category filters.
func (a *App) cycleCategoryFilter() {
	a.filterIndex = (a.filterIndex + 1) % len(a.categoryFilters)
	a.categoryFilter = a.categoryFilters[a.filterIndex]
	a.saveAndRefresh()
}

// showDeleteConfirmation shows a confirmation dialog before deleting a task.
func (a *App) showDeleteConfirmation() {
	selectedTask, ok := a.getSelectedTask()
	if !ok {
		return
	}

	deleteMsg := "Delete task \"" + selectedTask.Name + "\"?\n\nThis action cannot be undone."
	modal := tview.NewModal().
		SetText(deleteMsg).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, _ string) {
			if buttonIndex == 0 {
				a.deleteSelectedTask()
			}

			a.tviewApp.SetRoot(a.mainLayout, true)
		})
	modal.SetBackgroundColor(tcell.ColorDarkRed)
	a.tviewApp.SetRoot(modal, true)
}

// deleteSelectedTask removes the currently selected task.
func (a *App) deleteSelectedTask() {
	row, _ := a.table.GetSelection()
	currentIndex := a.getTaskIndex(row)

	if currentIndex < 0 || currentIndex >= len(a.watch.Tasks) {
		return
	}

	// Remove the task from the slice
	a.watch.Tasks = append(a.watch.Tasks[:currentIndex], a.watch.Tasks[currentIndex+1:]...)
	a.saveAndRefresh()
}

// showSegmentDetails displays a detailed view of all segments for the selected task.
func (a *App) showSegmentDetails() {
	selectedTask, ok := a.getSelectedTask()
	if !ok {
		return
	}

	segmentView := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(true).
		SetScrollable(true)
	segmentView.SetBorder(true).SetTitle("Segments for: " + selectedTask.Name)
	segmentView.SetText(a.buildSegmentDetailsContent(selectedTask))

	segmentLayout := a.createSegmentLayout(segmentView)
	a.tviewApp.SetRoot(segmentLayout, true)
}

// buildSegmentDetailsContent builds the content for the segment details view.
func (a *App) buildSegmentDetailsContent(selectedTask *task.Task) string {
	var content strings.Builder

	if selectedTask.Description != "" {
		content.WriteString("[cyan]Description:[-]\n")
		content.WriteString(selectedTask.Description + "\n\n")
		content.WriteString("[yellow]---[-]\n\n")
	}

	if len(selectedTask.Segments) == 0 {
		content.WriteString("[gray]No segments found for this task.[-]\n")

		return content.String()
	}

	for i, segment := range selectedTask.Segments {
		a.writeSegmentDetailEntry(&content, i+1, segment)
	}

	return content.String()
}

// writeSegmentDetailEntry writes a single segment entry to the content builder.
func (a *App) writeSegmentDetailEntry(content *strings.Builder, num int, segment *task.Segment) {
	_, _ = fmt.Fprintf(content, "[white]Segment %d:[-]\n", num)
	_, _ = fmt.Fprintf(content, "  [green]Created:[-] %s\n", segment.Create.Format("2006-01-02 15:04:05"))

	if segment.Finish.IsZero() {
		content.WriteString("  [red]Status:[-] Open\n")

		duration := time.Since(segment.Create)

		_, _ = fmt.Fprintf(content, "  [yellow]Duration:[-] %s (ongoing)\n", formatDuration(duration))
	} else {
		_, _ = fmt.Fprintf(content, "  [green]Finished:[-] %s\n", segment.Finish.Format("2006-01-02 15:04:05"))

		duration := segment.Finish.Sub(segment.Create)

		_, _ = fmt.Fprintf(content, "  [yellow]Duration:[-] %s\n", formatDuration(duration))
	}

	if segment.Note != "" {
		_, _ = fmt.Fprintf(content, "  [cyan]Note:[-] %s\n", segment.Note)
	} else {
		content.WriteString("  [gray]Note: (none)[-]\n")
	}

	content.WriteString("\n")
}

// createSegmentLayout creates the layout for the segment details view.
func (a *App) createSegmentLayout(segmentView *tview.TextView) *tview.Flex {
	backButton := tview.NewButton("Back to Tasks").SetSelectedFunc(func() {
		a.tviewApp.SetRoot(a.mainLayout, true)
	})
	backButton.SetBackgroundColor(tcell.ColorGray)
	backButton.SetLabelColor(tcell.ColorWhite)

	segmentView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			a.tviewApp.SetRoot(a.mainLayout, true)

			return nil
		}

		return event
	})

	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(segmentView, 0, 1, true).
		AddItem(backButton, 1, 0, false)
}

// parseTagsFromString parses a comma-separated string of tags into a slice.
func parseTagsFromString(tags string) []string {
	tagList := []string{}

	if tags != "" {
		for tag := range strings.SplitSeq(tags, ",") {
			tagList = append(tagList, strings.TrimSpace(tag))
		}
	}

	return tagList
}
