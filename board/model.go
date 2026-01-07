package board

import (
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rwirdemann/scheduled"
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

type Model struct {
	repository repository
	LastFocus  int
	Lists      map[int]*ListModel
}

func NewModel(repository repository) *Model {
	m := &Model{
		repository: repository,
		LastFocus:  Inbox,
		Lists:      make(map[int]*ListModel),
	}
	defaultDelegate := list.NewDefaultDelegate()
	defaultDelegate.ShowDescription = false
	defaultDelegate.SetSpacing(0)
	for i := Inbox; i <= Sunday; i++ {
		l := list.New([]list.Item{}, defaultDelegate, 0, 0)
		l.SetShowStatusBar(false)
		l.SetShowHelp(false)
		m.Lists[i] = NewListModel(l)
	}
	m.loadTasks()

	// Deselect all lists except the focused one (Inbox)
	for i := Inbox; i <= Sunday; i++ {
		if i != Inbox {
			m.Lists[i].Deselect()
		}
	}

	return m
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

	for day := range m.Lists {
		for i, item := range tasksByDay[day] {
			m.Lists[day].InsertItem(i, item)
		}
	}
}

func (m *Model) UpdateTask(listIndex int, name string, context int) {
	list := m.Lists[listIndex]
	task := list.SelectedItem().(scheduled.Task)
	task.Name = name
	task.Context = context
	index := list.Index()
	list.RemoveItem(index)
	list.InsertItem(index, task)
}

func (m *Model) CreateTask(listIndex int, name string, context int) {
	t := scheduled.Task{Name: name, Context: context, Day: listIndex}
	list := m.Lists[listIndex]
	list.InsertItem(len(list.Items()), t)
}

func (m *Model) SetListTitle(listIndex int, title string) {
	m.Lists[listIndex].Title = title
}

func (m *Model) MoveUp(listIndex int) {
	if l, exists := m.Lists[listIndex]; exists {
		l.MoveItemUp()
	}
}

func (m *Model) MoveDown(listIndex int) {
	if l, exists := m.Lists[listIndex]; exists {
		l.MoveItemDown()
	}
}

func (m *Model) ToggleDone(listIndex int) {
	if l, exists := m.Lists[listIndex]; exists {
		l.ToggleDone()
	}
}

func (m *Model) DeleteTask(listIndex int) {
	if l, exists := m.Lists[listIndex]; exists {
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

	if item := m.Lists[from].SelectedItem(); item != nil {
		t := item.(scheduled.Task)
		t.Day = to
		m.Lists[from].RemoveItem(m.Lists[from].Index())
		m.Lists[to].InsertItem(len(m.Lists[to].Items()), t)
	}
}

func (m *Model) GetSelectedTask(listIndex int) (scheduled.Task, bool) {
	if l, exists := m.Lists[listIndex]; exists {
		if i := l.SelectedItem(); i != nil {
			return i.(scheduled.Task), true
		}
	}
	return scheduled.Task{}, false
}

func (m *Model) Update(listIndex int, msg tea.Msg) tea.Cmd {
	if l, exists := m.Lists[listIndex]; exists {
		updated, cmd := l.Update(msg)
		m.Lists[listIndex].Model = updated
		return cmd
	}
	return nil
}

// DeselectAndRestoreIndex deselects the currently focused list and restores
// the selection of the newly focused list.
func (m *Model) DeselectAndRestoreIndex(focusedPanelID int) {
	if currentList, exists := m.Lists[m.LastFocus]; exists {
		currentList.SaveIndex()
		currentList.Deselect()
	}
	m.LastFocus = focusedPanelID
	if nextList, exists := m.Lists[focusedPanelID]; exists {
		nextList.RestoreIndex()
	}
}

type repository interface {
	Load() []scheduled.Task
	Save(tasks []scheduled.Task)
}
