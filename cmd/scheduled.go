package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rwirdemann/nestiles/panel"
)

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
		case "shift-right":
		}
	}
	m.root, cmd = m.root.Update(msg)
	cmds = append(cmds, cmd)

	// find focused pane and Update() its task list
	if focusedPanel, exists := m.root.Focused(); exists {
		m.lists[focusedPanel.ID], cmd = m.lists[focusedPanel.ID].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, cmd
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
	rootPanel := panel.New().WithId(10).WithRatio(100).WithLayout(panel.LayoutDirectionHorizontal)
	for i := range 2 {
		p := panel.New().WithId(i).WithRatio(25).WithBorder().WithContent(renderPanel)
		if i == 0 {
			p = p.Focus()
		}
		rootPanel = rootPanel.Append(p)
	}

	inboxItems := []list.Item{
		item{title: "Book Surf Course", desc: "Wingfoil prefered"},
		item{title: "Rent Equipment", desc: "Gong or Armstrong"},
	}

	mondayItems := []list.Item{
		item{title: "Flug buch", desc: "Möglichst bei Condor"},
		item{title: "Auto mieten", desc: "Bei einem lokalen Anbieter"},
	}

	m := model{root: rootPanel, lists: make(map[int]list.Model)}
	inbox := list.New(inboxItems, list.NewDefaultDelegate(), 0, 0)
	inbox.Title = "Inbox"
	m.lists[0] = inbox
	monday := list.New(mondayItems, list.NewDefaultDelegate(), 0, 0)
	monday.Title = "Monday"
	m.lists[1] = monday

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }
