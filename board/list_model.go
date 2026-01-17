package board

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/rwirdemann/scheduled"
)

type ListModel struct {
	list.Model
	savedIndex int
	context    scheduled.Context
	allItems   []list.Item
}

func NewListModel(l list.Model) *ListModel {
	return &ListModel{Model: l, savedIndex: 0, context: scheduled.ContextNone}
}

func (lm *ListModel) SaveIndex() {
	lm.savedIndex = lm.Index()
}

func (lm *ListModel) RestoreIndex() {
	lm.Select(lm.savedIndex)
}

func (lm *ListModel) SetContext(context scheduled.Context) {
	// Context switch from none to specific
	if lm.context == scheduled.ContextNone && context != scheduled.ContextNone {

		// Backup all items
		lm.allItems = make([]list.Item, len(lm.Items()))
		copy(lm.allItems, lm.Items())

		// Remove items that do not belong to new context, backward to avoid
		// index problems
		items := lm.Items()
		for i := len(items) - 1; i >= 0; i-- {
			task := items[i].(scheduled.Task)
			if task.Context != context.ID {
				lm.RemoveItem(i)
			}
		}

	} else if context == scheduled.ContextNone {
		if lm.allItems != nil {
			for len(lm.Items()) > 0 {
				lm.RemoveItem(0)
			}
			for i, item := range lm.allItems {
				lm.InsertItem(i, item)
			}
			lm.allItems = nil
		}
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

func (lm *ListModel) Deselect() {
	lm.Select(-1)
}

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
