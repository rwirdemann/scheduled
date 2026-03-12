package scheduled

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
