package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/joho/godotenv"
	"github.com/rwirdemann/nestiles/panel"
	"github.com/rwirdemann/scheduled"
	"github.com/rwirdemann/scheduled/board"
	clpboard "github.com/rwirdemann/scheduled/clipboard"
	"github.com/rwirdemann/scheduled/file"
	"github.com/rwirdemann/scheduled/telegram"
)

var version = "dev"

const (
	panelEdit        = 40
	panelHelp        = 50
	contextPanel     = 60
	leftPanel        = 70
	contextEditPanel = 80
	statusPanel      = 90
)

type mode int

const (
	modeNormal mode = iota
	modeEdit
	modeNew
	modeContexts
	modeSchedule
)

type clearStatusMsg struct{}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

type repository interface {
	LoadContexts() []scheduled.Context
	LoadOrder() []scheduled.TaskOrder
	SaveContexts(contexts []scheduled.Context)
	SaveOrder(orders []scheduled.TaskOrder)
}

type taskRepository interface {
	LoadTasks() []scheduled.Task
	Upsert(task scheduled.Task)
	DeleteTask(id string)
}

type model struct {
	root  panel.Model
	focus int

	board *board.Model

	taskForm         *huh.Form
	scheduleTaskForm *huh.Form
	taskRepository   taskRepository
	repository       repository

	showHelp        bool
	keys            scheduled.KeyMap
	contextViewKeys scheduled.ContextViewKeyMap
	help            help.Model

	termWidth  int
	termHeight int

	contextList      list.Model
	editContextShown bool
	contextEdit      textinput.Model
	mode             mode

	statusMessage string
	statusTimeout time.Time
	plan          *scheduled.Plan
}

func newModel(
	root panel.Model, taskRepository taskRepository, repository repository,
) model {
	h := help.New()
	h.Styles.FullKey = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	h.Styles.FullDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	h.Styles.FullSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	contextListDelegate := list.NewDefaultDelegate()
	contextListDelegate.ShowDescription = false
	contextListDelegate.SetSpacing(0)
	contexts := repository.LoadContexts()
	items := make([]list.Item, len(contexts))
	for i, v := range contexts {
		items[i] = v
	}
	contextList := list.New(items, contextListDelegate, 0, 0)
	contextList.SetShowHelp(false)
	contextList.SetShowStatusBar(false)
	contextList.Title = "Contexts"

	weekPlan := scheduled.NewPlan(taskRepository.LoadTasks(), repository.LoadOrder())

	m := model{
		root:            root,
		taskRepository:  taskRepository,
		repository:      repository,
		keys:            scheduled.Keys,
		contextViewKeys: scheduled.ContextViewKeys,
		help:            h,
		showHelp:        true,
		mode:            modeNormal,
		contextList:     contextList,
		contextEdit:     textinput.New(),
		plan:            weekPlan,
		board:           board.NewModel(weekPlan),
	}
	m.contextEdit.Placeholder = "Context"
	m.contextEdit.SetWidth(20)
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Save() {
	m.repository.SaveOrder(m.plan.AllOrders())
	m.repository.SaveContexts(m.contexts())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case scheduled.TelegramTaskMsg:
		task := m.board.AddTask(msg.Name, msg.Day)
		m.taskRepository.Upsert(task)
		dayLabel := board.Days[msg.Day]
		return m.showStatusMessage(fmt.Sprintf("Telegram: %q added to %s", msg.Name, dayLabel))
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.Save()
			return m, tea.Quit
		}
	case clearStatusMsg:
		if time.Now().After(m.statusTimeout) {
			m.statusMessage = ""
			m.root = m.root.Hide(statusPanel)
		}
		return m, nil
	}

	switch m.mode {
	case modeSchedule:
		form, cmd := m.scheduleTaskForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.scheduleTaskForm = f
			if f.State == huh.StateCompleted {
				day := m.scheduleTaskForm.GetInt("days")
				if task, ok := m.board.MoveTask(m.board.LastFocus, day); ok {
					m.taskRepository.Upsert(task)
				}
				m.mode = modeNormal
			}
			if f.State == huh.StateAborted {
				m.mode = modeNormal
			}
		}
		return m, cmd
	case modeNew, modeEdit:
		form, cmd := m.taskForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.taskForm = f
			if f.State == huh.StateCompleted {
				title := m.taskForm.GetString("title")
				context := m.taskForm.GetInt("context")
				description := m.taskForm.GetString("description")
				if m.mode == modeEdit {
					if task, ok := m.board.UpdateTask(title, context, description); ok {
						m.taskRepository.Upsert(task)
					}
				}
				if m.mode == modeNew {
					task := m.board.CreateTask(title, context, description)
					m.taskRepository.Upsert(task)
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
		if m.editContextShown {
			switch msg := msg.(type) {
			case tea.KeyPressMsg:
				switch {
				case key.Matches(msg, m.contextViewKeys.CloseView):
					m.root = m.root.Hide(contextEditPanel)
					m.editContextShown = false
					return m, nil
				case key.Matches(msg, m.keys.Enter):
					var err error
					if m, err = m.addContext(m.contextEdit.Value()); err != nil {
						return m.showStatusMessage(err.Error())
					}
					m.contextEdit.SetValue("")
					m.root = m.root.Hide(contextEditPanel)
					m.editContextShown = false
					return m, nil
				}
			}
			m.contextEdit, cmd = m.contextEdit.Update(msg)
			return m, cmd
		}
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch {
			case key.Matches(msg, m.contextViewKeys.CloseView):
				m.mode = modeNormal
				m.root = m.root.Hide(leftPanel)
				m.root = m.root.SetFocus(m.board.LastFocus)
				return m, nil
			case key.Matches(msg, m.contextViewKeys.SelectContext):
				m.mode = modeNormal
				i := m.contextList.SelectedItem()
				m.board.SetContext(i.(scheduled.Context))
				m.root = m.root.Hide(leftPanel)
				m.board.SetListTitle(board.Inbox,
					fmt.Sprintf("[ESC] Inbox (Week %d)", m.board.Week()))
				m.root = m.root.SetFocus(m.board.LastFocus)
				return m, nil
			case key.Matches(msg, m.contextViewKeys.NewContext):
				m.root = m.root.Show(contextEditPanel)
				m.root = m.root.SetFocus(contextEditPanel)
				m.editContextShown = true
				cmd = m.contextEdit.Focus()
				return m, cmd
			case key.Matches(msg, m.contextViewKeys.DeleteContext):
				var err error
				if m, err = m.deleteContext(); err != nil {
					return m.showStatusMessage(err.Error())
				}
			}
		}
		m.contextList, cmd = m.contextList.Update(msg)
		return m, cmd
	case modeNormal:
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.termWidth = msg.Width
			m.termHeight = msg.Height
		case tea.KeyPressMsg:
			switch {
			case key.Matches(msg, m.keys.ScheduleTask):
				focusedPanel, _ := m.root.Focused()
				if t, exists := m.board.GetSelectedTask(focusedPanel.ID); exists {
					m.scheduleTaskForm = scheduled.CreateScheduleTaskForm(&t, board.Days, m.board.LastFocus)
					m.mode = modeSchedule
					return m, m.scheduleTaskForm.Init()
				}
				return m, nil
			case key.Matches(msg, m.keys.Help):
				m.root = m.root.Hide(panelEdit)
				m.showHelp = !m.showHelp
				if m.showHelp {
					m.root = m.root.Show(panelHelp)
				} else {
					m.root = m.root.Hide(panelHelp)
				}
				return m, nil
			case key.Matches(msg, m.keys.Right):
				m.board.IncWeek()
				return m, nil
			case key.Matches(msg, m.keys.Left):
				m.board.DecWeek()
				return m, nil
			case key.Matches(msg, m.keys.ShiftLeft):
				if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
					if task, ok := m.board.MoveTask(focusedPanel.ID, focusedPanel.ID-1); ok {
						m.taskRepository.Upsert(task)
					}
				}
			case key.Matches(msg, m.keys.ShiftRight):
				if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
					if task, ok := m.board.MoveTask(focusedPanel.ID, focusedPanel.ID+1); ok {
						m.taskRepository.Upsert(task)
					}
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
				m.taskForm = scheduled.CreateTaskForm(prefilledTask, scheduled.LayoutVertical, m.contexts())
				m.mode = modeNew
				return m, m.taskForm.Init()
			case key.Matches(msg, m.keys.Esc):
				m.root = m.root.Hide(panelEdit)
				m.root = m.root.SetFocus(board.Inbox)
				m.board.DeselectAndRestoreIndex(board.Inbox)
				return m, nil
			case key.Matches(msg, m.keys.Space):
				if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
					if task, ok := m.board.ToggleDone(focusedPanel.ID); ok {
						m.taskRepository.Upsert(task)
					}
				}
				return m, nil
			case key.Matches(msg, m.keys.Back):
				if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
					if task, ok := m.board.GetSelectedTask(focusedPanel.ID); ok {
						m.board.DeleteTask(focusedPanel.ID)
						m.taskRepository.DeleteTask(task.ID)
					}
				}
			case key.Matches(msg, m.keys.Enter):
				focusedPanel, _ := m.root.Focused()
				if t, exists := m.board.GetSelectedTask(focusedPanel.ID); exists {
					m.taskForm = scheduled.CreateTaskForm(&t, scheduled.LayoutVertical, m.contexts())
					m.mode = modeEdit
					return m, m.taskForm.Init()
				}
			case key.Matches(msg, m.keys.Num):
				panelNum, _ := strconv.Atoi(msg.String())
				m.root = m.root.SetFocus(panelNum)
				m.board.DeselectAndRestoreIndex(panelNum)
				return m, nil
			case key.Matches(msg, m.keys.MoveToToday):
				today := time.Now().Weekday()
				if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
					if task, ok := m.board.MoveTask(focusedPanel.ID, int(today)); ok {
						m.taskRepository.Upsert(task)
					}
				}
			case key.Matches(msg, m.keys.MoveToInbox):
				if focusedPanel, _ := m.root.Focused(); focusedPanel.ID != panelEdit {
					if task, ok := m.board.MoveTask(focusedPanel.ID, board.Inbox); ok {
						m.taskRepository.Upsert(task)
					}
				}
			case key.Matches(msg, m.keys.Contexts):
				m.mode = modeContexts
				m.root = m.root.Show(leftPanel)
				m.root.SetFocus(leftPanel)
				return m, nil
			case key.Matches(msg, m.keys.CopyTasks):
				focusedPanel, _ := m.root.Focused()
				if focusedPanel.ID != panelEdit {
					tasks := m.board.GetTasksForPanel(focusedPanel.ID)
					clipboardText := clpboard.FormatTasks(m.contexts(), tasks)
					_ = clipboard.WriteAll(clipboardText)
					return m.showStatusMessage(fmt.Sprintf("%d tasks copied to clipboard", len(tasks)))
				}
				return m, nil
			}
		}
	}

	m.root, cmd = m.root.Update(msg)
	cmds = append(cmds, cmd)

	// Find focused panel and Update() its task list
	if focusedPanel, exists := m.root.Focused(); exists {
		m.board.DeselectAndRestoreIndex(focusedPanel.ID)
		cmd = m.board.Update(focusedPanel.ID, msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) deleteContext() (model, error) {
	items := m.contextList.Items()
	if len(items) == 0 {
		return m, nil
	}

	selected := m.contextList.SelectedItem()
	c := selected.(scheduled.Context)
	if c.ID == scheduled.ContextNone.ID {
		return m, fmt.Errorf("Context '%s' can not be deleted", scheduled.ContextNone.Name)
	}
	if m.board.IsContextUsed(c) {
		return m, fmt.Errorf("Context '%s' is beeing used", c.Name)
	}

	i := m.contextList.Index()
	items = append(items[:i], items[i+1:]...)
	m.contextList.SetItems(items)
	return m, nil
}

func (m model) showStatusMessage(s string) (model, tea.Cmd) {
	m.statusMessage = s
	m.statusTimeout = time.Now().Add(2 * time.Second)
	m.root = m.root.Show(statusPanel)
	return m, clearStatusAfter(2 * time.Second)
}

func (m model) addContext(name string) (model, error) {
	if name == "" {
		return m, errors.New("Context must not be empty")
	}
	maxID := 1
	for _, c := range m.contexts() {
		if strings.EqualFold(c.Name, name) {
			return m, fmt.Errorf("Context '%s' does already exist", name)
		}
		if c.ID > maxID {
			maxID = c.ID
		}
	}

	c := scheduled.Context{ID: maxID + 1, Name: name}
	m.contextList.InsertItem(len(m.contextList.Items()), c)
	return m, nil
}

func (m model) View() tea.View {
	const minWidth = 136
	const minHeight = 40

	if m.termWidth < minWidth || m.termHeight < minHeight {
		v := tea.NewView(fmt.Sprintf("\n\n  Terminal too small!\n\n  Current size: %dx%d\n  Minimum size: %dx%d\n\n  Please resize your terminal.\n",
			m.termWidth, m.termHeight, minWidth, minHeight))
		v.AltScreen = true
		return v
	}

	v := tea.NewView(m.root.View(m))
	v.AltScreen = true
	return v
}

func renderPanel(m tea.Model, panelID int, w, h int) string {
	model := m.(model)

	// Render task form in currently active list panel
	if (model.mode == modeEdit || model.mode == modeNew) && panelID == model.board.LastFocus {
		model.taskForm.WithHeight(h).WithWidth(w)
		return model.taskForm.View()
	}

	if model.mode == modeSchedule && panelID == model.board.LastFocus {
		model.scheduleTaskForm.WithHeight(h - 1).WithWidth(w)
		return model.scheduleTaskForm.View()
	}
	return model.board.Render(panelID, w, h)
}

func renderHelp(m tea.Model, _ int, _, _ int) string {
	model := m.(model)
	if model.mode == modeContexts {
		return model.help.ShortHelpView(model.keys.ShortHelp())
	}
	return model.help.FullHelpView(model.keys.FullHelp())
}

func renderContextPanel(m tea.Model, _ int, w, h int) string {
	model := m.(model)
	model.contextList.SetSize(w, h-4)
	fhv := model.help.FullHelpView(model.contextViewKeys.FullHelp())
	return model.contextList.View() + "\n" + fhv
}

func renderContextEditPanel(m tea.Model, _ int, _, _ int) string {
	model := m.(model)
	return model.contextEdit.View()
}

func renderStatus(m tea.Model, _ int, _, _ int) string {
	model := m.(model)
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("42")).
		Bold(true).
		Padding(0, 1)
	return statusStyle.Render(model.statusMessage)
}

func (m model) contexts() []scheduled.Context {
	items := m.contextList.Items()
	var contexts []scheduled.Context
	for _, i := range items {
		contexts = append(contexts, i.(scheduled.Context))
	}
	return contexts
}

func createModel(taskRepository taskRepository, repository repository) model {
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
	statusPanel := panel.New().WithId(statusPanel).WithRatio(18).WithContent(renderStatus).WithBorder().WithVisible(false).WithMaxHeight(3)
	editPanel := panel.New().WithId(panelEdit).WithRatio(18).WithContent(renderPanel).WithBorder().WithVisible(false).WithMaxHeight(6)
	helpPanel := panel.New().WithId(panelHelp).WithRatio(18).WithContent(renderHelp).WithBorder().WithVisible(true).WithMaxHeight(6)

	rightPanel := panel.New().WithRatio(84).WithLayout(panel.LayoutDirectionVertical).
		Append(statusPanel).
		Append(row1).
		Append(row2).
		Append(editPanel).
		Append(helpPanel)

	leftPanel := panel.New().WithId(leftPanel).WithRatio(16).WithVisible(false).WithLayout(panel.LayoutDirectionVertical)
	contextPanel := panel.New().WithId(contextPanel).WithRatio(82).WithBorder().WithContent(renderContextPanel)
	contextEditPanel := panel.New().WithId(contextEditPanel).WithRatio(18).WithBorder().WithVisible(false).WithContent(renderContextEditPanel).WithMaxHeight(6)
	leftPanel = leftPanel.
		Append(contextPanel).
		Append(contextEditPanel)

	rootPanel := panel.New().WithRatio(100).WithLayout(panel.LayoutDirectionHorizontal).
		Append(leftPanel).
		Append(rightPanel)

	return newModel(rootPanel, taskRepository, repository)
}

func main() {
	_ = godotenv.Load(filepath.Join(os.Getenv("HOME"), ".scheduled", ".env"))

	tasksFile := flag.String("f", "tasks.json", "tasks file to use")
	showVersion := flag.Bool("version", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	repo := file.NewRepository(*tasksFile)
	taskRepo := file.NewTaskRepository(*tasksFile)
	m := createModel(taskRepo, repo)

	p := tea.NewProgram(m)

	if poller := telegram.NewPoller(p.Send); poller != nil {
		poller.Start()
	}

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
	if fm, ok := finalModel.(model); ok {
		fm.Save()
	}
}
