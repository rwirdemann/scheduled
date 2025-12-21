package main

import (
	"errors"
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

type repository interface {
	Load() []scheduled.Task
	Save(tasks []scheduled.Task)
}

type model struct {
	root  panel.Model
	focus int
	week  int

	lists      map[int]*scheduled.ListModel
	form       *huh.Form
	repository repository
	lastFocus  int

	showHelp bool
	keys     scheduled.KeyMap
	help     help.Model

	termWidth  int
	termHeight int

	contextList list.Model

	mode mode
}

func (m *model) createTaskForm(task *scheduled.Task) *huh.Form {
	var titleInput *huh.Input
	if task != nil {
		titleInput = huh.NewInput().
			Title("Title").
			Key("title").
			Value(&task.Name).
			Validate(func(str string) error {
				if str == "" {
					return errors.New("Please enter a title")
				}
				return nil
			})
	} else {
		titleInput = huh.NewInput().
			Title("Title").
			Key("title").
			Validate(func(str string) error {
				if str == "" {
					return errors.New("Please enter a title")
				}
				return nil
			})
	}

	return huh.NewForm(
		huh.NewGroup(titleInput),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Context").
				Options(
					huh.NewOption("none", "none"),
					huh.NewOption("private", "private"),
					huh.NewOption("neonpulse", "neonpulse"),
				),
		),
	).WithLayout(huh.LayoutGrid(1, 2))
}

func newModel(root panel.Model) model {
	h := help.New()
	h.Styles.FullKey = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	h.Styles.FullDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	h.Styles.FullSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	contextListDelegate := list.NewDefaultDelegate()
	contextListDelegate.ShowDescription = false
	contextListDelegate.SetSpacing(0)
	contextList := list.New([]list.Item{scheduled.Context{Name: "private"}, scheduled.Context{Name: "neonpulse"}}, contextListDelegate, 0, 0)
	contextList.SetShowHelp(false)
	contextList.SetShowStatusBar(false)
	contextList.Title = "Contexts"

	m := model{root: root, lists: make(map[int]*scheduled.ListModel),
		repository:  file.Repository{},
		lastFocus:   Inbox,
		keys:        scheduled.Keys,
		help:        h,
		showHelp:    true,
		mode:        modeNormal,
		contextList: contextList,
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
				l := m.lists[m.lastFocus]
				if m.mode == modeEdit {
					item := l.SelectedItem()
					t := item.(scheduled.Task)
					t.Name = m.form.GetString("title")
					index := l.Index()
					l.RemoveItem(index)
					l.InsertItem(index, t)
				}
				if m.mode == modeNew {
					t := scheduled.Task{Name: m.form.GetString("title"), Day: m.lastFocus}
					l.InsertItem(len(l.Items()), t)
				}
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
			case key.Matches(msg, m.keys.Contexts):
				m.mode = modeNormal
				m.root = m.root.Hide(panelLeft)
				return m, nil
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
			m.showHelp = !m.showHelp
			if m.showHelp {
				m.root = m.root.Show(panelHelp)
			} else {
				m.root = m.root.Hide(panelHelp)
			}
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
			m.form = m.createTaskForm(nil)
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
			focusedPanel, _ := m.root.Focused()
			l := m.lists[focusedPanel.ID]
			if selected := l.SelectedItem(); selected != nil {
				t := selected.(scheduled.Task)
				m.form = m.createTaskForm(&t)
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
		case key.Matches(msg, m.keys.AltT):
			today := time.Now().Weekday()
			if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
				m = m.moveTask(focusedPanel.ID, int(today))
			}
		case key.Matches(msg, m.keys.AltI):
			if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
				m = m.moveTask(focusedPanel.ID, Inbox)
			}
		case key.Matches(msg, m.keys.Contexts):
			m.mode = modeContexts
			m.root = m.root.Show(panelLeft)
			return m, nil
		}
	}
	m.root, cmd = m.root.Update(msg)
	cmds = append(cmds, cmd)

	// find focused panel and Update() its task list
	if focusedPanel, exists := m.root.Focused(); exists {
		m = m.saveAndRestoreSelection(focusedPanel.ID)
		if list, exists := m.lists[focusedPanel.ID]; exists {
			list.Model, cmd = list.Model.Update(msg)
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) saveAndRestoreSelection(focusedPanelID int) model {
	if m.lastFocus != focusedPanelID && focusedPanelID != panelEdit {
		// Save index and deselect previously focused list
		if prevList, exists := m.lists[m.lastFocus]; exists {
			prevList.SaveIndex()
			prevList.Deselect()
		}
		// Restore index of newly focused list
		if currList, exists := m.lists[focusedPanelID]; exists {
			currList.RestoreIndex()
		}
		m.lastFocus = focusedPanelID
	}
	return m
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
			m.lists[i].Title = fmt.Sprintf("[ESC] Inbox (Week %d)", m.week)
		} else {
			day := monday.AddDate(0, 0, i-1)
			m.lists[i].Title = fmt.Sprintf("[%d] %s (%s)", i, days[i], day.Format("02.01.2006"))
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

func renderLeftPanel(m tea.Model, panelID int, w, h int) string {
	model := m.(model)
	model.contextList.SetSize(w, h)
	return model.contextList.View()
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

	rightPanel := panel.New().WithRatio(90).WithLayout(panel.LayoutDirectionVertical).
		Append(row1).
		Append(row2).
		Append(row3).
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
