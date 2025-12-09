package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rwirdemann/nestiles/panel"
	"github.com/rwirdemann/scheduled"
	"github.com/rwirdemann/scheduled/date"
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

	panelEdit = 40
	panelHelp = 50
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
	Save(tasks []scheduled.Task)
}

type model struct {
	root  panel.Model
	focus int
	week  int

	lists      map[int]*scheduled.ListModel
	textInput  textinput.Model
	repository repository
	lastFocus  int

	keys scheduled.KeyMap
	help help.Model

	termWidth  int
	termHeight int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if focusedPanel, exists := m.root.Focused(); exists {
		if focusedPanel.ID == panelEdit {
			m.textInput, cmd = m.textInput.Update(msg)
			switch msg := msg.(type) {
			case tea.KeyMsg:
				switch msg.String() {
				case "enter":
					value := m.textInput.Value()
					if len(strings.TrimSpace(value)) > 0 {
						t := scheduled.Task{Name: value}
						m.lists[0].InsertItem(0, t)
					}
					m.textInput.Reset()
				}
			}
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help):
			m.root = m.root.Hide(panelEdit)
			m.root = m.root.Show(panelHelp)
			return m, nil
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Right):
			if m.week < 52 {
				return m.setWeek(m.week + 1), nil
			} else {
				return m.setWeek(1), nil
			}
		case key.Matches(msg, m.keys.Left):
			if m.week > 1 {
				return m.setWeek(m.week - 1), nil
			} else {
				return m.setWeek(52), nil
			}
		case key.Matches(msg, m.keys.ShiftLeft):
			if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
				m = m.moveTask(focusedPanel.ID, focusedPanel.ID-1)
			}
		case key.Matches(msg, m.keys.ShiftRight):
			if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
				m = m.moveTask(focusedPanel.ID, focusedPanel.ID+1)
			}
		case key.Matches(msg, m.keys.ShiftUp):
			focusedPanel, _ := m.root.Focused()
			if l, exists := m.lists[focusedPanel.ID]; exists {
				l.MoveItemUp()
			}
		case key.Matches(msg, m.keys.ShiftDown):
			focusedPanel, _ := m.root.Focused()
			if l, exists := m.lists[focusedPanel.ID]; exists {
				l.MoveItemDown()
			}
		case key.Matches(msg, m.keys.New):
			m.root = m.root.Hide(panelHelp)
			m.root = m.root.Show(panelEdit)
			m.root = m.root.SetFocus(panelEdit)
			m.textInput.Focus()
			return m, nil
		case key.Matches(msg, m.keys.Esc):
			m.root = m.root.Hide(panelEdit)
			m.root = m.root.Hide(panelHelp)
			m.root = m.root.SetFocus(Inbox)
			m.textInput.Reset()
			m.textInput.Blur()
			return m, nil
		case key.Matches(msg, m.keys.Space):
			if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
				if l, exists := m.lists[focusedPanel.ID]; exists {
					l.ToggleDone()
				}
			}
		case key.Matches(msg, m.keys.Back):
			if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
				l := m.lists[focusedPanel.ID]
				selected := l.SelectedItem()
				t := selected.(scheduled.Task)
				if t.Done {
					l.RemoveItem(l.Index())
				}
			}
		case key.Matches(msg, m.keys.Enter):
			value := m.textInput.Value()
			if len(strings.TrimSpace(value)) > 0 {
				t := scheduled.Task{Name: value}
				m.lists[0].InsertItem(0, t)
			}
			m.textInput.Reset()
		}
	}
	m.root, cmd = m.root.Update(msg)
	cmds = append(cmds, cmd)

	// find focused panel and Update() its task list
	if focusedPanel, exists := m.root.Focused(); exists {
		// Handle focus changes
		if m.lastFocus != focusedPanel.ID && focusedPanel.ID != panelEdit {
			// Save index and deselect previously focused list
			if prevList, exists := m.lists[m.lastFocus]; exists {
				prevList.SaveIndex()
				prevList.Deselect()
			}
			// Restore index of newly focused list
			if currList, exists := m.lists[focusedPanel.ID]; exists {
				currList.RestoreIndex()
			}
			m.lastFocus = focusedPanel.ID
		}

		if list, exists := m.lists[focusedPanel.ID]; exists {
			list.Model, cmd = list.Model.Update(msg)
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

	if item := m.lists[from].SelectedItem(); item != nil {
		t := item.(scheduled.Task)
		t.Day = to
		m.lists[from].RemoveItem(m.lists[from].Index())
		m.lists[to].InsertItem(len(m.lists[to].Items()), t)
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

func (m model) saveTasks() {
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

func (m model) View() string {
	m.saveTasks()

	const minWidth = 120
	const minHeight = 40

	if m.termWidth < minWidth || m.termHeight < minHeight {
		return fmt.Sprintf("\n\n  Terminal too small!\n\n  Current size: %dx%d\n  Minimum size: %dx%d\n\n  Please resize your terminal.\n",
			m.termWidth, m.termHeight, minWidth, minHeight)
	}

	return m.root.View(m)
}

func (m model) setWeek(week int) model {
	m.week = week
	for i := Inbox; i <= Sunday; i++ {
		monday := date.GetMondayOfWeek(m.week)
		if i == Inbox {
			m.lists[i].Title = fmt.Sprintf("Inbox (Week %d)", m.week)
		} else {
			day := monday.AddDate(0, 0, i)
			m.lists[i].Title = fmt.Sprintf("%s (%s)", days[i], day.Format("02.01.2006"))
		}
	}
	return m
}

func renderPanel(m tea.Model, panelID int, w, h int) string {
	model := m.(model)
	if panelID == 40 {
		return model.textInput.View()
	}
	if list, exists := model.lists[panelID]; exists {
		list.Model.SetSize(w, h)
		return list.Model.View()
	}
	return ""
}

func renderHelp(m tea.Model, panelID int, w, h int) string {
	model := m.(model)
	return model.help.FullHelpView(model.keys.FullHelp())
}

func main() {
	row1 := panel.New().WithId(20).WithRatio(42).WithLayout(panel.LayoutDirectionHorizontal)
	for i := range 4 {
		p := panel.New().WithId(i).WithRatio(25).WithBorder().WithContent(renderPanel)
		if i == 0 {
			p = p.Focus()
		}
		row1 = row1.Append(p)
	}

	row2 := panel.New().WithId(30).WithRatio(42).WithLayout(panel.LayoutDirectionHorizontal)
	for i := 4; i < 8; i++ {
		p := panel.New().WithId(i).WithRatio(25).WithBorder().WithContent(renderPanel)
		row2 = row2.Append(p)
	}
	row3 := panel.New().WithId(panelEdit).WithRatio(16).WithContent(renderPanel).WithBorder().WithVisible(false)
	helpPanel := panel.New().WithId(panelHelp).WithRatio(16).WithContent(renderHelp).WithBorder().WithVisible(true)

	rootPanel := panel.New().WithId(10).WithRatio(100).WithLayout(panel.LayoutDirectionVertical).
		Append(row1).
		Append(row2).
		Append(row3).
		Append(helpPanel)

	ti := textinput.New()
	ti.Placeholder = "Edit or update task (enter to submit, esc to cancel)"
	ti.Width = 80

	h := help.New()
	h.Styles.FullKey = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))
	h.Styles.FullDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))
	h.Styles.FullSeparator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	m := model{root: rootPanel, lists: make(map[int]*scheduled.ListModel),
		repository: file.Repository{},
		textInput:  ti,
		lastFocus:  Inbox,
		keys:       scheduled.Keys,
		help:       h,
	}
	defaultDelegate := list.NewDefaultDelegate()
	defaultDelegate.ShowDescription = false
	defaultDelegate.SetSpacing(0)
	for i := Inbox; i <= Sunday; i++ {
		l := list.New([]list.Item{}, defaultDelegate, 0, 0)
		l.SetShowStatusBar(false)
		l.SetShowHelp(false)
		m.lists[i] = scheduled.NewListModel(l)
	}
	_, w := time.Now().ISOWeek()
	m = m.setWeek(w)
	m.loadTasks()

	// Deselect all lists except the focused one (Inbox)
	for i := Inbox; i <= Sunday; i++ {
		if i != Inbox {
			m.lists[i].Deselect()
		}
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
}
