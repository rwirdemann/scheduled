package board

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/google/uuid"
	"github.com/rwirdemann/scheduled"
)

func TestListModel_MoveItemUp(t *testing.T) {
	tests := []struct {
		name          string
		initialIndex  int
		expectedIndex int
		shouldMove    bool
	}{
		{
			name:          "move item from middle",
			initialIndex:  1,
			expectedIndex: 0,
			shouldMove:    true,
		},
		{
			name:          "cannot move first item up",
			initialIndex:  0,
			expectedIndex: 0,
			shouldMove:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := createListModelWithTasks(3)
			lm.Select(tt.initialIndex)

			moved := lm.MoveItemUp()

			if moved != tt.shouldMove {
				t.Errorf("MoveItemUp() = %v, want %v", moved, tt.shouldMove)
			}

			if lm.Index() != tt.expectedIndex {
				t.Errorf("Index after move = %d, want %d", lm.Index(), tt.expectedIndex)
			}
		})
	}
}

func TestListModel_MoveItemDown(t *testing.T) {
	tests := []struct {
		name          string
		initialIndex  int
		expectedIndex int
		shouldMove    bool
	}{
		{
			name:          "move item from middle",
			initialIndex:  1,
			expectedIndex: 2,
			shouldMove:    true,
		},
		{
			name:          "cannot move last item down",
			initialIndex:  2,
			expectedIndex: 2,
			shouldMove:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := createListModelWithTasks(3)
			lm.Select(tt.initialIndex)

			moved := lm.MoveItemDown()

			if moved != tt.shouldMove {
				t.Errorf("MoveItemDown() = %v, want %v", moved, tt.shouldMove)
			}

			if lm.Index() != tt.expectedIndex {
				t.Errorf("Index after move = %d, want %d", lm.Index(), tt.expectedIndex)
			}
		})
	}
}

func TestListModel_ToggleDone(t *testing.T) {
	lm := createListModelWithTasks(2)
	lm.Select(0)

	// Get initial state
	initialTask := lm.SelectedItem().(scheduled.Task)
	if initialTask.Done {
		t.Fatal("Task should start as not done")
	}

	// Toggle to done
	toggled := lm.ToggleDone()
	if !toggled {
		t.Error("ToggleDone() should return true")
	}

	afterToggle := lm.SelectedItem().(scheduled.Task)
	if !afterToggle.Done {
		t.Error("Task should be marked as done after toggle")
	}

	// Toggle back to not done
	lm.ToggleDone()
	afterSecondToggle := lm.SelectedItem().(scheduled.Task)
	if afterSecondToggle.Done {
		t.Error("Task should be marked as not done after second toggle")
	}
}

func TestListModel_SetContext(t *testing.T) {
	t.Run("filter tasks by context", func(t *testing.T) {
		lm := NewListModel(list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0))

		// Add tasks with different contexts
		task1 := scheduled.Task{ID: uuid.NewString(), Name: "Task 1", Context: 1}
		task2 := scheduled.Task{ID: uuid.NewString(), Name: "Task 2", Context: 2}
		task3 := scheduled.Task{ID: uuid.NewString(), Name: "Task 3", Context: 1}

		lm.InsertItem(0, task1)
		lm.InsertItem(1, task2)
		lm.InsertItem(2, task3)

		// Set context filter to context 1
		lm.SetContext(scheduled.Context{ID: 1, Name: "Work"})

		// Should only show 2 items with context 1
		items := lm.Items()
		if len(items) != 2 {
			t.Errorf("Expected 2 filtered items, got %d", len(items))
		}

		// Verify filtered items are correct
		for _, item := range items {
			task := item.(scheduled.Task)
			if task.Context != 1 {
				t.Errorf("Filtered item has wrong context: %d", task.Context)
			}
		}

		// Verify allItems backup exists
		if lm.allItems == nil {
			t.Error("allItems should be backed up when filtering")
		}
		if len(lm.allItems) != 3 {
			t.Errorf("Expected 3 items in backup, got %d", len(lm.allItems))
		}
	})

	t.Run("restore all tasks when context is none", func(t *testing.T) {
		lm := NewListModel(list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0))

		task1 := scheduled.Task{ID: uuid.NewString(), Name: "Task 1", Context: 1}
		task2 := scheduled.Task{ID: uuid.NewString(), Name: "Task 2", Context: 2}

		lm.InsertItem(0, task1)
		lm.InsertItem(1, task2)

		// Filter by context
		lm.SetContext(scheduled.Context{ID: 1, Name: "Work"})
		if len(lm.Items()) != 1 {
			t.Fatalf("Expected 1 filtered item, got %d", len(lm.Items()))
		}

		// Restore all
		lm.SetContext(scheduled.ContextNone)

		if len(lm.Items()) != 2 {
			t.Errorf("Expected 2 items after restoring, got %d", len(lm.Items()))
		}

		if lm.allItems != nil {
			t.Error("allItems should be nil after restoring")
		}
	})
}

func TestListModel_SaveAndRestoreIndex(t *testing.T) {
	lm := createListModelWithTasks(5)
	lm.Select(3)

	// Save current index
	lm.SaveIndex()

	// Change selection
	lm.Select(0)
	if lm.Index() != 0 {
		t.Error("Index should be changed to 0")
	}

	// Restore
	lm.RestoreIndex()
	if lm.Index() != 3 {
		t.Errorf("Index should be restored to 3, got %d", lm.Index())
	}
}

// Helper function to create a ListModel with test tasks
func createListModelWithTasks(count int) *ListModel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	lm := NewListModel(l)

	for i := range count {
		task := scheduled.Task{
			ID:   uuid.NewString(),
			Name: "Task " + string(rune('A'+i)),
			Day:  0,
			Done: false,
		}
		lm.InsertItem(i, task)
	}

	return lm
}
