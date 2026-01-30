package board

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/rwirdemann/scheduled"
)

// ListModel represents a custom list model with additional context management.
type ListModel struct {
	list.Model
	savedIndex int
	context    scheduled.Context
	allItems   []list.Item
}

// NewListModel creates and returns a new instance of ListModel.
func NewListModel(l list.Model) *ListModel {
	return &ListModel{Model: l, savedIndex: 0, context: scheduled.ContextNone}
}

// SaveIndex saves the current index of the list model.
func (lm *ListModel) SaveIndex() {
	lm.savedIndex = lm.Index()
}

// RestoreIndex restores the previously saved index of the list model.
func (lm *ListModel) RestoreIndex() {
	lm.Select(lm.savedIndex)
}

// SetContext updates the list model to display items for the specified context.
// It switches between contexts, filters items, or restores all items if needed.
func (lm *ListModel) SetContext(context scheduled.Context) {
	// Context switch from none to specific
	if lm.context == scheduled.ContextNone && context != scheduled.ContextNone {

		// Back up all items
		lm.allItems = make([]list.Item, len(lm.Items()))
		copy(lm.allItems, lm.Items())

		// Remove items that do not belong to the new context, backward to avoid
		// index problems
		items := lm.Items()
		for i := len(items) - 1; i >= 0; i-- {
			task := items[i].(scheduled.Task)
			if task.Context != context.ID {
				lm.RemoveItem(i)
			}
		}

	} else if context == scheduled.ContextNone && lm.allItems != nil {
		for len(lm.Items()) > 0 {
			lm.RemoveItem(0)
		}
		for i, item := range lm.allItems {
			lm.InsertItem(i, item)
		}
		lm.allItems = nil
	} else if lm.context != scheduled.ContextNone && context != scheduled.ContextNone {

		// Reinsert all items
		if lm.allItems != nil {
			for len(lm.Items()) > 0 {
				lm.RemoveItem(0)
			}
			for i, item := range lm.allItems {
				lm.InsertItem(i, item)
			}
		}

		// Filter for new context
		lm.allItems = make([]list.Item, len(lm.Items()))
		copy(lm.allItems, lm.Items())

		items := lm.Items()
		for i := len(items) - 1; i >= 0; i-- {
			task := items[i].(scheduled.Task)
			if task.Context != context.ID {
				lm.RemoveItem(i)
			}
		}
	}

	lm.context = context
}

// Deselect clears the selection in the list model.
func (lm *ListModel) Deselect() {
	lm.Select(-1)
}

// MoveItemUp moves the selected item up in the list.
func (lm *ListModel) MoveItemUp() bool {
	if lm.Index() <= 0 {
		return false
	}
	selected := lm.SelectedItem()
	if selected == nil {
		return false
	}
	t := selected.(scheduled.Task)
	lm.RemoveItem(lm.Index())
	lm.InsertItem(lm.Index()-1, t)
	lm.Select(lm.Index() - 1)
	return true
}

// MoveItemDown moves the selected item down in the list.
func (lm *ListModel) MoveItemDown() bool {
	if lm.Index() < 0 || lm.Index() >= len(lm.Items())-1 {
		return false
	}
	selected := lm.SelectedItem()
	if selected == nil {
		return false
	}
	t := selected.(scheduled.Task)
	lm.RemoveItem(lm.Index())
	lm.InsertItem(lm.Index()+1, t)
	lm.Select(lm.Index() + 1)
	return true
}

// ToggleDone toggles the done state of the selected task.
func (lm *ListModel) ToggleDone() bool {
	selected := lm.SelectedItem()
	if selected == nil {
		return false
	}
	oldTask := selected.(scheduled.Task)
	t := oldTask
	t.Done = !t.Done
	idx := lm.Index()
	lm.RemoveItem(idx)
	lm.InsertItem(idx, t)
	lm.Select(idx)

	// Synchronize allItems when a context filter is active
	if lm.allItems != nil {
		for i, item := range lm.allItems {
			task := item.(scheduled.Task)
			if task.ID == oldTask.ID {
				lm.allItems[i] = t
				break
			}
		}
	}

	return true
}
