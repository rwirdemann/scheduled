package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rwirdemann/nestiles/panel"
	"github.com/rwirdemann/scheduled"
	"github.com/rwirdemann/scheduled/file"
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

type repository interface {
	Load() []scheduled.Task
}

type model struct {
	root  panel.Model
	focus int

	lists      map[int]*list.Model
	repository repository
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl-c", "q":
			return m, tea.Quit
		case "shift+right":
			if focusedPanel, exists := m.root.Focused(); exists {
				m = m.moveTask(focusedPanel.ID, focusedPanel.ID+1)
			}
		case "shift+left":
			if focusedPanel, exists := m.root.Focused(); exists {
				m = m.moveTask(focusedPanel.ID, focusedPanel.ID-1)
			}
		}
	}
	m.root, cmd = m.root.Update(msg)
	cmds = append(cmds, cmd)

	// find focused pane and Update() its task list
	if focusedPanel, exists := m.root.Focused(); exists {
		if list, exists := m.lists[focusedPanel.ID]; exists {
			*list, cmd = list.Update(msg)
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) moveTask(from, to int) model {
	if _, exists := m.lists[to]; !exists {
		to = 0
	}
	if from == to {
		return m
	}

	if t := m.lists[from].SelectedItem(); t != nil {
		m.lists[from].RemoveItem(m.lists[from].Index())
		m.lists[to].InsertItem(0, t)
	}
	return m
}

func (m model) loadTasks() {
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

func (m model) View() string {
	return m.root.View(m)
}

func renderPanel(m tea.Model, panelID int, w, h int) string {
	model := m.(model)
	if list, exists := model.lists[panelID]; exists {
		list.SetSize(w, h)
		return list.View()
	}
	return ""
}

func main() {
	row1 := panel.New().WithId(20).WithRatio(50).WithLayout(panel.LayoutDirectionHorizontal)
	for i := range 4 {
		p := panel.New().WithId(i).WithRatio(25).WithBorder().WithContent(renderPanel)
		if i == 0 {
			p = p.Focus()
		}
		row1 = row1.Append(p)
	}

	row2 := panel.New().WithId(30).WithRatio(50).WithLayout(panel.LayoutDirectionHorizontal)
	for i := 4; i < 8; i++ {
		p := panel.New().WithId(i).WithRatio(25).WithBorder().WithContent(renderPanel)
		row2 = row2.Append(p)
	}

	rootPanel := panel.New().WithId(10).WithRatio(100).WithLayout(panel.LayoutDirectionVertical).
		Append(row1).
		Append(row2)

	m := model{root: rootPanel, lists: make(map[int]*list.Model), repository: file.Repository{}}
	defaultDelegate := list.NewDefaultDelegate()
	defaultDelegate.ShowDescription = false
	defaultDelegate.SetSpacing(0)
	for i := Inbox; i <= Sunday; i++ {
		l := list.New([]list.Item{}, defaultDelegate, 0, 0)
		l.Title = days[i]
		m.lists[i] = &l
	}
	m.loadTasks()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
}
