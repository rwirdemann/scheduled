package board

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/rwirdemann/scheduled"
)

type ListModel struct {
	list.Model
	savedIndex int
}

func NewListModel(l list.Model) *ListModel {
	return &ListModel{Model: l, savedIndex: 0}
}

func (lm *ListModel) SaveIndex() {
	lm.savedIndex = lm.Index()
}

func (lm *ListModel) RestoreIndex() {
	lm.Select(lm.savedIndex)
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
	t := selected.(scheduled.Task)
	t.Done = !t.Done
	idx := lm.Index()
	lm.RemoveItem(idx)
	lm.InsertItem(idx, t)
	lm.Select(idx)
	return true
}
