package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/PRNVBAJAJ/context-os/internal/adapter"
	"github.com/PRNVBAJAJ/context-os/internal/event"
	"github.com/PRNVBAJAJ/context-os/internal/memory"
	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/workflow"
)

const maxDashboardEvents = 10

type tab int

const (
	tabWorkflows tab = iota
	tabMemories
	tabProviders
	tabEvents
	tabCount
)

var tabNames = [tabCount]string{"Workflows", "Memories", "Providers", "Events"}

var (
	headerStyle      = lipgloss.NewStyle().Bold(true)
	activeTabStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Underline(true)
	inactiveTabStyle = lipgloss.NewStyle().Faint(true)
	sectionStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	dimStyle         = lipgloss.NewStyle().Faint(true)
	hintStyle        = lipgloss.NewStyle().Faint(true).MarginTop(1)
	cursorStyle      = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	colWidth         = 20
)

// Model is the Bubble Tea model for the Context OS dashboard.
// All data is populated before the program starts — no async loading.
type Model struct {
	project   *project.Project
	workflows []*workflow.Workflow
	memories  []*memory.Memory
	events    []*event.Event
	providers []adapter.DetectionResult
	activeTab tab
	cursor    int
	quitting  bool
}

// New returns a Model populated with the given data.
func New(
	p *project.Project,
	workflows []*workflow.Workflow,
	memories []*memory.Memory,
	events []*event.Event,
	providers []adapter.DetectionResult,
) Model {
	return Model{
		project:   p,
		workflows: workflows,
		memories:  memories,
		events:    events,
		providers: providers,
	}
}

// Quitting reports whether the model has received a quit signal.
func (m Model) Quitting() bool { return m.quitting }

// ActiveTab returns the index of the currently active tab.
func (m Model) ActiveTab() int { return int(m.activeTab) }

// Cursor returns the current cursor position within the active tab.
func (m Model) Cursor() int { return m.cursor }

// Init satisfies tea.Model. No startup commands are needed — data is pre-loaded.
func (m Model) Init() tea.Cmd { return nil }

// Update handles key messages and returns the updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "tab", "l":
			m.activeTab = (m.activeTab + 1) % tabCount
			m.cursor = 0
		case "shift+tab", "h":
			m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
			m.cursor = 0
		case "1":
			m.activeTab = tabWorkflows
			m.cursor = 0
		case "2":
			m.activeTab = tabMemories
			m.cursor = 0
		case "3":
			m.activeTab = tabProviders
			m.cursor = 0
		case "4":
			m.activeTab = tabEvents
			m.cursor = 0
		case "j", "down":
			if n := m.tabLen(); n > 0 {
				m.cursor = min(m.cursor+1, n-1)
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		}
	}
	return m, nil
}

// tabLen returns the number of navigable rows in the active tab.
func (m Model) tabLen() int {
	switch m.activeTab {
	case tabWorkflows:
		return len(m.workflows)
	case tabMemories:
		return len(m.memories)
	case tabProviders:
		return len(m.providers)
	case tabEvents:
		n := len(m.events)
		if n > maxDashboardEvents {
			n = maxDashboardEvents
		}
		return n
	}
	return 0
}

// View renders the dashboard. Returns an empty string once quitting to avoid
// a stale frame after the program exits.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Header
	b.WriteString("\n")
	b.WriteString(headerStyle.Render(fmt.Sprintf("  Context OS — %s (%s)", m.project.Name, m.project.RuntimeVersion)))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf("  %s", m.project.RootPath)))
	b.WriteString("\n\n")

	// Tab bar
	b.WriteString("  ")
	for i := tab(0); i < tabCount; i++ {
		label := fmt.Sprintf(" %s ", tabNames[i])
		if i == m.activeTab {
			b.WriteString(activeTabStyle.Render(label))
		} else {
			b.WriteString(inactiveTabStyle.Render(label))
		}
		if i < tabCount-1 {
			b.WriteString(dimStyle.Render(" │ "))
		}
	}
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  " + strings.Repeat("─", 60)))
	b.WriteString("\n\n")

	// Active tab content
	switch m.activeTab {
	case tabWorkflows:
		m.renderWorkflows(&b)
	case tabMemories:
		m.renderMemories(&b)
	case tabProviders:
		m.renderProviders(&b)
	case tabEvents:
		m.renderEvents(&b)
	}

	// Footer
	b.WriteString(hintStyle.Render("\n  tab/shift+tab switch tab  j/k navigate  1-4 jump  q quit"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) renderRow(selected bool, content string) string {
	if selected {
		return cursorStyle.Render("  ▶ " + content)
	}
	return "    " + content
}

func (m Model) renderWorkflows(b *strings.Builder) {
	if len(m.workflows) == 0 {
		b.WriteString(dimStyle.Render("  No workflows started."))
		b.WriteString("\n")
		return
	}
	for i, wf := range m.workflows {
		startedStr := ""
		if wf.StartedAt != nil {
			startedStr = wf.StartedAt.Format("2006-01-02T15:04Z")
		}
		line := fmt.Sprintf("%-8s  %-*s  %-10s  %s",
			wf.ID.String()[:8],
			colWidth, wf.Name,
			string(wf.Status),
			startedStr,
		)
		b.WriteString(m.renderRow(i == m.cursor, line))
		b.WriteString("\n")
	}
}

func (m Model) renderMemories(b *strings.Builder) {
	if len(m.memories) == 0 {
		b.WriteString(dimStyle.Render("  No memories recorded."))
		b.WriteString("\n")
		return
	}
	for i, mem := range m.memories {
		line := fmt.Sprintf("%-*s  %s", colWidth, mem.Key, mem.Title)
		b.WriteString(m.renderRow(i == m.cursor, line))
		b.WriteString("\n")
	}
}

func (m Model) renderProviders(b *strings.Builder) {
	if len(m.providers) == 0 {
		b.WriteString(dimStyle.Render("  No providers detected."))
		b.WriteString("\n")
		return
	}
	for i, r := range m.providers {
		detected := "no "
		if r.Detected {
			detected = "yes"
		}
		injected := "-  "
		if r.Injected {
			injected = "yes"
		}
		line := fmt.Sprintf("%-10s  detected:%-3s  injected:%-3s  %s",
			r.Provider.Name,
			detected,
			injected,
			r.Provider.ConfigPath,
		)
		b.WriteString(m.renderRow(i == m.cursor, line))
		b.WriteString("\n")
	}
}

func (m Model) renderEvents(b *strings.Builder) {
	events := m.events
	if len(events) > maxDashboardEvents {
		events = events[len(events)-maxDashboardEvents:]
	}
	if len(events) == 0 {
		b.WriteString(dimStyle.Render("  No events recorded."))
		b.WriteString("\n")
		return
	}
	// Display most recent first; cursor 0 = most recent.
	for i, idx := 0, len(events)-1; idx >= 0; i, idx = i+1, idx-1 {
		e := events[idx]
		line := fmt.Sprintf("%s  %s",
			e.Timestamp.Format("2006-01-02T15:04Z"),
			string(e.Type),
		)
		b.WriteString(m.renderRow(i == m.cursor, line))
		b.WriteString("\n")
	}
}

// sectionStyle is kept for potential future use by sub-panels.
var _ = sectionStyle
