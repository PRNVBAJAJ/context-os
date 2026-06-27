package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/PRNVBAJAJ/context-os/internal/application"
	"github.com/spf13/cobra"
)

func newMemoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Manage durable project knowledge",
		Long:  `Add and list project memories — durable knowledge that persists across provider switches.`,
	}
	cmd.AddCommand(newMemoryAddCommand())
	cmd.AddCommand(newMemoryListCommand())
	return cmd
}

func newMemoryAddCommand() *cobra.Command {
	var title string

	cmd := &cobra.Command{
		Use:   "add <key> <content>",
		Short: "Add a new memory entry",
		Long: `Add a named piece of project knowledge.

KEY must be a slug: lowercase letters, digits, and hyphens only (e.g. "auth-strategy").
CONTENT is the knowledge body — use quotes for multi-word content.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}

			m, err := application.AddMemory(cmd.Context(), application.AddMemoryOptions{
				RootPath: rootPath,
				Key:      args[0],
				Title:    title,
				Content:  args[1],
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Memory %q added.\n", m.Key)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Human-readable title (defaults to key)")
	return cmd
}

func newMemoryListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all project memories",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}

			memories, err := application.ListMemories(cmd.Context(), application.ListMemoriesOptions{
				RootPath: rootPath,
			})
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()

			if len(memories) == 0 {
				fmt.Fprintln(w, "No memories recorded. Use 'context memory add' to add one.")
				return nil
			}

			tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "KEY\tTITLE\tCREATED")
			for _, m := range memories {
				fmt.Fprintf(tw, "%s\t%s\t%s\n",
					m.Key,
					m.Title,
					m.CreatedAt.Format("2006-01-02T15:04Z"),
				)
			}
			return tw.Flush()
		},
	}
}
