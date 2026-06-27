package tui_test

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/PRNVBAJAJ/context-os/internal/event"
	"github.com/PRNVBAJAJ/context-os/internal/memory"
	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
	"github.com/PRNVBAJAJ/context-os/internal/tui"
	"github.com/PRNVBAJAJ/context-os/internal/workflow"
)

func makeTestProject(t *testing.T) *project.Project {
	t.Helper()
	p, err := project.New("test-project", t.TempDir(), "go")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func makeTestWorkflow(t *testing.T, name string) *workflow.Workflow {
	t.Helper()
	w, err := workflow.New(name, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Start(); err != nil {
		t.Fatal(err)
	}
	return w
}

func makeTestMemory(t *testing.T, key, title string) *memory.Memory {
	t.Helper()
	m, err := memory.New(key, title, "content")
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestModel_Init_ReturnsNil(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil)
	if cmd := m.Init(); cmd != nil {
		t.Error("Init should return nil for a pre-loaded model")
	}
}

func TestModel_View_ContainsProjectName(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil)
	view := m.View()
	if !strings.Contains(view, "test-project") {
		t.Errorf("View should contain project name, got:\n%s", view)
	}
}

func TestModel_View_ContainsWorkflows(t *testing.T) {
	p := makeTestProject(t)
	w1 := makeTestWorkflow(t, "implement auth")
	w2 := makeTestWorkflow(t, "refactor db")

	m := tui.New(p, []*workflow.Workflow{w1, w2}, nil, nil)
	view := m.View()

	for _, want := range []string{"implement auth", "refactor db", "running"} {
		if !strings.Contains(view, want) {
			t.Errorf("View should contain %q:\n%s", want, view)
		}
	}
}

func TestModel_View_ContainsMemories(t *testing.T) {
	p := makeTestProject(t)
	mem := makeTestMemory(t, "auth-strategy", "Auth Strategy")

	m := tui.New(p, nil, []*memory.Memory{mem}, nil)
	view := m.View()

	if !strings.Contains(view, "auth-strategy") {
		t.Errorf("View should contain memory key:\n%s", view)
	}
	if !strings.Contains(view, "Auth Strategy") {
		t.Errorf("View should contain memory title:\n%s", view)
	}
}

func TestModel_View_ContainsEvents(t *testing.T) {
	p := makeTestProject(t)
	e := event.New(event.TypeProjectInitialized, "{}")
	e.Timestamp = time.Date(2026, 6, 27, 9, 0, 0, 0, time.UTC)

	m := tui.New(p, nil, nil, []*event.Event{e})
	view := m.View()

	if !strings.Contains(view, "project.initialized") {
		t.Errorf("View should contain event type:\n%s", view)
	}
}

func TestModel_View_EmptyStateMessages(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil)
	view := m.View()

	for _, want := range []string{"No workflows", "No memories", "No events"} {
		if !strings.Contains(view, want) {
			t.Errorf("View should contain empty-state %q:\n%s", want, view)
		}
	}
}

func TestModel_Update_QuitOnQ(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	updatedModel := updated.(tui.Model)

	if !updatedModel.Quitting() {
		t.Error("model should be quitting after 'q' key")
	}
	if updatedModel.View() != "" {
		t.Error("View() should return empty string when quitting")
	}
}

func TestModel_Update_QuitOnCtrlC(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if !updated.(tui.Model).Quitting() {
		t.Error("model should be quitting after ctrl+c")
	}
}

func TestModel_Update_QuitOnEsc(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !updated.(tui.Model).Quitting() {
		t.Error("model should be quitting after esc")
	}
}

func TestModel_Update_IgnoresOtherKeys(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if updated.(tui.Model).Quitting() {
		t.Error("model should not quit on unrecognised key")
	}
	if cmd != nil {
		t.Error("unrecognised key should return nil cmd")
	}
}

func TestModel_View_ContainsQuitHint(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil)
	if !strings.Contains(m.View(), "quit") {
		t.Error("View should contain quit hint")
	}
}

func TestModel_View_WorkflowCount(t *testing.T) {
	p := makeTestProject(t)
	workflows := []*workflow.Workflow{
		makeTestWorkflow(t, "first"),
		makeTestWorkflow(t, "second"),
		makeTestWorkflow(t, "third"),
	}
	m := tui.New(p, workflows, nil, nil)
	view := m.View()
	if !strings.Contains(view, "WORKFLOWS (3)") {
		t.Errorf("View should show workflow count:\n%s", view)
	}
}

func TestModel_IDs_AreNotEmpty(t *testing.T) {
	// Verify that shared.NewID() produces non-empty IDs used in the view.
	id := shared.NewID()
	if id.IsEmpty() {
		t.Error("NewID() should not return EmptyID")
	}
}
