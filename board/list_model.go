package board

import "github.com/charmbracelet/bubbles/list"

// ListModel represents a custom list model with additional UI state.
type ListModel struct {
	list.Model
	savedIndex int
}

// NewListModel creates and returns a new instance of ListModel.
func NewListModel(l list.Model) *ListModel {
	return &ListModel{Model: l, savedIndex: 0}
}

// SaveIndex saves the current index of the list model.
func (lm *ListModel) SaveIndex() {
	lm.savedIndex = lm.Index()
}

// RestoreIndex restores the previously saved index of the list model.
func (lm *ListModel) RestoreIndex() {
	lm.Select(lm.savedIndex)
}

// Deselect clears the selection in the list model.
func (lm *ListModel) Deselect() {
	lm.Select(-1)
}
