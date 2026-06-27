package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/PRNVBAJAJ/context-os/internal/application"
	"github.com/spf13/cobra"
)

func newProvidersCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "providers",
		Short: "Manage AI CLI provider config injection",
		Long:  `Detect installed AI coding assistants and inject Context OS usage blocks into their project config files.`,
	}
	cmd.AddCommand(newProvidersListCommand())
	cmd.AddCommand(newProvidersInjectCommand())
	return cmd
}

func newProvidersListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Show detected providers and injection status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}

			results, err := application.DetectProviders(cmd.Context(), application.DetectProvidersOptions{
				RootPath: rootPath,
			})
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "PROVIDER\tDETECTED\tCONFIG FILE\tINJECTED")
			for _, r := range results {
				detected := "no"
				if r.Detected {
					detected = "yes"
				}
				injected := "-"
				if r.Injected {
					injected = "yes"
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
					r.Provider.Name,
					detected,
					r.Provider.ConfigPath,
					injected,
				)
			}
			return tw.Flush()
		},
	}
}

func newProvidersInjectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "inject [provider]",
		Short: "Inject Context OS block into provider config files",
		Long: `Inject the Context OS usage block into all detected provider config files,
or into a specific provider by name (e.g. 'context providers inject claude').

Injection is idempotent — running it multiple times is safe.
Works even if the provider binary is not installed (allows manual injection).`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()

			if len(args) == 1 {
				name := args[0]
				if err := application.InjectProvider(cmd.Context(), application.InjectProviderOptions{
					RootPath:     rootPath,
					ProviderName: name,
				}); err != nil {
					return err
				}
				fmt.Fprintf(w, "Injected Context OS block for provider %q.\n", name)
				return nil
			}

			// Snapshot pre-inject state to report only what actually changed.
			before, _ := application.DetectProviders(cmd.Context(), application.DetectProvidersOptions{
				RootPath: rootPath,
			})
			beforeInjected := make(map[string]bool, len(before))
			for _, r := range before {
				beforeInjected[r.Provider.Name] = r.Injected
			}

			results, err := application.InjectProviders(cmd.Context(), application.DetectProvidersOptions{
				RootPath: rootPath,
			})
			if err != nil {
				return err
			}

			newlyInjected := 0
			for _, r := range results {
				if r.Injected && !beforeInjected[r.Provider.Name] {
					newlyInjected++
				}
			}

			if newlyInjected == 0 {
				detected := 0
				for _, r := range results {
					if r.Detected {
						detected++
					}
				}
				if detected == 0 {
					fmt.Fprintln(w, "No AI CLI tools detected. Install one and re-run.")
				} else {
					fmt.Fprintln(w, "All detected providers are already configured.")
				}
				return nil
			}

			fmt.Fprintln(w, "Provider config updated:")
			for _, r := range results {
				if r.Injected && !beforeInjected[r.Provider.Name] {
					fmt.Fprintf(w, "  %-10s → %s\n", r.Provider.Name, r.Provider.ConfigPath)
				}
			}
			return nil
		},
	}
}
