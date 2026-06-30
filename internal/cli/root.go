package cli

import (
	"github.com/PRNVBAJAJ/context-os/internal/shared"
	"github.com/spf13/cobra"
)

// Execute builds the root command and runs it against os.Args.
// It is the single entrypoint called by cmd/context/main.go.
func Execute() error {
	return newRootCommand().Execute()
}

// newRootCommand constructs the cobra root command for the context binary.
// Subcommands are registered here as future milestones deliver them.
func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Context OS — provider-agnostic runtime for AI-assisted engineering",
		Long: `Context OS is a local-first runtime that provides persistent project state,
durable workflow execution, and shared memory across AI coding assistants.

It is not a coding assistant. It is the operating system beneath them.`,
		Version: shared.Version,
		// SilenceUsage prevents Cobra from printing usage on every non-usage error,
		// keeping output clean when domain errors are returned.
		SilenceUsage: true,
	}

	cmd.AddCommand(newInitCommand())
	cmd.AddCommand(newStatusCommand())
	cmd.AddCommand(newDoctorCommand())
	cmd.AddCommand(newMemoryCommand())
	cmd.AddCommand(newWorkflowCommand())
	cmd.AddCommand(newCheckpointCommand())
	cmd.AddCommand(newProvidersCommand())
	cmd.AddCommand(newTuiCommand())
	cmd.AddCommand(newTrackCommand())

	return cmd
}
