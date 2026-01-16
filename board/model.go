package board

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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

var days = map[int]string{
	Inbox:     "Inbox",
	Monday:    "Monday",
	Tuesday:   "Tuesday",
	Wednesday: "Wednesday",
	Thursday:  "Thursday",
	Friday:    "Friday",
	Saturday:  "Saturday",
	Sunday:    "Sunday",
}

type Model struct {
	repository      repository
	LastFocus       int
	lists           map[int]*ListModel
	week            int
	selectedContext scheduled.Context
}

func NewModel(repository repository) *Model {
	m := &Model{
		repository:      repository,
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
	m.loadTasks()

	// Deselect all lists except the focused one (Inbox)
	for i := Inbox; i <= Sunday; i++ {
		if i != Inbox {
			m.lists[i].Deselect()
		}
	}

	_, w := time.Now().ISOWeek()
	m.setWeek(w)

	return m
}

func (m *Model) Week() int {
	return m.week
}

func (m *Model) SetContext(context scheduled.Context) {
	m.selectedContext = context
}

func (m *Model) DecWeek() {
	if m.week > 1 {
		m.setWeek(m.week - 1)
	} else {
		m.setWeek(52)
	}
}

func (m *Model) IncWeek() {
	if m.week < 52 {
		m.setWeek(m.week + 1)
	} else {
		m.setWeek(1)
	}
}

func (m *Model) loadTasks() {
	var tasksByDay = make(map[int][]list.Item)
	tasks := m.repository.Load()
	for _, task := range tasks {
		tasksByDay[task.Day] = append(tasksByDay[task.Day], task)
	}

	// Sort tasks by their Pos field
	for _, items := range tasksByDay {
		sort.Slice(items, func(i, j int) bool {
			return items[i].(scheduled.Task).Pos < items[j].(scheduled.Task).Pos
		})
	}

	for day := range m.lists {
		for i, item := range tasksByDay[day] {
			m.lists[day].InsertItem(i, item)
		}
	}
}

func (m *Model) UpdateTask(name string, context int) {
	list := m.lists[m.LastFocus]
	task := list.SelectedItem().(scheduled.Task)
	task.Name = name
	task.Context = context
	index := list.Index()
	list.RemoveItem(index)
	list.InsertItem(index, task)
}

func (m *Model) CreateTask(name string, context int) {
	t := scheduled.Task{Name: name, Context: context, Day: m.LastFocus}
	list := m.lists[m.LastFocus]
	list.InsertItem(len(list.Items()), t)
}

func (m *Model) SetListTitle(listIndex int, title string) {
	m.lists[listIndex].Title = fmt.Sprintf("%s - %s", title, m.selectedContext.Name)
}

func (m *Model) MoveUp(listIndex int) {
	if l, exists := m.lists[listIndex]; exists {
		l.MoveItemUp()
	}
}

func (m *Model) MoveDown(listIndex int) {
	if l, exists := m.lists[listIndex]; exists {
		l.MoveItemDown()
	}
}

func (m *Model) ToggleDone(listIndex int) {
	if l, exists := m.lists[listIndex]; exists {
		l.ToggleDone()
	}
}

func (m *Model) DeleteTask(listIndex int) {
	if l, exists := m.lists[listIndex]; exists {
		i := l.SelectedItem()
		if i == nil {
			return
		}
		task := i.(scheduled.Task)
		if task.Done {
			l.RemoveItem(l.Index())
		}
	}
}
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

	if item := m.lists[from].SelectedItem(); item != nil {
		t := item.(scheduled.Task)
		t.Day = to
		m.lists[from].RemoveItem(m.lists[from].Index())
		m.lists[to].InsertItem(len(m.lists[to].Items()), t)
	}
}

func (m *Model) GetSelectedTask(listIndex int) (scheduled.Task, bool) {
	if l, exists := m.lists[listIndex]; exists {
		if i := l.SelectedItem(); i != nil {
			return i.(scheduled.Task), true
		}
	}
	return scheduled.Task{}, false
}

func (m *Model) Update(listIndex int, msg tea.Msg) tea.Cmd {
	if l, exists := m.lists[listIndex]; exists {
		updated, cmd := l.Update(msg)
		m.lists[listIndex].Model = updated
		return cmd
	}
	return nil
}

// DeselectAndRestoreIndex deselects the currently focused list and restores
// the selection of the newly focused list.
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

func (m *Model) SaveTasks() {
	var tasks []scheduled.Task
	for _, list := range m.lists {
		for i, item := range list.Items() {
			t := item.(scheduled.Task)
			t.Pos = i
			tasks = append(tasks, t)
		}
	}
	m.repository.Save(tasks)
}
func (m *Model) Render(panelID int, w, h int) string {
	if list, exists := m.lists[panelID]; exists {
		list.Model.SetSize(w, h)
		return list.Model.View()
	}
	return ""
}

func (m *Model) setWeek(week int) {
	m.week = week
	for i := Inbox; i <= Sunday; i++ {
		monday := date.GetMondayOfWeek(m.week)
		if i == Inbox {
			m.lists[i].Title = fmt.Sprintf("[ESC] Inbox (Week %d) - %s", m.week, m.selectedContext.Name)
		} else {
			day := monday.AddDate(0, 0, i-1)
			m.lists[i].Title = fmt.Sprintf("[%d] %s (%s)", i, days[i], day.Format("02.01.2006"))
		}
	}
}

type repository interface {
	Load() []scheduled.Task
	Save(tasks []scheduled.Task)
}
