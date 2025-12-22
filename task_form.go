package scheduled

import (
	"errors"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
)

func CreateTaskForm(task *Task) *huh.Form {
	titleInput := huh.NewInput().
		Title("Title").
		Key("title").
		Validate(func(str string) error {
			if str == "" {
				return errors.New("Please enter a title")
			}
			return nil
		})
	contextSelect := huh.NewSelect[int]().
		Title("Context").
		Key("context").
		Options(
			huh.NewOption(ContextNone.Name, ContextNone.ID),
			huh.NewOption(ContextPrivate.Name, ContextPrivate.ID),
			huh.NewOption(ContextiNeonpulse.Name, ContextiNeonpulse.ID),
		)
	if task != nil {
		titleInput = titleInput.Value(&task.Name)
		contextSelect = contextSelect.Value(&task.Context)
	}

	k := huh.NewDefaultKeyMap()
	k.Quit = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Cancel"))
	return huh.NewForm(huh.NewGroup(titleInput), huh.NewGroup(contextSelect)).
		WithLayout(huh.LayoutGrid(1, 2)).WithKeyMap(k)
}
