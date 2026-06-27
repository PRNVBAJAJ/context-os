package tui_test

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/PRNVBAJAJ/context-os/internal/adapter"
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

func pressKey(m tui.Model, key string) tui.Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated.(tui.Model)
}

func pressSpecialKey(m tui.Model, kt tea.KeyType) tui.Model {
	updated, _ := m.Update(tea.KeyMsg{Type: kt})
	return updated.(tui.Model)
}

func TestModel_Init_ReturnsNil(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil, nil)
	if cmd := m.Init(); cmd != nil {
		t.Error("Init should return nil for a pre-loaded model")
	}
}

func TestModel_View_ContainsProjectName(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil, nil)
	view := m.View()
	if !strings.Contains(view, "test-project") {
		t.Errorf("View should contain project name, got:\n%s", view)
	}
}

func TestModel_View_ContainsTabBar(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil, nil)
	view := m.View()
	for _, tab := range []string{"Workflows", "Memories", "Providers", "Events"} {
		if !strings.Contains(view, tab) {
			t.Errorf("View should contain tab %q:\n%s", tab, view)
		}
	}
}

func TestModel_View_DefaultTabIsWorkflows(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil, nil)
	if m.ActiveTab() != 0 {
		t.Errorf("default active tab should be 0 (Workflows), got %d", m.ActiveTab())
	}
}

func TestModel_View_ContainsWorkflows(t *testing.T) {
	p := makeTestProject(t)
	w1 := makeTestWorkflow(t, "implement auth")
	w2 := makeTestWorkflow(t, "refactor db")

	m := tui.New(p, []*workflow.Workflow{w1, w2}, nil, nil, nil)
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

	// Switch to Memories tab (tab index 1).
	m := tui.New(p, nil, []*memory.Memory{mem}, nil, nil)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	view := updated.(tui.Model).View()

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

	m := tui.New(p, nil, nil, []*event.Event{e}, nil)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	view := updated.(tui.Model).View()

	if !strings.Contains(view, "project.initialized") {
		t.Errorf("View should contain event type:\n%s", view)
	}
}

func TestModel_View_EmptyStateMessages(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil, nil)
	view := m.View()
	if !strings.Contains(view, "No workflows") {
		t.Errorf("View should contain 'No workflows':\n%s", view)
	}

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	if !strings.Contains(m2.(tui.Model).View(), "No memories") {
		t.Errorf("Memories tab should show 'No memories'")
	}

	m3, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	if !strings.Contains(m3.(tui.Model).View(), "No providers") {
		t.Errorf("Providers tab should show 'No providers'")
	}

	m4, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	if !strings.Contains(m4.(tui.Model).View(), "No events") {
		t.Errorf("Events tab should show 'No events'")
	}
}

func TestModel_Update_QuitOnQ(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil, nil)

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
	m := tui.New(p, nil, nil, nil, nil)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if !updated.(tui.Model).Quitting() {
		t.Error("model should be quitting after ctrl+c")
	}
}

func TestModel_Update_QuitOnEsc(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil, nil)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !updated.(tui.Model).Quitting() {
		t.Error("model should be quitting after esc")
	}
}

func TestModel_Update_IgnoresOtherKeys(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil, nil)

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
	m := tui.New(p, nil, nil, nil, nil)
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
	m := tui.New(p, workflows, nil, nil, nil)
	view := m.View()
	if !strings.Contains(view, "Workflows") {
		t.Errorf("View should show Workflows tab:\n%s", view)
	}
	for _, name := range []string{"first", "second", "third"} {
		if !strings.Contains(view, name) {
			t.Errorf("View should contain workflow %q:\n%s", name, view)
		}
	}
}

func TestModel_IDs_AreNotEmpty(t *testing.T) {
	id := shared.NewID()
	if id.IsEmpty() {
		t.Error("NewID() should not return EmptyID")
	}
}

// Tab navigation tests

func TestModel_TabSwitch_Forward(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil, nil)

	if m.ActiveTab() != 0 {
		t.Fatalf("expected tab 0, got %d", m.ActiveTab())
	}
	m = pressKey(m, "l")
	if m.ActiveTab() != 1 {
		t.Errorf("expected tab 1 after 'l', got %d", m.ActiveTab())
	}
	m = pressSpecialKey(m, tea.KeyTab)
	if m.ActiveTab() != 2 {
		t.Errorf("expected tab 2 after tab key, got %d", m.ActiveTab())
	}
}

func TestModel_TabSwitch_Backward(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil, nil)

	m = pressKey(m, "h")
	if m.ActiveTab() != 3 {
		t.Errorf("expected tab 3 (wrap) after 'h', got %d", m.ActiveTab())
	}
}

func TestModel_TabSwitch_ByNumber(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil, nil)

	for i, key := range []string{"1", "2", "3", "4"} {
		m2 := pressKey(m, key)
		if m2.ActiveTab() != i {
			t.Errorf("key %q: expected tab %d, got %d", key, i, m2.ActiveTab())
		}
	}
}

func TestModel_TabSwitch_ResetsCursor(t *testing.T) {
	p := makeTestProject(t)
	wfs := []*workflow.Workflow{
		makeTestWorkflow(t, "a"),
		makeTestWorkflow(t, "b"),
		makeTestWorkflow(t, "c"),
	}
	m := tui.New(p, wfs, nil, nil, nil)

	// Move cursor down twice.
	m = pressKey(m, "j")
	m = pressKey(m, "j")
	if m.Cursor() != 2 {
		t.Fatalf("expected cursor 2, got %d", m.Cursor())
	}

	// Switch tab — cursor must reset.
	m = pressKey(m, "l")
	if m.Cursor() != 0 {
		t.Errorf("cursor should reset to 0 on tab switch, got %d", m.Cursor())
	}
}

// Cursor navigation tests

func TestModel_Cursor_MoveDownAndUp(t *testing.T) {
	p := makeTestProject(t)
	wfs := []*workflow.Workflow{
		makeTestWorkflow(t, "a"),
		makeTestWorkflow(t, "b"),
		makeTestWorkflow(t, "c"),
	}
	m := tui.New(p, wfs, nil, nil, nil)

	m = pressKey(m, "j")
	if m.Cursor() != 1 {
		t.Errorf("expected cursor 1 after j, got %d", m.Cursor())
	}
	m = pressKey(m, "k")
	if m.Cursor() != 0 {
		t.Errorf("expected cursor 0 after k, got %d", m.Cursor())
	}
}

func TestModel_Cursor_DoesNotGoNegative(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, []*workflow.Workflow{makeTestWorkflow(t, "only")}, nil, nil, nil)

	m = pressKey(m, "k")
	if m.Cursor() != 0 {
		t.Errorf("cursor should not go below 0, got %d", m.Cursor())
	}
}

func TestModel_Cursor_DoesNotExceedLen(t *testing.T) {
	p := makeTestProject(t)
	wfs := []*workflow.Workflow{
		makeTestWorkflow(t, "a"),
		makeTestWorkflow(t, "b"),
	}
	m := tui.New(p, wfs, nil, nil, nil)

	// Press j many times.
	for i := 0; i < 10; i++ {
		m = pressKey(m, "j")
	}
	if m.Cursor() != 1 {
		t.Errorf("cursor should cap at last index (1), got %d", m.Cursor())
	}
}

func TestModel_Cursor_EmptyTab_NoMovement(t *testing.T) {
	p := makeTestProject(t)
	m := tui.New(p, nil, nil, nil, nil)

	m = pressKey(m, "j")
	if m.Cursor() != 0 {
		t.Errorf("cursor should stay at 0 on empty tab, got %d", m.Cursor())
	}
}

// Providers tab tests

func TestModel_View_ProvidersTab(t *testing.T) {
	p := makeTestProject(t)
	providers := []adapter.DetectionResult{
		{
			Provider: adapter.Provider{Name: "claude", Binary: "claude", ConfigPath: "CLAUDE.md"},
			Detected: true,
			Injected: true,
		},
		{
			Provider: adapter.Provider{Name: "gemini", Binary: "gemini", ConfigPath: "GEMINI.md"},
			Detected: false,
			Injected: false,
		},
	}

	m := tui.New(p, nil, nil, nil, providers)
	m = pressKey(m, "3")
	view := m.View()

	for _, want := range []string{"claude", "gemini", "CLAUDE.md", "GEMINI.md"} {
		if !strings.Contains(view, want) {
			t.Errorf("Providers view should contain %q:\n%s", want, view)
		}
	}
}

func TestModel_View_ProvidersCursorHighlight(t *testing.T) {
	p := makeTestProject(t)
	providers := []adapter.DetectionResult{
		{Provider: adapter.Provider{Name: "claude", ConfigPath: "CLAUDE.md"}, Detected: true, Injected: true},
		{Provider: adapter.Provider{Name: "gemini", ConfigPath: "GEMINI.md"}, Detected: false, Injected: false},
	}

	m := tui.New(p, nil, nil, nil, providers)
	m = pressKey(m, "3") // switch to Providers tab
	view := m.View()
	// Cursor on first row — cursor marker should appear.
	if !strings.Contains(view, "▶") {
		t.Errorf("cursor marker ▶ should appear on selected row:\n%s", view)
	}
}
