package scheduled

import "fmt"

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
		return "\x1b[90m✓ " + i.Name + "\x1b[0m"
	}
	return fmt.Sprintf("%s%s (%d)", checkbox, i.Name, i.Context)
}

func (i Task) Description() string { return i.Desc }
func (i Task) FilterValue() string { return i.Name }
