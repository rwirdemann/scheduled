package scheduled

var (
	ContextNone       = Context{ID: 1, Name: "all"}
	ContextPrivate    = Context{ID: 2, Name: "private"}
	ContextiNeonpulse = Context{ID: 3, Name: "neonpulse"}
)

type Context struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (c Context) Title() string {
	return c.Name
}

func (c Context) Description() string { return c.Name }
func (c Context) FilterValue() string { return c.Name }
