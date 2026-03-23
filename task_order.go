package scheduled

// TaskOrder holds the display position of a task within its day.
type TaskOrder struct {
	TaskID string `json:"task_id"`
	Pos    int    `json:"pos"`
}
