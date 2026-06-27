package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/PRNVBAJAJ/context-os/internal/application"
	"github.com/spf13/cobra"
)

func newWorkflowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Manage durable engineering workflows",
		Long:  `Start, list, and transition workflows — structured engineering tasks that persist across provider switches.`,
	}
	cmd.AddCommand(newWorkflowStartCommand())
	cmd.AddCommand(newWorkflowListCommand())
	cmd.AddCommand(newWorkflowCompleteCommand())
	cmd.AddCommand(newWorkflowFailCommand())
	cmd.AddCommand(newWorkflowPauseCommand())
	cmd.AddCommand(newWorkflowResumeCommand())
	return cmd
}

func newWorkflowStartCommand() *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "start <name>",
		Short: "Start a new workflow",
		Long: `Create and immediately start a new workflow.

NAME is the human-readable label for this engineering task (e.g. "implement auth", "refactor db layer").`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}

			w, err := application.StartWorkflow(cmd.Context(), application.StartWorkflowOptions{
				RootPath:    rootPath,
				Name:        args[0],
				Description: description,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Workflow %q started (ID: %s).\n", w.Name, w.ID.String()[:8])
			return nil
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "Optional description of the workflow goal")
	return cmd
}

func newWorkflowListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all workflows",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}

			workflows, err := application.ListWorkflows(cmd.Context(), application.ListWorkflowsOptions{
				RootPath: rootPath,
			})
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()

			if len(workflows) == 0 {
				fmt.Fprintln(w, "No workflows started. Use 'context workflow start' to begin one.")
				return nil
			}

			tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tNAME\tSTATUS\tSTARTED")
			for _, wf := range workflows {
				startedStr := ""
				if wf.StartedAt != nil {
					startedStr = wf.StartedAt.Format("2006-01-02T15:04Z")
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
					wf.ID.String()[:8],
					wf.Name,
					string(wf.Status),
					startedStr,
				)
			}
			return tw.Flush()
		},
	}
}

func newWorkflowCompleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "complete <id>",
		Short: "Mark a running workflow as completed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}
			w, err := application.CompleteWorkflow(cmd.Context(), application.CompleteWorkflowOptions{
				RootPath: rootPath, IDPrefix: args[0],
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Workflow %q completed.\n", w.Name)
			return nil
		},
	}
}

func newWorkflowFailCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "fail <id>",
		Short: "Mark a running workflow as failed",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}
			w, err := application.FailWorkflow(cmd.Context(), application.FailWorkflowOptions{
				RootPath: rootPath, IDPrefix: args[0],
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Workflow %q failed.\n", w.Name)
			return nil
		},
	}
}

func newWorkflowPauseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "pause <id>",
		Short: "Pause a running workflow",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}
			w, err := application.PauseWorkflow(cmd.Context(), application.PauseWorkflowOptions{
				RootPath: rootPath, IDPrefix: args[0],
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Workflow %q paused.\n", w.Name)
			return nil
		},
	}
}

func newWorkflowResumeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "resume <id>",
		Short: "Resume a paused workflow",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}
			w, err := application.ResumeWorkflow(cmd.Context(), application.ResumeWorkflowOptions{
				RootPath: rootPath, IDPrefix: args[0],
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Workflow %q resumed.\n", w.Name)
			return nil
		},
	}
}
