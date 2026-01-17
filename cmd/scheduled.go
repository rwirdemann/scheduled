package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/rwirdemann/nestiles/panel"
	"github.com/rwirdemann/scheduled"
	"github.com/rwirdemann/scheduled/board"
	"github.com/rwirdemann/scheduled/file"
)

const (
	panelEdit = 40
	panelHelp = 50
	panelLeft = 60
)

type mode int

const (
	modeNormal mode = iota
	modeEdit
	modeNew
	modeContexts
)

type repository interface {
	LoadContexts() []scheduled.Context
}

type model struct {
	root  panel.Model
	focus int

	board *board.Model

	form           *huh.Form
	taskRepository repository

	showHelp bool
	keys     scheduled.KeyMap
	help     help.Model

	termWidth  int
	termHeight int

	contexts    []scheduled.Context
	contextList list.Model

	mode mode
}

func newModel(root panel.Model, tasksFile string) model {
	h := help.New()
	h.Styles.FullKey = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	h.Styles.FullDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	h.Styles.FullSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	repo := file.NewRepository(tasksFile)
	contextListDelegate := list.NewDefaultDelegate()
	contextListDelegate.ShowDescription = false
	contextListDelegate.SetSpacing(0)
	contexts := repo.LoadContexts()
	items := make([]list.Item, len(contexts))
	for i, v := range contexts {
		items[i] = v
	}
	contextList := list.New(items, contextListDelegate, 0, 0)
	contextList.SetShowHelp(false)
	contextList.SetShowStatusBar(false)
	contextList.Title = "Contexts"

	m := model{
		root:           root,
		taskRepository: repo,
		keys:           scheduled.Keys,
		help:           h,
		showHelp:       true,
		mode:           modeNormal,
		contexts:       contexts,
		contextList:    contextList,
		board:          board.NewModel(repo),
	}
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch m.mode {
	case modeNew, modeEdit:
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
			if f.State == huh.StateCompleted {
				title := m.form.GetString("title")
				context := m.form.GetInt("context")
				if m.mode == modeEdit {
					m.board.UpdateTask(title, context)
				}
				if m.mode == modeNew {
					m.board.CreateTask(title, context)
				}
				m.root = m.root.Hide(panelEdit)
				if m.showHelp {
					m.root = m.root.Show(panelHelp)
				}
				m.root = m.root.SetFocus(m.board.LastFocus)
				m.mode = modeNormal
			}
			if f.State == huh.StateAborted {
				m.root = m.root.Hide(panelEdit)
				if m.showHelp {
					m.root = m.root.Show(panelHelp)
				}
				m.root = m.root.SetFocus(m.board.LastFocus)
				m.mode = modeNormal
			}
		}

		return m, cmd
	case modeContexts:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.keys.Contexts), key.Matches(msg, m.keys.Esc):
				m.mode = modeNormal
				m.root = m.root.Hide(panelLeft)
				m.root = m.root.SetFocus(m.board.LastFocus)
				return m, nil
			case key.Matches(msg, m.keys.Enter):
				m.mode = modeNormal
				i := m.contextList.SelectedItem()
				m.board.SetContext(i.(scheduled.Context))
				m.root = m.root.Hide(panelLeft)
				m.board.SetListTitle(board.Inbox, fmt.Sprintf("[ESC] Inbox (Week %d)", m.board.Week()))
				m.root = m.root.SetFocus(m.board.LastFocus)
				return m, nil
			}
		}
		m.contextList, cmd = m.contextList.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help):
			m.root = m.root.Hide(panelEdit)
			m.showHelp = !m.showHelp
			if m.showHelp {
				m.root = m.root.Show(panelHelp)
			} else {
				m.root = m.root.Hide(panelHelp)
			}
			return m, nil
		case key.Matches(msg, m.keys.Quit):
			m.board.SaveTasks()
			return m, tea.Quit
		case key.Matches(msg, m.keys.Right):
			m.board.IncWeek()
			return m, nil
		case key.Matches(msg, m.keys.Left):
			m.board.DecWeek()
			return m, nil
		case key.Matches(msg, m.keys.ShiftLeft):
			if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
				m.board.MoveTask(focusedPanel.ID, focusedPanel.ID-1)
			}
		case key.Matches(msg, m.keys.ShiftRight):
			if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
				m.board.MoveTask(focusedPanel.ID, focusedPanel.ID+1)
			}
		case key.Matches(msg, m.keys.ShiftUp):
			focusedPanel, _ := m.root.Focused()
			m.board.MoveUp(focusedPanel.ID)
		case key.Matches(msg, m.keys.ShiftDown):
			focusedPanel, _ := m.root.Focused()
			m.board.MoveDown(focusedPanel.ID)
		case key.Matches(msg, m.keys.New):
			// Preselect the currently selected context
			selectedContext := m.board.GetSelectedContext()
			prefilledTask := &scheduled.Task{Context: selectedContext.ID}
			m.form = scheduled.CreateTaskForm(prefilledTask, m.contexts)
			m.root = m.root.Hide(panelHelp)
			m.root = m.root.Show(panelEdit)
			m.root = m.root.SetFocus(panelEdit)
			m.mode = modeNew
			return m, m.form.Init()
		case key.Matches(msg, m.keys.Esc):
			m.root = m.root.Hide(panelEdit)
			m.root = m.root.SetFocus(board.Inbox)
			m.board.DeselectAndRestoreIndex(board.Inbox)
			return m, nil
		case key.Matches(msg, m.keys.Space):
			if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
				m.board.ToggleDone(focusedPanel.ID)
			}
		case key.Matches(msg, m.keys.Back):
			if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
				m.board.DeleteTask(focusedPanel.ID)
			}
		case key.Matches(msg, m.keys.Enter):
			focusedPanel, _ := m.root.Focused()
			if t, exists := m.board.GetSelectedTask(focusedPanel.ID); exists {
				m.form = scheduled.CreateTaskForm(&t, m.contexts)
				m.root = m.root.Hide(panelHelp)
				m.root = m.root.Show(panelEdit)
				m.root = m.root.SetFocus(panelEdit)
				m.mode = modeEdit
				return m, m.form.Init()
			}
		case key.Matches(msg, m.keys.Num):
			key := msg.String()
			panelNum, _ := strconv.Atoi(key)
			m.root = m.root.SetFocus(panelNum)
			m.board.DeselectAndRestoreIndex(panelNum)
			return m, nil
		case key.Matches(msg, m.keys.MoveToToday):
			today := time.Now().Weekday()
			if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
				m.board.MoveTask(focusedPanel.ID, int(today))
			}
		case key.Matches(msg, m.keys.MoveToInbox):
			if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
				m.board.MoveTask(focusedPanel.ID, board.Inbox)
			}
		case key.Matches(msg, m.keys.Contexts):
			m.mode = modeContexts
			m.root = m.root.Show(panelLeft)
			m.root.SetFocus(panelLeft)
			return m, nil
		}
	}
	m.root, cmd = m.root.Update(msg)
	cmds = append(cmds, cmd)

	// find focused panel and Update() its task list
	if focusedPanel, exists := m.root.Focused(); exists {
		m.board.DeselectAndRestoreIndex(focusedPanel.ID)
		cmd = m.board.Update(focusedPanel.ID, msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	const minWidth = 120
	const minHeight = 40

	if m.termWidth < minWidth || m.termHeight < minHeight {
		return fmt.Sprintf("\n\n  Terminal too small!\n\n  Current size: %dx%d\n  Minimum size: %dx%d\n\n  Please resize your terminal.\n",
			m.termWidth, m.termHeight, minWidth, minHeight)
	}

	return m.root.View(m)
}

func renderPanel(m tea.Model, panelID int, w, h int) string {
	model := m.(model)
	if panelID == panelEdit {
		model.form.WithHeight(h).WithWidth(w / 2)
		return model.form.View()
	}
	return model.board.Render(panelID, w, h)
}

func renderHelp(m tea.Model, panelID int, w, h int) string {
	model := m.(model)
	return model.help.FullHelpView(model.keys.FullHelp())
}

func renderLeftPanel(m tea.Model, panelID int, w, h int) string {
	model := m.(model)
	model.contextList.SetSize(w, h)
	return model.contextList.View()
}

func main() {
	tasksFile := flag.String("f", "tasks.json", "tasks file to use")
	flag.Parse()

	row1 := panel.New().WithId(20).WithRatio(41).WithLayout(panel.LayoutDirectionHorizontal)
	for i := range 4 {
		p := panel.New().WithId(i).WithRatio(25).WithBorder().WithContent(renderPanel)
		if i == 0 {
			p = p.Focus()
		}
		row1 = row1.Append(p)
	}

	row2 := panel.New().WithId(30).WithRatio(41).WithLayout(panel.LayoutDirectionHorizontal)
	for i := 4; i < 8; i++ {
		p := panel.New().WithId(i).WithRatio(25).WithBorder().WithContent(renderPanel)
		row2 = row2.Append(p)
	}
	editPanel := panel.New().WithId(panelEdit).WithRatio(18).WithContent(renderPanel).WithBorder().WithVisible(false).WithMaxHeight(6)
	helpPanel := panel.New().WithId(panelHelp).WithRatio(18).WithContent(renderHelp).WithBorder().WithVisible(true).WithMaxHeight(6)

	rightPanel := panel.New().WithRatio(88).WithLayout(panel.LayoutDirectionVertical).
		Append(row1).
		Append(row2).
		Append(editPanel).
		Append(helpPanel)

	leftPanel := panel.New().WithId(panelLeft).WithRatio(12).WithBorder().WithVisible(false).WithContent(renderLeftPanel)

	rootPanel := panel.New().WithRatio(100).WithLayout(panel.LayoutDirectionHorizontal).
		Append(leftPanel).
		Append(rightPanel)

	m := newModel(rootPanel, *tasksFile)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
}
