package board

import (
	"github.com/rwirdemann/scheduled"
)

// ListItem wraps a Task and implements the list.Item interface.
type ListItem struct {
	Task scheduled.Task
}

func (i ListItem) Title() string {
	if i.Task.Done {
		return DoneStyle.Render("✓ " + i.Task.Name)
	}
	return "○ " + i.Task.Name
}

func (i ListItem) Description() string { return "" }
func (i ListItem) FilterValue() string { return i.Task.Name }
