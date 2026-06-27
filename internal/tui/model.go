package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/PRNVBAJAJ/context-os/internal/event"
	"github.com/PRNVBAJAJ/context-os/internal/memory"
	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/workflow"
)

const maxDashboardEvents = 10

var (
	headerStyle  = lipgloss.NewStyle().Bold(true)
	sectionStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	dimStyle     = lipgloss.NewStyle().Faint(true)
	hintStyle    = lipgloss.NewStyle().Faint(true).MarginTop(1)
	colWidth     = 20
)

// Model is the Bubble Tea model for the Context OS dashboard.
// All data is populated before the program starts — no async loading.
type Model struct {
	project   *project.Project
	workflows []*workflow.Workflow
	memories  []*memory.Memory
	events    []*event.Event
	quitting  bool
}

// New returns a Model populated with the given data.
func New(
	p *project.Project,
	workflows []*workflow.Workflow,
	memories []*memory.Memory,
	events []*event.Event,
) Model {
	return Model{
		project:   p,
		workflows: workflows,
		memories:  memories,
		events:    events,
	}
}

// Quitting reports whether the model has received a quit signal.
// Used by tests to verify key-press handling without a real terminal.
func (m Model) Quitting() bool { return m.quitting }

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
		}
	}
	return m, nil
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
	b.WriteString("\n")

	// Workflows section
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render(fmt.Sprintf("  ── WORKFLOWS (%d) ", len(m.workflows))))
	b.WriteString("\n")
	if len(m.workflows) == 0 {
		b.WriteString(dimStyle.Render("  No workflows started."))
		b.WriteString("\n")
	} else {
		for _, wf := range m.workflows {
			startedStr := ""
			if wf.StartedAt != nil {
				startedStr = wf.StartedAt.Format("2006-01-02T15:04Z")
			}
			fmt.Fprintf(&b, "  %s  %-*s  %-10s  %s\n",
				wf.ID.String()[:8],
				colWidth, wf.Name,
				string(wf.Status),
				startedStr,
			)
		}
	}

	// Memories section
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render(fmt.Sprintf("  ── MEMORIES (%d) ", len(m.memories))))
	b.WriteString("\n")
	if len(m.memories) == 0 {
		b.WriteString(dimStyle.Render("  No memories recorded."))
		b.WriteString("\n")
	} else {
		for _, mem := range m.memories {
			fmt.Fprintf(&b, "  %-*s  %s\n", colWidth, mem.Key, mem.Title)
		}
	}

	// Recent events section
	events := m.events
	if len(events) > maxDashboardEvents {
		events = events[len(events)-maxDashboardEvents:]
	}
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render(fmt.Sprintf("  ── RECENT EVENTS (%d) ", len(m.events))))
	b.WriteString("\n")
	if len(events) == 0 {
		b.WriteString(dimStyle.Render("  No events recorded."))
		b.WriteString("\n")
	} else {
		// Display most recent first.
		for i := len(events) - 1; i >= 0; i-- {
			e := events[i]
			fmt.Fprintf(&b, "  %s  %s\n",
				dimStyle.Render(e.Timestamp.Format("2006-01-02T15:04Z")),
				string(e.Type),
			)
		}
	}

	// Footer hint
	b.WriteString(hintStyle.Render("\n  q quit"))
	b.WriteString("\n")

	return b.String()
}
