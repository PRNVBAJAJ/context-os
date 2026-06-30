package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/PRNVBAJAJ/context-os/internal/application"
	"github.com/spf13/cobra"
)

func newMemorySuggestCommand() *cobra.Command {
	var topN int
	cmd := &cobra.Command{
		Use:   "suggest",
		Short: "Suggest memory entries based on recurring checkpoint note patterns",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}

			suggestions, err := application.SuggestMemories(cmd.Context(), application.SuggestMemoriesOptions{
				RootPath: rootPath,
				TopN:     topN,
			})
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			if len(suggestions) == 0 {
				fmt.Fprintln(w, "No suggestions yet — add more checkpoints with detailed notes.")
				return nil
			}

			fmt.Fprintf(w, "Suggested memory entries (run 'context memory add <key> <content>' to save):\n\n")
			tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "PHRASE\tSEEN IN\tEXAMPLE NOTE")
			for _, s := range suggestions {
				example := s.ExampleNote
				if len(example) > 55 {
					example = example[:52] + "..."
				}
				fmt.Fprintf(tw, "%s\t%d checkpoints\t%s\n", s.Phrase, s.Occurrences, example)
			}
			return tw.Flush()
		},
	}
	cmd.Flags().IntVar(&topN, "top", 10, "Maximum number of suggestions to show")
	return cmd
}

func newMemoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Manage durable project knowledge",
		Long:  `Add and list project memories — durable knowledge that persists across provider switches.`,
	}
	cmd.AddCommand(newMemoryAddCommand())
	cmd.AddCommand(newMemoryListCommand())
	cmd.AddCommand(newMemoryShowCommand())
	cmd.AddCommand(newMemoryUpdateCommand())
	cmd.AddCommand(newMemoryDeleteCommand())
	cmd.AddCommand(newMemorySuggestCommand())
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

func newMemoryShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <key>",
		Short: "Show a single memory entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}

			m, err := application.ShowMemory(cmd.Context(), application.ShowMemoryOptions{
				RootPath: rootPath,
				Key:      args[0],
			})
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "Key:     %s\n", m.Key)
			fmt.Fprintf(w, "Title:   %s\n", m.Title)
			fmt.Fprintf(w, "Created: %s\n", m.CreatedAt.Format("2006-01-02T15:04Z"))
			fmt.Fprintf(w, "\n%s\n", m.Content)
			return nil
		},
	}
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
			fmt.Fprintln(tw, "KEY\tTITLE\tCREATED\tPREVIEW")
			for _, m := range memories {
				preview := m.Content
				if len(preview) > 60 {
					preview = preview[:57] + "..."
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
					m.Key,
					m.Title,
					m.CreatedAt.Format("2006-01-02T15:04Z"),
					preview,
				)
			}
			return tw.Flush()
		},
	}
}

func newMemoryUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update <key> <content>",
		Short: "Update the content of a memory entry",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}

			_, err = application.UpdateMemory(cmd.Context(), application.UpdateMemoryOptions{
				RootPath: rootPath,
				Key:      args[0],
				Content:  args[1],
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Memory %q updated.\n", args[0])
			return nil
		},
	}
}

func newMemoryDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a memory entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}

			if err := application.DeleteMemory(cmd.Context(), application.DeleteMemoryOptions{
				RootPath: rootPath,
				Key:      args[0],
			}); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Memory %q deleted.\n", args[0])
			return nil
		},
	}
}
