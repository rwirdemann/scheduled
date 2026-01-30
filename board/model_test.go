package board

import (
	"testing"

	"github.com/google/uuid"
	"github.com/rwirdemann/scheduled"
)

// Mock repository for testing
type mockRepository struct {
	tasks []scheduled.Task
}

func (m *mockRepository) LoadTasks() []scheduled.Task {
	return m.tasks
}

func (m *mockRepository) SaveTasks(tasks []scheduled.Task) {
	m.tasks = tasks
}

func TestModel_DecWeek(t *testing.T) {
	tests := []struct {
		name         string
		initialWeek  int
		expectedWeek int
	}{
		{
			name:         "decrement normal week",
			initialWeek:  10,
			expectedWeek: 9,
		},
		{
			name:         "wrap from week 1 to week 52",
			initialWeek:  1,
			expectedWeek: 52,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{tasks: []scheduled.Task{}}
			m := NewModel(repo)
			m.setWeek(tt.initialWeek)

			m.DecWeek()

			if m.Week() != tt.expectedWeek {
				t.Errorf("Week() = %d, want %d", m.Week(), tt.expectedWeek)
			}
		})
	}
}

func TestModel_IncWeek(t *testing.T) {
	tests := []struct {
		name         string
		initialWeek  int
		expectedWeek int
	}{
		{
			name:         "increment normal week",
			initialWeek:  10,
			expectedWeek: 11,
		},
		{
			name:         "wrap from week 52 to week 1",
			initialWeek:  52,
			expectedWeek: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{tasks: []scheduled.Task{}}
			m := NewModel(repo)
			m.setWeek(tt.initialWeek)

			m.IncWeek()

			if m.Week() != tt.expectedWeek {
				t.Errorf("Week() = %d, want %d", m.Week(), tt.expectedWeek)
			}
		})
	}
}

func TestModel_CreateTask(t *testing.T) {
	repo := &mockRepository{tasks: []scheduled.Task{}}
	m := NewModel(repo)
	m.LastFocus = Monday

	initialCount := len(m.GetTasksForPanel(Monday))

	// Create a new task
	m.CreateTask("New Task", 1)

	tasks := m.GetTasksForPanel(Monday)
	if len(tasks) != initialCount+1 {
		t.Errorf("Expected %d tasks, got %d", initialCount+1, len(tasks))
	}

	// Verify task properties
	newTask := tasks[len(tasks)-1]
	if newTask.Name != "New Task" {
		t.Errorf("Task name = %s, want 'New Task'", newTask.Name)
	}
	if newTask.Context != 1 {
		t.Errorf("Task context = %d, want 1", newTask.Context)
	}
	if newTask.Day != Monday {
		t.Errorf("Task day = %d, want %d", newTask.Day, Monday)
	}
}

func TestModel_UpdateTask(t *testing.T) {
	task := scheduled.Task{
		ID:      uuid.NewString(),
		Name:    "Original Name",
		Context: 1,
		Day:     Monday,
	}

	repo := &mockRepository{tasks: []scheduled.Task{task}}
	m := NewModel(repo)
	m.LastFocus = Monday

	// Select the task first
	m.lists[Monday].Select(0)

	// Update the task
	m.UpdateTask("Updated Name", 2)

	tasks := m.GetTasksForPanel(Monday)
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	updatedTask := tasks[0]
	if updatedTask.Name != "Updated Name" {
		t.Errorf("Task name = %s, want 'Updated Name'", updatedTask.Name)
	}
	if updatedTask.Context != 2 {
		t.Errorf("Task context = %d, want 2", updatedTask.Context)
	}
}

func TestModel_MoveTask(t *testing.T) {
	task := scheduled.Task{
		ID:   uuid.NewString(),
		Name: "Task to Move",
		Day:  Monday,
	}

	repo := &mockRepository{tasks: []scheduled.Task{task}}
	m := NewModel(repo)

	// Select the task first
	m.lists[Monday].Select(0)

	// Move task from Monday to Tuesday
	m.MoveTask(Monday, Tuesday)

	mondayTasks := m.GetTasksForPanel(Monday)
	tuesdayTasks := m.GetTasksForPanel(Tuesday)

	if len(mondayTasks) != 0 {
		t.Errorf("Monday should have 0 tasks, got %d", len(mondayTasks))
	}

	if len(tuesdayTasks) != 1 {
		t.Fatalf("Tuesday should have 1 task, got %d", len(tuesdayTasks))
	}

	movedTask := tuesdayTasks[0]
	if movedTask.Day != Tuesday {
		t.Errorf("Task day = %d, want %d", movedTask.Day, Tuesday)
	}
	if movedTask.Name != "Task to Move" {
		t.Errorf("Task name changed unexpectedly: %s", movedTask.Name)
	}
}

func TestModel_MoveTask_InvalidRange(t *testing.T) {
	task := scheduled.Task{
		ID:   uuid.NewString(),
		Name: "Task",
		Day:  Monday,
	}

	repo := &mockRepository{tasks: []scheduled.Task{task}}
	m := NewModel(repo)

	// Select the task first
	m.lists[Monday].Select(0)

	// Try to move to invalid day
	m.MoveTask(Monday, 99)

	// Task should still be in Monday
	mondayTasks := m.GetTasksForPanel(Monday)
	if len(mondayTasks) != 1 {
		t.Errorf("Task should remain in Monday after invalid move")
	}
}

func TestModel_DeleteTask(t *testing.T) {
	task := scheduled.Task{
		ID:   uuid.NewString(),
		Name: "Task to Delete",
		Day:  Monday,
		Done: true, // Must be done to be deletable
	}

	repo := &mockRepository{tasks: []scheduled.Task{task}}
	m := NewModel(repo)

	initialCount := len(m.GetTasksForPanel(Monday))

	// Select the task first
	m.lists[Monday].Select(0)

	// Delete the task
	m.DeleteTask(Monday)

	tasks := m.GetTasksForPanel(Monday)
	if len(tasks) != initialCount-1 {
		t.Errorf("Expected %d tasks after deletion, got %d", initialCount-1, len(tasks))
	}
}

func TestModel_DeleteTask_NotDone(t *testing.T) {
	task := scheduled.Task{
		ID:   uuid.NewString(),
		Name: "Task",
		Day:  Monday,
		Done: false, // Not done
	}

	repo := &mockRepository{tasks: []scheduled.Task{task}}
	m := NewModel(repo)

	// Select the task first
	m.lists[Monday].Select(0)

	// Try to delete task that's not done
	m.DeleteTask(Monday)

	// Task should still exist
	tasks := m.GetTasksForPanel(Monday)
	if len(tasks) != 1 {
		t.Error("Task should not be deleted if not done")
	}
}

func TestModel_ToggleDone(t *testing.T) {
	task := scheduled.Task{
		ID:   uuid.NewString(),
		Name: "Task",
		Day:  Monday,
		Done: false,
	}

	repo := &mockRepository{tasks: []scheduled.Task{task}}
	m := NewModel(repo)

	// Select the task first
	m.lists[Monday].Select(0)

	// Toggle done
	m.ToggleDone(Monday)

	tasks := m.GetTasksForPanel(Monday)
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}

	if !tasks[0].Done {
		t.Error("Task should be marked as done")
	}
}

func TestModel_SetContext(t *testing.T) {
	repo := &mockRepository{tasks: []scheduled.Task{}}
	m := NewModel(repo)

	context := scheduled.Context{ID: 1, Name: "Work"}
	m.SetContext(context)

	if m.GetSelectedContext() != context {
		t.Error("Context should be set correctly")
	}
}

func TestModel_IsContextUsed(t *testing.T) {
	task1 := scheduled.Task{ID: uuid.NewString(), Name: "Task 1", Context: 1, Day: Monday}
	task2 := scheduled.Task{ID: uuid.NewString(), Name: "Task 2", Context: 2, Day: Tuesday}

	repo := &mockRepository{tasks: []scheduled.Task{task1, task2}}
	m := NewModel(repo)

	workContext := scheduled.Context{ID: 1, Name: "Work"}
	homeContext := scheduled.Context{ID: 2, Name: "Home"}
	unusedContext := scheduled.Context{ID: 3, Name: "Unused"}

	if !m.IsContextUsed(workContext) {
		t.Error("Work context should be used")
	}

	if !m.IsContextUsed(homeContext) {
		t.Error("Home context should be used")
	}

	if m.IsContextUsed(unusedContext) {
		t.Error("Unused context should not be marked as used")
	}
}

func TestModel_SaveTasks(t *testing.T) {
	task1 := scheduled.Task{ID: uuid.NewString(), Name: "Task 1", Day: Monday}
	task2 := scheduled.Task{ID: uuid.NewString(), Name: "Task 2", Day: Tuesday}

	repo := &mockRepository{tasks: []scheduled.Task{task1, task2}}
	m := NewModel(repo)

	// Modify and save
	m.SaveTasks()

	// Verify positions are set
	savedTasks := repo.tasks
	foundTask1 := false
	foundTask2 := false

	for _, task := range savedTasks {
		if task.ID == task1.ID {
			foundTask1 = true
		}
		if task.ID == task2.ID {
			foundTask2 = true
		}
	}

	if !foundTask1 || !foundTask2 {
		t.Error("All tasks should be saved")
	}
}

func TestModel_GetSelectedTask(t *testing.T) {
	task := scheduled.Task{
		ID:   uuid.NewString(),
		Name: "Selected Task",
		Day:  Inbox, // Use Inbox since it's selected by default
	}

	repo := &mockRepository{tasks: []scheduled.Task{task}}
	m := NewModel(repo)

	selectedTask, ok := m.GetSelectedTask(Inbox)
	if !ok {
		t.Fatal("Should have a selected task")
	}

	if selectedTask.Name != "Selected Task" {
		t.Errorf("Selected task name = %s, want 'Selected Task'", selectedTask.Name)
	}
}
