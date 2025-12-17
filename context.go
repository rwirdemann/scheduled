package scheduled

type Context struct {
	Name string `json:"name"`
}

func (c Context) Title() string {
	return c.Name
}

func (c Context) Description() string { return c.Name }
func (c Context) FilterValue() string { return c.Name }
