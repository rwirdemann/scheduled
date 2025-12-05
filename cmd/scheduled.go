package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rwirdemann/nestiles/panel"
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

type model struct {
	root  panel.Model
	focus int

	lists map[int]list.Model
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
		if _, exists := m.lists[focusedPanel.ID]; exists {
			m.lists[focusedPanel.ID], cmd = m.lists[focusedPanel.ID].Update(msg)
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

	fromList := m.lists[from]
	toList := m.lists[to]
	if t := fromList.SelectedItem(); t != nil {
		fromList.RemoveItem(fromList.Index())
		toList.InsertItem(0, t)
		m.lists[from] = fromList
		m.lists[to] = toList
	}
	return m
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

	inboxItems := []list.Item{
		Task{Name: "Book Surf Course", Desc: "Wingfoil prefered"},
		Task{Name: "Rent Equipment", Desc: "Gong or Armstrong"},
	}

	mondayItems := []list.Item{
		Task{Name: "Flug buch", Desc: "Möglichst bei Condor"},
		Task{Name: "Auto mieten", Desc: "Bei einem lokalen Anbieter"},
	}

	m := model{root: rootPanel, lists: make(map[int]list.Model)}
	inbox := list.New(inboxItems, list.NewDefaultDelegate(), 0, 0)
	inbox.Title = "Inbox"
	m.lists[0] = inbox
	monday := list.New(mondayItems, list.NewDefaultDelegate(), 0, 0)
	monday.Title = "Monday"
	m.lists[1] = monday

	for i := 2; i <= Sunday; i++ {
		l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
		l.Title = days[i]
		m.lists[i] = l
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
}

type Task struct {
	Name string `json:"name"`
	Desc string `json:"description"`
	Day  int    `json:"day"`
	Done bool   `json:"done"`
	Pos  int    `json:"pos"`
}

func (i Task) Title() string       { return i.Name }
func (i Task) Description() string { return i.Desc }
func (i Task) FilterValue() string { return i.Name }
