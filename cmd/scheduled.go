package main

import (
	"fmt"
	"os"
	"sort"
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
	panelLeft = 60
)

type mode int

const (
	modeNormal mode = iota
	modeEdit
	modeNew
	modeContexts
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

type taskRepository interface {
	Load() []scheduled.Task
	Save(tasks []scheduled.Task)
}

type model struct {
	root  panel.Model
	focus int
	week  int

	board *board.Model

	form           *huh.Form
	taskRepository taskRepository
	lastFocus      int

	showHelp bool
	keys     scheduled.KeyMap
	help     help.Model

	termWidth  int
	termHeight int

	contextList     list.Model
	selectedContext scheduled.Context

	mode mode
}

func newModel(root panel.Model) model {
	h := help.New()
	h.Styles.FullKey = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	h.Styles.FullDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	h.Styles.FullSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	contextListDelegate := list.NewDefaultDelegate()
	contextListDelegate.ShowDescription = false
	contextListDelegate.SetSpacing(0)
	contextList := list.New([]list.Item{scheduled.ContextNone, scheduled.ContextPrivate, scheduled.ContextiNeonpulse}, contextListDelegate, 0, 0)
	contextList.SetShowHelp(false)
	contextList.SetShowStatusBar(false)
	contextList.Title = "Contexts"

	m := model{
		root:            root,
		taskRepository:  file.Repository{},
		lastFocus:       Inbox,
		keys:            scheduled.Keys,
		help:            h,
		showHelp:        true,
		mode:            modeNormal,
		contextList:     contextList,
		selectedContext: scheduled.ContextNone,
		board:           board.NewModel(file.Repository{}),
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
					m.board.UpdateTask(m.lastFocus, title, context)
				}
				if m.mode == modeNew {
					m.board.CreateTask(m.lastFocus, title, context)
				}
				m.root = m.root.Hide(panelEdit)
				if m.showHelp {
					m.root = m.root.Show(panelHelp)
				}
				m.root = m.root.SetFocus(m.lastFocus)
				m.mode = modeNormal
			}
			if f.State == huh.StateAborted {
				m.root = m.root.Hide(panelEdit)
				if m.showHelp {
					m.root = m.root.Show(panelHelp)
				}
				m.root = m.root.SetFocus(m.lastFocus)
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
				m.root = m.root.SetFocus(m.lastFocus)
				return m, nil
			case key.Matches(msg, m.keys.Enter):
				m.mode = modeNormal
				i := m.contextList.SelectedItem()
				m.selectedContext = i.(scheduled.Context)
				m.root = m.root.Hide(panelLeft)
				m.board.SetListTitle(board.Inbox, fmt.Sprintf("[ESC] Inbox (Week %d) - %s", m.week, m.selectedContext.Name))
				m.root = m.root.SetFocus(m.lastFocus)
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
			m.saveTasks()
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
			m.board.MoveUp(focusedPanel.ID)
		case key.Matches(msg, m.keys.ShiftDown):
			focusedPanel, _ := m.root.Focused()
			m.board.MoveDown(focusedPanel.ID)
		case key.Matches(msg, m.keys.New):
			m.form = scheduled.CreateTaskForm(nil)
			m.root = m.root.Hide(panelHelp)
			m.root = m.root.Show(panelEdit)
			m.root = m.root.SetFocus(panelEdit)
			m.mode = modeNew
			return m, m.form.Init()
		case key.Matches(msg, m.keys.Esc):
			m.root = m.root.Hide(panelEdit)
			m.root = m.root.SetFocus(Inbox)
			m = m.saveAndRestoreSelection(Inbox)
			return m, nil
		case key.Matches(msg, m.keys.Space):
			if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
				m.board.ToggleDone(focusedPanel.ID)
			}
		case key.Matches(msg, m.keys.Back):
			if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
				l := m.board.Lists[focusedPanel.ID]
				selected := l.SelectedItem()
				t := selected.(scheduled.Task)
				if t.Done {
					l.RemoveItem(l.Index())
				}
			}
		case key.Matches(msg, m.keys.Enter):
			focusedPanel, _ := m.root.Focused()
			l := m.board.Lists[focusedPanel.ID]
			if selected := l.SelectedItem(); selected != nil {
				t := selected.(scheduled.Task)
				m.form = scheduled.CreateTaskForm(&t)
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
			m = m.saveAndRestoreSelection(panelNum)
			return m, nil
		case key.Matches(msg, m.keys.MoveToToday):
			today := time.Now().Weekday()
			if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
				m = m.moveTask(focusedPanel.ID, int(today))
			}
		case key.Matches(msg, m.keys.MoveToInbox):
			if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
				m = m.moveTask(focusedPanel.ID, Inbox)
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
		m = m.saveAndRestoreSelection(focusedPanel.ID)
		if list, exists := m.board.Lists[focusedPanel.ID]; exists {
			list.Model, cmd = list.Model.Update(msg)
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) saveAndRestoreSelection(focusedPanelID int) model {
	if m.lastFocus != focusedPanelID && focusedPanelID != panelEdit {
		// Save index and deselect previously focused list
		if prevList, exists := m.board.Lists[m.lastFocus]; exists {
			prevList.SaveIndex()
			prevList.Deselect()
		}
		// Restore index of newly focused list
		if currList, exists := m.board.Lists[focusedPanelID]; exists {
			currList.RestoreIndex()
		}
		m.lastFocus = focusedPanelID
	}
	return m
}

func (m model) moveTask(from, to int) model {
	if _, exists := m.board.Lists[to]; !exists {
		to = 0
	}
	if from == to {
		return m
	}

	if item := m.board.Lists[from].SelectedItem(); item != nil {
		t := item.(scheduled.Task)
		t.Day = to
		m.board.Lists[from].RemoveItem(m.board.Lists[from].Index())
		m.board.Lists[to].InsertItem(len(m.board.Lists[to].Items()), t)
	}
	return m
}

func (m model) loadTasks() {
	var tasksByDay = make(map[int][]list.Item)
	tasks := m.taskRepository.Load()
	for _, task := range tasks {
		tasksByDay[task.Day] = append(tasksByDay[task.Day], task)
	}

	// Sort tasks by their Pos field
	for _, items := range tasksByDay {
		sort.Slice(items, func(i, j int) bool {
			return items[i].(scheduled.Task).Pos < items[j].(scheduled.Task).Pos
		})
	}

	for day := range m.board.Lists {
		for i, item := range tasksByDay[day] {
			m.board.Lists[day].InsertItem(i, item)
		}
	}
}

func (m model) saveTasks() {
	var tasks []scheduled.Task
	for _, list := range m.board.Lists {
		for i, item := range list.Items() {
			t := item.(scheduled.Task)
			t.Pos = i
			tasks = append(tasks, t)
		}
	}
	m.taskRepository.Save(tasks)
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

func (m model) setWeek(week int) model {
	m.week = week
	for i := Inbox; i <= Sunday; i++ {
		monday := date.GetMondayOfWeek(m.week)
		if i == Inbox {
			m.board.Lists[i].Title = fmt.Sprintf("[ESC] Inbox (Week %d) - %s", m.week, m.selectedContext.Name)
		} else {
			day := monday.AddDate(0, 0, i-1)
			m.board.Lists[i].Title = fmt.Sprintf("[%d] %s (%s)", i, days[i], day.Format("02.01.2006"))
		}
	}
	return m
}

func renderPanel(m tea.Model, panelID int, w, h int) string {
	model := m.(model)
	if panelID == panelEdit {
		model.form.WithHeight(h).WithWidth(w / 2)
		return model.form.View()
	}
	if list, exists := model.board.Lists[panelID]; exists {
		list.Model.SetSize(w, h)
		return list.Model.View()
	}
	return ""
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
	row1 := panel.New().WithId(20).WithRatio(44).WithLayout(panel.LayoutDirectionHorizontal)
	for i := range 4 {
		p := panel.New().WithId(i).WithRatio(25).WithBorder().WithContent(renderPanel)
		if i == 0 {
			p = p.Focus()
		}
		row1 = row1.Append(p)
	}

	row2 := panel.New().WithId(30).WithRatio(44).WithLayout(panel.LayoutDirectionHorizontal)
	for i := 4; i < 8; i++ {
		p := panel.New().WithId(i).WithRatio(25).WithBorder().WithContent(renderPanel)
		row2 = row2.Append(p)
	}
	editPanel := panel.New().WithId(panelEdit).WithRatio(12).WithContent(renderPanel).WithBorder().WithVisible(false)
	helpPanel := panel.New().WithId(panelHelp).WithRatio(12).WithContent(renderHelp).WithBorder().WithVisible(true)

	rightPanel := panel.New().WithRatio(90).WithLayout(panel.LayoutDirectionVertical).
		Append(row1).
		Append(row2).
		Append(editPanel).
		Append(helpPanel)

	leftPanel := panel.New().WithId(panelLeft).WithRatio(10).WithBorder().WithVisible(false).WithContent(renderLeftPanel)

	rootPanel := panel.New().WithRatio(100).WithLayout(panel.LayoutDirectionHorizontal).
		Append(leftPanel).
		Append(rightPanel)

	m := newModel(rootPanel)
	defaultDelegate := list.NewDefaultDelegate()
	defaultDelegate.ShowDescription = false
	defaultDelegate.SetSpacing(0)
	for i := Inbox; i <= Sunday; i++ {
		l := list.New([]list.Item{}, defaultDelegate, 0, 0)
		l.SetShowStatusBar(false)
		l.SetShowHelp(false)
		m.board.Lists[i] = board.NewListModel(l)
	}
	_, w := time.Now().ISOWeek()
	m = m.setWeek(w)
	m.loadTasks()

	// Deselect all lists except the focused one (Inbox)
	for i := Inbox; i <= Sunday; i++ {
		if i != Inbox {
			m.board.Lists[i].Deselect()
		}
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
}
