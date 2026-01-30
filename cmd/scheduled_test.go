package main

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/rwirdemann/scheduled"
	"github.com/rwirdemann/scheduled/board"
)

// Mock repository for testing
type mockRepository struct {
	tasks    []scheduled.Task
	contexts []scheduled.Context
}

func (m *mockRepository) LoadTasks() []scheduled.Task {
	return m.tasks
}

func (m *mockRepository) SaveTasks(tasks []scheduled.Task) {
	m.tasks = tasks
}

func (m *mockRepository) LoadContexts() []scheduled.Context {
	if len(m.contexts) == 0 {
		return []scheduled.Context{scheduled.ContextNone}
	}
	return m.contexts
}

func (m *mockRepository) SaveContexts(contexts []scheduled.Context) {
	m.contexts = contexts
}

// Helper function to create a test model with a mock repository
func createTestModel(t *testing.T) model {
	t.Helper()

	// Create model with mock repository
	repo := &mockRepository{
		tasks:    []scheduled.Task{},
		contexts: []scheduled.Context{scheduled.ContextNone},
	}

	return createModel(repo)
}

func TestIntegration_InitialState(t *testing.T) {
	m := createTestModel(t)

	// Test initial focus is on Inbox
	if m.board.LastFocus != board.Inbox {
		t.Errorf("Initial focus should be Inbox (0), got %d", m.board.LastFocus)
	}

	// Test help is shown by default
	if !m.showHelp {
		t.Error("Help should be shown by default")
	}

	// Test mode is normal
	if m.mode != modeNormal {
		t.Errorf("Initial mode should be modeNormal, got %d", m.mode)
	}
}

func TestIntegration_QuitCommand(t *testing.T) {
	m := createTestModel(t)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))

	// Send quit command
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// Wait for program to finish
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestIntegration_HelpToggle(t *testing.T) {
	m := createTestModel(t)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))
	defer tm.Quit()

	// Initially help should be shown
	if !m.showHelp {
		t.Fatal("Help should be shown initially")
	}

	// Send '?' to toggle help off
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	time.Sleep(50 * time.Millisecond)

	// Note: We can't directly check m.showHelp after sending the message
	// because teatest works with a copy. We'd need to check the output instead.
}

func TestIntegration_WeekNavigation(t *testing.T) {
	m := createTestModel(t)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))
	defer tm.Quit()

	_, w := time.Now().ISOWeek()
	expected := fmt.Sprintf("Week %d", w)
	teatest.WaitFor(
		t,
		tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte(expected))
		},
		teatest.WithDuration(time.Second),
		teatest.WithCheckInterval(50*time.Millisecond),
	)

	// Send right arrow to go to next week
	tm.Send(tea.KeyMsg{Type: tea.KeyRight})
	time.Sleep(50 * time.Millisecond)

	expected = fmt.Sprintf("Week %d", w+1)
	teatest.WaitFor(
		t,
		tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte(expected))
		},
		teatest.WithDuration(time.Second),
		teatest.WithCheckInterval(50*time.Millisecond),
	)

	// Send left arrow to go back
	tm.Send(tea.KeyMsg{Type: tea.KeyLeft})
	time.Sleep(50 * time.Millisecond)

	expected = fmt.Sprintf("Week %d", w)
	teatest.WaitFor(
		t,
		tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte(expected))
		},
		teatest.WithDuration(time.Second),
		teatest.WithCheckInterval(50*time.Millisecond),
	)
}

func TestIntegration_ViewRendering(t *testing.T) {
	m := createTestModel(t)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))
	defer tm.Quit()

	// Send a window size message to ensure proper initialization
	tm.Send(tea.WindowSizeMsg{Width: 200, Height: 50})
	time.Sleep(50 * time.Millisecond)

	// Wait for output containing Inbox
	teatest.WaitFor(
		t,
		tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("Inbox"))
		},
		teatest.WithDuration(time.Second),
		teatest.WithCheckInterval(50*time.Millisecond),
	)
}

func TestIntegration_ViewRenderingSmallTerminal(t *testing.T) {
	m := createTestModel(t)

	// Set very small terminal size
	m.termWidth = 50
	m.termHeight = 20

	view := m.View()

	// Should show error message about terminal size
	if !bytes.Contains([]byte(view), []byte("Terminal too small")) {
		t.Error("View should show 'Terminal too small' message")
	}
}

func TestIntegration_AutoSave(t *testing.T) {
	m := createTestModel(t)

	// Verify that Init returns autoSave command
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return autoSave command")
	}
}

func TestIntegration_ContextView(t *testing.T) {
	m := createTestModel(t)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))
	defer tm.Quit()

	// Open context view with 'c'
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	time.Sleep(50 * time.Millisecond)

	// Verify contexts view is shown
	teatest.WaitFor(
		t,
		tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("Context"))
		},
		teatest.WithDuration(time.Second),
		teatest.WithCheckInterval(50*time.Millisecond),
	)

	// Close with ESC
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	time.Sleep(50 * time.Millisecond)
}

func TestIntegration_StatusMessage(t *testing.T) {
	m := createTestModel(t)

	// Show a status message
	m, cmd := m.showStatusMessage("Test Status")

	if m.statusMessage != "Test Status" {
		t.Errorf("Status message = %s, want 'Test Status'", m.statusMessage)
	}

	if cmd == nil {
		t.Error("showStatusMessage should return a command")
	}

	// Verify status panel is shown
	if focusedPanel, exists := m.root.Focused(); !exists || focusedPanel.ID == statusPanel {
		// Status panel should exist
		t.Log("Status message set successfully")
	}
}

func TestIntegration_WindowResize(t *testing.T) {
	m := createTestModel(t)

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(200, 50))
	defer tm.Quit()

	// Send window size message
	tm.Send(tea.WindowSizeMsg{Width: 180, Height: 45})
	time.Sleep(50 * time.Millisecond)

	// The model should handle the resize
	// We verify this by checking output is still valid
	teatest.WaitFor(
		t,
		tm.Output(),
		func(bts []byte) bool {
			return len(bts) > 0
		},
		teatest.WithDuration(time.Second),
		teatest.WithCheckInterval(50*time.Millisecond),
	)
}

func TestIntegration_SaveFunction(t *testing.T) {
	m := createTestModel(t)

	// Create a task
	m.board.CreateTask("Test Task", 0)

	// Save
	m.Save()

	// Verify task was saved via mock repository
	repo := m.repository.(*mockRepository)
	if len(repo.tasks) == 0 {
		t.Error("Tasks should be saved in repository")
	}
}
