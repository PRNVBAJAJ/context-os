package cli

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/PRNVBAJAJ/context-os/internal/adapter"
	"github.com/PRNVBAJAJ/context-os/internal/application"
	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/storage"
	instui "github.com/PRNVBAJAJ/context-os/internal/tui"
)

func newTuiCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Open the interactive dashboard",
		Long:  `Launch the Context OS terminal dashboard showing project state, workflows, memories, and recent events.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}

			ctx := cmd.Context()

			// Load project metadata.
			p, err := project.Load(rootPath)
			if err != nil {
				return err
			}

			// Load workflows.
			workflows, err := application.ListWorkflows(ctx, application.ListWorkflowsOptions{
				RootPath: rootPath,
			})
			if err != nil {
				return err
			}

			// Load memories.
			memories, err := application.ListMemories(ctx, application.ListMemoriesOptions{
				RootPath: rootPath,
			})
			if err != nil {
				return err
			}

			// Load recent events via the storage layer directly — RunDoctor is overkill here.
			dbPath := filepath.Join(project.Dir(rootPath), "runtime.db")
			store, err := storage.Open(ctx, dbPath)
			if err != nil {
				return err
			}
			events, storeErr := store.Events().List(ctx, storage.EventFilter{})
			_ = store.Close()
			if storeErr != nil {
				return storeErr
			}

			providers := adapter.Detect(rootPath)

			m := instui.New(p, workflows, memories, events, providers)

			// Launch the interactive Bubble Tea program only when stdout is a
			// real terminal. In piped / CI environments, fall back to a static
			// plain-text render so the command always produces useful output.
			if isTTY(cmd.OutOrStdout()) {
				prog := tea.NewProgram(m,
					tea.WithInput(cmd.InOrStdin()),
					tea.WithOutput(cmd.OutOrStdout()),
				)
				_, err = prog.Run()
				return err
			}

			fmt.Fprint(cmd.OutOrStdout(), m.View())
			return nil
		},
	}
}

// isTTY reports whether w is a real terminal (character device).
// When false the TUI falls back to a static plain-text render.
func isTTY(w interface{}) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
