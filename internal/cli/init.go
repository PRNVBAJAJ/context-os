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
		noInject bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a Context OS project in the current directory",
		Long: `Initialize a new Context OS project.

Creates .context/ with the runtime directory layout, writes project.yaml,
and bootstraps the SQLite metadata database (runtime.db).

AI CLI tools detected on this machine (claude, cursor, gemini, etc.) will
have a Context OS usage block appended to their project config files.
Use --no-inject to skip this step.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rootPath, err := os.Getwd()
			if err != nil {
				return shared.Wrap(shared.CodeInternal, "failed to determine working directory", err)
			}

			result, err := application.InitializeProject(cmd.Context(), application.InitOptions{
				Name:     name,
				RootPath: rootPath,
				Language: language,
				NoInject: noInject,
			})
			if err != nil {
				var domainErr *shared.Error
				if errors.As(err, &domainErr) && domainErr.Code == shared.CodeConflict {
					fmt.Fprintf(cmd.ErrOrStderr(), "already initialized: .context/ already exists in %s\n", rootPath)
					return err
				}
				return err
			}

			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Initialized Context OS project %q in .context/\n", result.Project.Name)

			if len(result.Providers) > 0 {
				injected := 0
				for _, r := range result.Providers {
					if r.Injected {
						injected++
					}
				}
				if injected > 0 {
					fmt.Fprintln(w, "\nProvider config updated:")
					for _, r := range result.Providers {
						if r.Injected {
							fmt.Fprintf(w, "  %-10s → %s\n", r.Provider.Name, r.Provider.ConfigPath)
						}
					}
					fmt.Fprintln(w, "\nRun 'context providers list' to see all detected providers.")
				} else {
					fmt.Fprintln(w, "\nNo AI CLI tools detected. Run 'context providers inject' after installing one.")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "project name (default: directory name)")
	cmd.Flags().StringVar(&language, "language", "", "primary programming language (e.g. go, python, typescript)")
	cmd.Flags().BoolVar(&noInject, "no-inject", false, "skip injecting Context OS block into AI tool config files")

	return cmd
}
