package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/PRNVBAJAJ/context-os/internal/application"
	"github.com/PRNVBAJAJ/context-os/internal/project"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
	"github.com/spf13/cobra"
)

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the current project status",
		Long:  `Display project metadata and runtime state for the nearest Context OS project.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				var domainErr *shared.Error
				if errors.As(err, &domainErr) && domainErr.Code == shared.CodeNotFound {
					fmt.Fprintln(cmd.ErrOrStderr(), domainErr.Message)
					return err
				}
				return err
			}

			status, err := application.GetProjectStatus(cmd.Context(), rootPath)
			if err != nil {
				return err
			}

			p := status.Project
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Project:  %s\n", p.Name)
			fmt.Fprintf(w, "ID:       %s\n", p.ID)
			fmt.Fprintf(w, "Path:     %s\n", p.RootPath)
			if p.Language != "" {
				fmt.Fprintf(w, "Language: %s\n", p.Language)
			}
			fmt.Fprintf(w, "Runtime:  %s\n", shared.Version)
			fmt.Fprintf(w, "Memories: %d\n", status.MemoryCount)
			if status.ActiveWorkflow != nil {
				fmt.Fprintf(w, "Workflow: %s (%s)\n", status.ActiveWorkflow.Name, status.ActiveWorkflow.ID.String()[:8])
			}
			if status.LastCheckpoint != nil {
				note := status.LastCheckpoint.Note
				if note == "" {
					note = "(no note)"
				}
				fmt.Fprintf(w, "Last checkpoint: %s — %s\n",
					status.LastCheckpoint.CreatedAt.Format("2006-01-02T15:04Z"),
					note,
				)
			}
			if len(status.HotFiles) > 0 {
				fmt.Fprintln(w, "Hot files:")
				for _, f := range status.HotFiles {
					fmt.Fprintf(w, "  %s (%dx)\n", f.Filepath, f.AccessCount)
				}
			}
			return nil
		},
	}
}

// discoverProjectRoot finds the nearest project root from the working directory.
func discoverProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", shared.Wrap(shared.CodeInternal, "failed to determine working directory", err)
	}
	return project.Discover(cwd)
}
