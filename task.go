package scheduled

import "fmt"

// Task represents a task in the task list.
type Task struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Desc    string `json:"description"`
	Day     int    `json:"day"`
	Done    bool   `json:"done"`
	Pos     int    `json:"pos"`
	Context int    `json:"context"`
}

func (i Task) Title() string {
	checkbox := "○ "
	if i.Done {
		// Gray color using ANSI escape code
		return "\x1b[90m✓ " + fmt.Sprintf("%s", i.Name+"\x1b[0m")
	}
	return fmt.Sprintf("%s%s", checkbox, i.Name)
}

func (i Task) Description() string { return "hello" }
func (i Task) FilterValue() string { return i.Name }
