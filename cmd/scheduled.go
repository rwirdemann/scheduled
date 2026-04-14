package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	contextPanel     = 60
	leftPanel        = 70
	contextEditPanel = 80
	statusPanel      = 90
	footerPanel      = 100
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

type systemThemeMsg bool

// pollSystemTheme returns a command that waits 5 seconds, then queries
// the OS system appearance and returns a systemThemeMsg.
func pollSystemTheme() tea.Cmd {
	return tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
		return systemThemeMsg(isDarkMode())
	})
}

// isDarkMode reports whether the OS is currently in dark mode.
func isDarkMode() bool {
	switch runtime.GOOS {
	case "darwin":
		out, _ := exec.Command(
			"defaults", "read", "-g", "AppleInterfaceStyle",
		).Output()
		return strings.TrimSpace(string(out)) == "Dark"
	case "linux":
		out, _ := exec.Command(
			"gsettings", "get",
			"org.gnome.desktop.interface", "color-scheme",
		).Output()
		return strings.Contains(string(out), "prefer-dark")
	case "windows":
		out, _ := exec.Command(
			"reg", "query",
			`HKCU\Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`,
			"/v", "AppsUseLightTheme",
		).Output()
		return strings.Contains(string(out), "0x0")
	default:
		return false
	}
}

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

	isDark bool
	theme  scheduled.Theme
}

func newModel(
	root panel.Model, taskRepository taskRepository, repository repository,
) model {
	isDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	theme := scheduled.NewTheme(isDark)

	h := help.New()
	h.Styles.FullKey = theme.HelpKey
	h.Styles.FullDesc = theme.HelpDesc
	h.Styles.FullSeparator = theme.HelpSep

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
		showHelp:        false,
		mode:            modeNormal,
		contextList:     contextList,
		contextEdit:     textinput.New(),
		plan:            weekPlan,
		board:           board.NewModel(weekPlan),
		isDark:          isDark,
		theme:           theme,
	}
	m.contextEdit.Placeholder = "Context"
	m.contextEdit.SetWidth(20)
	m.board.SetTheme(isDark)
	return m
}

func (m model) Init() tea.Cmd {
	return pollSystemTheme()
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
	case systemThemeMsg:
		if bool(msg) != m.isDark {
			m.isDark = bool(msg)
			m.theme = scheduled.NewTheme(m.isDark)
			m.applyTheme()
		}
		return m, pollSystemTheme()
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
				m.root = m.root.SetFocus(m.board.LastFocus)
				m.mode = modeNormal
			}
			if f.State == huh.StateAborted {
				m.root = m.root.Hide(panelEdit)
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
				m.showHelp = !m.showHelp
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
				m.taskForm = scheduled.CreateTaskForm(
					prefilledTask,
					scheduled.LayoutVertical,
					m.contexts(),
				)
				m.mode = modeNew
				return m, m.taskForm.Init()
			case key.Matches(msg, m.keys.Esc):
				if m.showHelp {
					m.showHelp = false
					return m, nil
				}
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
					m.taskForm = scheduled.CreateTaskForm(
						&t,
						scheduled.LayoutVertical,
						m.contexts(),
					)
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
			case key.Matches(msg, m.keys.ToggleTheme):
				m.isDark = !m.isDark
				m.theme = scheduled.NewTheme(m.isDark)
				m.applyTheme()
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

// View renders the current UI frame. It returns an error screen when the
// terminal is too small, a composited help overlay when showHelp is true,
// or the plain panel layout otherwise. All views run in the alternate screen
// buffer to preserve the shell's scroll history.
func (m model) View() tea.View {
	const minWidth = 136
	const minHeight = 40

	if m.termWidth < minWidth || m.termHeight < minHeight {
		v := tea.NewView(fmt.Sprintf(
			"\n\n  Terminal too small!\n\n"+
				"  Current size: %dx%d\n"+
				"  Minimum size: %dx%d\n\n"+
				"  Please resize your terminal.\n",
			m.termWidth, m.termHeight, minWidth, minHeight,
		))
		v.AltScreen = true
		return v
	}

	bg := m.root.View(m)
	if m.showHelp {
		box := scheduled.NewHelpOverlay(m.theme, m.help.Styles).
			Render(m.keys.FullHelp())
		boxW := lipgloss.Width(box)
		boxH := lipgloss.Height(box)
		x := (m.termWidth - boxW) / 2
		y := (m.termHeight - boxH) / 2
		c := lipgloss.NewCompositor(
			lipgloss.NewLayer(bg),
			lipgloss.NewLayer(box).X(x).Y(y).Z(1),
		)
		v := tea.NewView(c.Render())
		v.AltScreen = true
		return v
	}

	v := tea.NewView(bg)
	v.AltScreen = true
	return v
}

func renderPanel(m tea.Model, panelID int, w, h int) string {
	model := m.(model)

	// Render task form in currently active list panel
	if (model.mode == modeEdit || model.mode == modeNew) &&
		panelID == model.board.LastFocus {
		model.taskForm.WithHeight(h).WithWidth(w)
		return model.taskForm.View()
	}

	if model.mode == modeSchedule && panelID == model.board.LastFocus {
		model.scheduleTaskForm.WithHeight(h - 1).WithWidth(w)
		return model.scheduleTaskForm.View()
	}
	return model.board.Render(panelID, w, h)
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
	return model.theme.Status.Render(model.statusMessage)
}

func renderFooter(m tea.Model, _ int, w, _ int) string {
	model := m.(model)
	k := model.help.Styles.FullKey
	d := model.help.Styles.FullDesc
	sep := model.help.Styles.FullSeparator
	entries := []key.Binding{
		model.keys.New,
		model.keys.Help,
		model.keys.Quit,
	}
	var parts []string
	for _, b := range entries {
		h := b.Help()
		parts = append(parts, k.Render(h.Key)+" "+d.Render(h.Desc))
	}
	left := "  " + strings.Join(parts, "  "+sep.Render("·")+"  ")
	ctx := model.board.GetSelectedContext()
	right := ""
	if ctx.ID != scheduled.ContextNone.ID {
		right = d.Render(ctx.Name) + "  " + sep.Render("·") + "  "
	}
	right += "Scheduled - " + d.Render(version) + " "
	rightWidth := lipgloss.Width(right)
	leftWidth := lipgloss.Width(left)
	padding := w - leftWidth - rightWidth
	if padding < 1 {
		padding = 1
	}
	return left + strings.Repeat(" ", padding) + right
}

// applyTheme updates help styles and board theme to match m.theme.
func (m *model) applyTheme() {
	m.help.Styles.FullKey = m.theme.HelpKey
	m.help.Styles.FullDesc = m.theme.HelpDesc
	m.help.Styles.FullSeparator = m.theme.HelpSep
	m.board.SetTheme(m.isDark)
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
	// The root panel contains a left and a right panel, horizontally arranged.
	// The right panel contains the main elemenets, i.e. the day lists, the
	// help and the status bar. The left panel is for contexts.
	rightPanel := panel.New().
		WithRatio(84).
		WithLayout(panel.LayoutDirectionVertical)
	row1 := panel.New().
		WithId(20).
		WithRatio(41).
		WithLayout(panel.LayoutDirectionHorizontal)
	for i := range 4 {
		p := panel.New().
			WithId(i).
			WithRatio(25).
			WithBorder().
			WithContent(renderPanel)
		if i == 0 {
			p = p.Focus()
		}
		row1 = row1.Append(p)
	}
	row2 := panel.New().
		WithId(30).
		WithRatio(41).
		WithLayout(panel.LayoutDirectionHorizontal)
	for i := 4; i < 8; i++ {
		p := panel.New().
			WithId(i).
			WithRatio(25).
			WithBorder().
			WithContent(renderPanel)
		row2 = row2.Append(p)
	}
	statusPanel := panel.New().
		WithId(statusPanel).
		WithRatio(18).
		WithContent(renderStatus).
		WithBorder().WithVisible(false).
		WithMaxHeight(3)
	footerPanel := panel.New().
		WithId(footerPanel).
		WithRatio(18).
		WithContent(renderFooter).
		WithMaxHeight(1)
	rightPanel = rightPanel.
		Append(statusPanel).
		Append(row1).
		Append(row2).
		Append(footerPanel)

	// the left panel
	leftPanel := panel.New().
		WithId(leftPanel).
		WithRatio(16).
		WithVisible(false).
		WithLayout(panel.LayoutDirectionVertical)
	contextPanel := panel.New().
		WithId(contextPanel).
		WithRatio(82).
		WithBorder().
		WithContent(renderContextPanel)
	contextEditPanel := panel.New().
		WithId(contextEditPanel).
		WithRatio(18).
		WithBorder().
		WithVisible(false).
		WithContent(renderContextEditPanel).
		WithMaxHeight(6)
	leftPanel = leftPanel.
		Append(contextPanel).
		Append(contextEditPanel)

	// finally, the root panel
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
