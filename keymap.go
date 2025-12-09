package scheduled

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	ShiftRight key.Binding
	ShiftLeft  key.Binding
	ShiftUp    key.Binding
	ShiftDown  key.Binding
	Right      key.Binding
	Left       key.Binding
	New        key.Binding
	Esc        key.Binding
	Back       key.Binding
	Space      key.Binding
	Help       key.Binding
	Enter      key.Binding
	Quit       key.Binding
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
		{k.Esc, k.New, k.Space, k.Back},
		{k.ShiftRight, k.ShiftLeft, k.ShiftDown, k.ShiftUp},
		{k.Right, k.Left},
		{k.Help, k.Quit},
	}
}

var Keys = KeyMap{
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
		key.WithHelp("enter", "submits task"),
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
		key.WithHelp("?", "show help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
}
