package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/PRNVBAJAJ/context-os/internal/application"
	"github.com/PRNVBAJAJ/context-os/internal/shared"
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	var (
		name     string
		language string
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a Context OS project in the current directory",
		Long: `Initialize a new Context OS project.

Creates .context/ with the runtime directory layout, writes project.yaml,
and bootstraps the SQLite metadata database (runtime.db).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rootPath, err := os.Getwd()
			if err != nil {
				return shared.Wrap(shared.CodeInternal, "failed to determine working directory", err)
			}

			p, err := application.InitializeProject(cmd.Context(), application.InitOptions{
				Name:     name,
				RootPath: rootPath,
				Language: language,
			})
			if err != nil {
				var domainErr *shared.Error
				if errors.As(err, &domainErr) && domainErr.Code == shared.CodeConflict {
					fmt.Fprintf(cmd.ErrOrStderr(), "already initialized: .context/ already exists in %s\n", rootPath)
					return err
				}
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Initialized Context OS project %q in .context/\n", p.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "project name (default: directory name)")
	cmd.Flags().StringVar(&language, "language", "", "primary programming language (e.g. go, python, typescript)")

	return cmd
}
