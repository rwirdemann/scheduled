package scheduled

import (
	"errors"
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
)

type Layout int

const (
	LayoutHorizontal Layout = iota
	LayoutVertical
)

// CreateTaskForm creates a form to create new or edit existings tasks.
func CreateTaskForm(task *Task, layout Layout, contexts []Context) *huh.Form {
	titleInput := huh.NewText().
		Title("Title").
		Key("title").
		Lines(2).
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

	descText := huh.NewText().
		Title("Description").
		Placeholder("alt-enter - newline\nenter     - submit").
		Key("description")

	contextSelect := huh.NewSelect[int]().
		Title("Context").
		Key("context").
		Options(options...)
	if task != nil {
		titleInput = titleInput.Value(&task.Name)
		contextSelect = contextSelect.Value(&task.Context)
		descText = descText.Value(&task.Desc)
	}

	// Height needs to be set after value assignment
	contextSelect.Height(1)

	k := huh.NewDefaultKeyMap()
	k.Quit = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Cancel"))

	var withLayout huh.Layout
	switch layout {
	case LayoutHorizontal:
		withLayout = huh.LayoutGrid(1, 3)
	case LayoutVertical:
		withLayout = huh.LayoutStack
	}

	return huh.NewForm(
		huh.NewGroup(titleInput),
		huh.NewGroup(contextSelect),
		huh.NewGroup(descText)).
		WithLayout(withLayout).WithKeyMap(k).WithShowHelp(false)
}

func CreateScheduleTaskForm(task *Task, days map[int]string) *huh.Form {
	var options []huh.Option[int]
	for k, v := range days {
		options = append(options, huh.NewOption(v, k))
	}

	s := huh.NewSelect[int]().Title(fmt.Sprintf("Move tasks '%s' to", task.Name)).Key("days").Options(options...)
	s = s.Value(&task.Day)

	k := huh.NewDefaultKeyMap()
	k.Quit = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Cancel"))
	return huh.NewForm(huh.NewGroup(s)).WithLayout(huh.LayoutGrid(1, 2)).WithKeyMap(k)
}
