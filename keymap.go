package scheduled

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	NextDay     key.Binding
	PrevDay     key.Binding
	ShiftRight  key.Binding
	ShiftLeft   key.Binding
	ShiftUp     key.Binding
	ShiftDown   key.Binding
	Right       key.Binding
	Left        key.Binding
	New         key.Binding
	Esc         key.Binding
	Back        key.Binding
	Space       key.Binding
	Help        key.Binding
	Enter       key.Binding
	Quit        key.Binding
	Num         key.Binding
	MoveToToday key.Binding
	MoveToInbox key.Binding
	Contexts    key.Binding
	CopyTasks   key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k ContextViewKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.SelectContext, k.NewContext, k.DeleteContext, k.CloseView}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k ContextViewKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NewContext, k.SelectContext, k.DeleteContext, k.CloseView},
	}
}

var Keys = KeyMap{
	NextDay: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next day"),
	),
	PrevDay: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev day"),
	),
	Esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "focus inbox"),
	),
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new task"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "edit task"),
	),
	Back: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "delete task"),
	),
	Space: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "check / uncheck task"),
	),
	ShiftRight: key.NewBinding(
		key.WithKeys("shift+right"),
		key.WithHelp("shift+→", "move task right"),
	),
	ShiftLeft: key.NewBinding(
		key.WithKeys("shift+left"),
		key.WithHelp("shift+←", "move task left"),
	),
	ShiftUp: key.NewBinding(
		key.WithKeys("shift+up"),
		key.WithHelp("shift+↑", "move task up"),
	),
	ShiftDown: key.NewBinding(
		key.WithKeys("shift+down"),
		key.WithHelp("shift+↓", "move task down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "prev week"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "next week"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
	Num: key.NewBinding(
		key.WithKeys("1", "2", "3", "4", "5", "6", "7"),
		key.WithHelp("{num}", "focus day {num}"),
	),
	MoveToToday: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "move task to today"),
	),
	MoveToInbox: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "move task inbox"),
	),
	Contexts: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "show contexts"),
	),
	CopyTasks: key.NewBinding(
		key.WithKeys("k"),
		key.WithHelp("k", "copy tasks"),
	),
}

type ContextViewKeyMap struct {
	NewContext    key.Binding
	SelectContext key.Binding
	DeleteContext key.Binding
	CloseView     key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ShiftRight, k.ShiftLeft, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.New, k.Enter, k.Space, k.Back},
		{k.NextDay, k.PrevDay, k.Right, k.Left},
		{k.ShiftRight, k.ShiftLeft, k.ShiftDown, k.ShiftUp},
		{k.Num, k.MoveToToday, k.MoveToInbox, k.Esc},
		{k.Help, k.Contexts, k.CopyTasks, k.Quit},
	}
}

var ContextViewKeys = ContextViewKeyMap{
	SelectContext: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	NewContext: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new context"),
	),
	DeleteContext: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "del context"),
	),
	CloseView: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "close view"),
	),
}
