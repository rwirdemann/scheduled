package scheduled

type Task struct {
	Name string `json:"name"`
}

func (t Task) FilterValue() string {
	return ""
}
