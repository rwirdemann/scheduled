package scheduled

import (
	"errors"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
)

func CreateTaskForm(task *Task, contexts []Context) *huh.Form {
	titleInput := huh.NewInput().
		Title("Title").
		Key("title").
		Validate(func(str string) error {
			if str == "" {
				return errors.New("please enter a title")
			}
			return nil
		})

	var options []huh.Option[int]
	for _, c := range contexts {
		options = append(options, huh.NewOption(c.Name, c.ID))
	}

	contextSelect := huh.NewSelect[int]().
		Title("Context").
		Key("context").
		Options(options...)
	if task != nil {
		titleInput = titleInput.Value(&task.Name)
		contextSelect = contextSelect.Value(&task.Context)
	}

	k := huh.NewDefaultKeyMap()
	k.Quit = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Cancel"))
	return huh.NewForm(huh.NewGroup(titleInput), huh.NewGroup(contextSelect)).
		WithLayout(huh.LayoutGrid(1, 2)).WithKeyMap(k)
}
