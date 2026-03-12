package board

import (
	"fmt"

	"github.com/rwirdemann/scheduled"
)

// ListItem wraps a Task and implements the list.Item interface.
type ListItem struct {
	Task scheduled.Task
}

func (i ListItem) Title() string {
	checkbox := "○ "
	if i.Task.Done {
		return "\x1b[90m✓ " + fmt.Sprintf("%s", i.Task.Name+"\x1b[0m")
	}
	return fmt.Sprintf("%s%s", checkbox, i.Task.Name)
}

func (i ListItem) Description() string { return "" }
func (i ListItem) FilterValue() string { return i.Task.Name }
