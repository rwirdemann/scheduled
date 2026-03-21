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
		Placeholder("enter - next field").
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
		Placeholder("alt+enter - newline").
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

// CreateScheduleTaskForm creates a form to move a task to a different day.
// The current day is excluded from the list of options.
func CreateScheduleTaskForm(task *Task, days map[int]string, currentDay int) *huh.Form {
	var options []huh.Option[int]
	// Monday(1)..Sunday(7), then Inbox(0)
	order := []int{0, 1, 2, 3, 4, 5, 6, 7}
	for _, k := range order {
		if k == currentDay {
			continue
		}
		options = append(options, huh.NewOption(days[k], k))
	}

	s := huh.NewSelect[int]().Title(fmt.Sprintf("Move '%s' to", task.Name)).Key("days").Options(options...)
	s = s.Value(&task.Day)

	k := huh.NewDefaultKeyMap()
	k.Quit = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "Cancel"))
	return huh.NewForm(huh.NewGroup(s)).WithKeyMap(k)
}
