package board

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/rwirdemann/scheduled"
	"github.com/rwirdemann/scheduled/date"
)

const (
	Inbox     = 0
	Monday    = 1
	Tuesday   = 2
	Wednesday = 3
	Thursday  = 4
	Friday    = 5
	Saturday  = 6
	Sunday    = 7
)

var Days = map[int]string{
	Inbox:     "Inbox",
	Monday:    "Monday",
	Tuesday:   "Tuesday",
	Wednesday: "Wednesday",
	Thursday:  "Thursday",
	Friday:    "Friday",
	Saturday:  "Saturday",
	Sunday:    "Sunday",
}

// Model represents the main application model managing tasks and their context.
type Model struct {
	plan            *scheduled.Plan
	LastFocus       int
	lists           map[int]*ListModel
	week            int
	selectedContext scheduled.Context
}

// NewModel creates a new instance of the application model with the provided
// Plan.
func NewModel(plan *scheduled.Plan) *Model {
	m := &Model{
		plan:            plan,
		LastFocus:       Inbox,
		selectedContext: scheduled.ContextNone,
		lists:           make(map[int]*ListModel),
	}
	defaultDelegate := list.NewDefaultDelegate()
	defaultDelegate.ShowDescription = false
	defaultDelegate.SetSpacing(0)
	for i := Inbox; i <= Sunday; i++ {
		l := list.New([]list.Item{}, defaultDelegate, 0, 0)
		l.SetShowStatusBar(false)
		l.SetShowHelp(false)
		m.lists[i] = NewListModel(l)
	}

	m.refreshLists()

	for i := Monday; i <= Sunday; i++ {
		m.lists[i].Deselect()
	}

	_, w := time.Now().ISOWeek()
	m.setWeek(w)

	return m
}

// Week returns the current week number stored in the Model.
func (m *Model) Week() int {
	return m.week
}

// GetSelectedContext returns the currently selected context in the Model.
func (m *Model) GetSelectedContext() scheduled.Context {
	return m.selectedContext
}

// SetContext sets the currently selected context in the Model.
func (m *Model) SetContext(context scheduled.Context) {
	m.selectedContext = context
	m.refreshLists()
}

// DecWeek decreases the current week, wrapping to 52 if below 1.
func (m *Model) DecWeek() {
	if m.week > 1 {
		m.setWeek(m.week - 1)
	} else {
		m.setWeek(52)
	}
}

// IncWeek increases the current week, wrapping to 1 if above 52.
func (m *Model) IncWeek() {
	if m.week < 52 {
		m.setWeek(m.week + 1)
	} else {
		m.setWeek(1)
	}
}

// refreshLists populates all lists with its actual task according to the
// selected context.
func (m *Model) refreshLists() {
	for day := Inbox; day <= Sunday; day++ {
		tasks := m.plan.TasksForDayAndContext(day, m.selectedContext.ID)
		items := make([]list.Item, len(tasks))
		for i, task := range tasks {
			items[i] = ListItem{Task: task}
		}
		m.lists[day].SetItems(items)
	}
}

// UpdateTask updates the name and context of the selected task.
func (m *Model) UpdateTask(name string, context int, description string) {
	task, ok := m.GetSelectedTask(m.LastFocus)
	if !ok {
		return
	}

	if err := m.plan.UpdateTask(task.ID, name, context, description); err != nil {
		return
	}

	m.refreshLists()
}

// CreateTask creates a new task with the given name and context.
func (m *Model) CreateTask(name string, context int, description string) {
	day := m.LastFocus
	if parsedDay, cleanName := scheduled.ParseWeekday(name); parsedDay != 0 {
		day = parsedDay
		name = cleanName
	}
	m.plan.CreateTask(name, context, description, day)
	m.refreshLists()
}

// AddTask implements scheduled.InputPort. Adds a task to the given day
// with ContextNone. Pass day=0 (Inbox) when no weekday prefix was detected.
// Must be called from the Bubble Tea Update goroutine.
func (m *Model) AddTask(name string, day int) {
	m.plan.CreateTask(name, scheduled.ContextNone.ID, "", day)
	m.refreshLists()
}

// Compile-time check: ensures *Model implements scheduled.InputPort.
// If AddTask is ever removed or renamed, the build fails immediately.
var _ scheduled.InputPort = (*Model)(nil)

// SetListTitle sets the title of the list at the given index.
func (m *Model) SetListTitle(listIndex int, title string) {
	m.lists[listIndex].Title = fmt.Sprintf("%s - %s", title, m.selectedContext.Name)
}

// MoveUp moves the selected item up in the list at the given index.
func (m *Model) MoveUp(listIndex int) {
	task, ok := m.GetSelectedTask(listIndex)
	if !ok {
		return
	}
	if err := m.plan.MoveTaskUp(task.ID); err != nil {
		return
	}
	m.refreshLists()

	// Keep selection
	l := m.lists[listIndex]
	l.Select(l.Index() - 1)
}

// MoveDown moves the selected item down in the list at the given index.
func (m *Model) MoveDown(listIndex int) {
	task, ok := m.GetSelectedTask(listIndex)
	if !ok {
		return
	}
	if err := m.plan.MoveTaskDown(task.ID); err != nil {
		return
	}
	m.refreshLists()

	// Keep selection
	l := m.lists[listIndex]
	l.Select(l.Index() + 1)
}

// ToggleDone toggles the done state of the selected task in the list at the
// given index.
func (m *Model) ToggleDone(listIndex int) {
	task, ok := m.GetSelectedTask(listIndex)
	if !ok {
		return
	}
	if err := m.plan.ToggleDone(task.ID); err != nil {
		return
	}
	m.refreshLists()
}

// DeleteTask deletes the selected task in the list at the given index.
func (m *Model) DeleteTask(listIndex int) {
	task, ok := m.GetSelectedTask(listIndex)
	if !ok {
		return
	}
	if err := m.plan.DeleteDoneTask(task.ID); err != nil {
		return
	}
	m.refreshLists()
}

// MoveTask moves the selected task from one list to another.
func (m *Model) MoveTask(from, to int) {
	if from < Inbox || from > Sunday {
		return
	}
	if to < Inbox || to > Sunday {
		return
	}
	if from == to {
		return
	}

	task, ok := m.GetSelectedTask(from)
	if !ok {
		return
	}

	l := m.lists[from]
	idx := l.Index()
	count := len(l.Items())

	if err := m.plan.MoveTaskToDay(task.ID, to); err != nil {
		return
	}
	m.refreshLists()

	if count > 1 {
		if idx >= count-1 {
			m.lists[from].Select(idx - 1)
		} else {
			m.lists[from].Select(idx)
		}
	}
}

// GetSelectedTask returns the selected task in the list at the given index,
// if any.
func (m *Model) GetSelectedTask(listIndex int) (scheduled.Task, bool) {
	if l, exists := m.lists[listIndex]; exists {
		if i := l.SelectedItem(); i != nil {
			return i.(ListItem).Task, true
		}
	}
	return scheduled.Task{}, false
}

// GetTasksForPanel returns a slice of tasks for the specified panel index.
func (m *Model) GetTasksForPanel(listIndex int) []scheduled.Task {
	if listIndex < Inbox || listIndex > Sunday {
		return []scheduled.Task{}
	}
	return m.plan.TasksForDayAndContext(listIndex, m.selectedContext.ID)
}

// Update updates the model based on the given message and returns a command to
// be executed after the update.
func (m *Model) Update(listIndex int, msg tea.Msg) tea.Cmd {
	if l, exists := m.lists[listIndex]; exists {
		updated, cmd := l.Update(msg)
		m.lists[listIndex].Model = updated
		return cmd
	}
	return nil
}

// DeselectAndRestoreIndex deselects the currently focused list and restores the
// selection of the newly focused list.
func (m *Model) DeselectAndRestoreIndex(focusedPanelID int) {
	if currentList, exists := m.lists[m.LastFocus]; exists {
		currentList.SaveIndex()
		currentList.Deselect()
	}
	m.LastFocus = focusedPanelID
	if nextList, exists := m.lists[focusedPanelID]; exists {
		nextList.RestoreIndex()
	}
}

// Render returns the rendered view of the list at the given index.
func (m *Model) Render(panelID int, w, h int) string {
	if l, exists := m.lists[panelID]; exists {
		l.Model.SetSize(w-2, h)
		return l.Model.View()
	}
	return ""
}

// IsContextUsed returns true if the given context is used in any of the tasks
// in the model.
func (m *Model) IsContextUsed(c scheduled.Context) bool {
	return m.plan.IsContextUsed(c.ID)
}

func (m *Model) setWeek(week int) {
	m.week = week
	for i := Inbox; i <= Sunday; i++ {
		monday := date.GetMondayOfWeek(m.week)
		if i == Inbox {
			m.lists[i].Title = fmt.Sprintf("[ESC] Inbox (Week %d) - %s",
				m.week, m.selectedContext.Name)
		} else {
			day := monday.AddDate(0, 0, i-1)
			m.lists[i].Title = fmt.Sprintf("[%d] %s (%s)",
				i, Days[i], day.Format("02.01.2006"))
		}
	}
}
