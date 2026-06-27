package cli

import (
	"fmt"

	"github.com/PRNVBAJAJ/context-os/internal/application"
	"github.com/spf13/cobra"
)

func newDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check the health of the current Context OS project",
		Long:  `Validate the runtime directory, database connectivity, and display recent events.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}

			result, err := application.RunDoctor(cmd.Context(), rootPath)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Runtime:  OK\n")
			fmt.Fprintf(w, "Project:  %s (%s)\n", result.ProjectName, result.RuntimeVersion)
			fmt.Fprintf(w, "Database: OK\n")
			fmt.Fprintf(w, "Events:   %d recorded\n", result.EventCount)

			if len(result.RecentEvents) > 0 {
				fmt.Fprintln(w)
				fmt.Fprintln(w, "Recent events:")
				for _, e := range result.RecentEvents {
					fmt.Fprintf(w, "  %s  %s\n", e.Timestamp.Format("2006-01-02T15:04:05Z"), e.Type)
				}
			}

			return nil
		},
	}
}
