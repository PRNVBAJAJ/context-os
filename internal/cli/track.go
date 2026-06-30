package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/PRNVBAJAJ/context-os/internal/application"
	"github.com/spf13/cobra"
)

func newTrackCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "track [filepath]",
		Short:  "Record a file access for the active workflow",
		Long:   `Called automatically by the PostToolUse hook. Reads PostToolUse JSON from stdin when no filepath is given.`,
		Hidden: true, // internal command — not shown in help
		Args:   cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return nil // not in a project — silently exit
			}

			opts := application.TrackFileOptions{RootPath: rootPath}

			if len(args) == 1 {
				opts.Filepath = args[0]
			} else {
				payload, err := io.ReadAll(os.Stdin)
				if err != nil {
					return nil // best-effort
				}
				opts.Payload = string(payload)
			}

			if err := application.TrackFile(cmd.Context(), opts); err != nil {
				// Silent: this command is called from hooks and must never surface errors.
				fmt.Fprintf(os.Stderr, "context track: %v\n", err)
			}
			return nil
		},
	}
}
