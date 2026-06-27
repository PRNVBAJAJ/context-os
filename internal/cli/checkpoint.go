package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/PRNVBAJAJ/context-os/internal/application"
	"github.com/spf13/cobra"
)

func newCheckpointCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkpoint",
		Short: "Manage recovery snapshots",
		Long:  `Create and list checkpoints — point-in-time snapshots that capture where you left off.`,
	}
	cmd.AddCommand(newCheckpointCreateCommand())
	cmd.AddCommand(newCheckpointListCommand())
	cmd.AddCommand(newCheckpointRestoreCommand())
	return cmd
}

func newCheckpointCreateCommand() *cobra.Command {
	var (
		workflowPrefix string
		note           string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a recovery checkpoint",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}

			cp, err := application.CreateCheckpoint(cmd.Context(), application.CreateCheckpointOptions{
				RootPath:         rootPath,
				WorkflowIDPrefix: workflowPrefix,
				Note:             note,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Checkpoint created (ID: %s).\n", cp.ID.String()[:8])
			return nil
		},
	}

	cmd.Flags().StringVar(&workflowPrefix, "workflow", "", "Scope checkpoint to a workflow by ID prefix")
	cmd.Flags().StringVar(&note, "note", "", "Human description of the current state")
	return cmd
}

func newCheckpointListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all checkpoints",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}

			checkpoints, err := application.ListCheckpoints(cmd.Context(), application.ListCheckpointsOptions{
				RootPath: rootPath,
			})
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()

			if len(checkpoints) == 0 {
				fmt.Fprintln(w, "No checkpoints recorded. Use 'context checkpoint create' to add one.")
				return nil
			}

			tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tWORKFLOW\tNOTE\tCREATED")
			for _, cp := range checkpoints {
				wfStr := ""
				if !cp.WorkflowID.IsEmpty() {
					wfStr = cp.WorkflowID.String()[:8]
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
					cp.ID.String()[:8],
					wfStr,
					cp.Note,
					cp.CreatedAt.Format("2006-01-02T15:04Z"),
				)
			}
			return tw.Flush()
		},
	}
}

func newCheckpointRestoreCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "restore <id>",
		Short: "Show a checkpoint's recorded state",
		Long: `Display the state captured at the time a checkpoint was created.

In v0.1, restore surfaces the checkpoint note and associated workflow so you
or your AI assistant can resume from exactly where work was recorded.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootPath, err := discoverProjectRoot()
			if err != nil {
				return err
			}

			result, err := application.RestoreCheckpoint(cmd.Context(), application.RestoreCheckpointOptions{
				RootPath: rootPath,
				IDPrefix: args[0],
			})
			if err != nil {
				return err
			}

			cp := result.Checkpoint
			w := cmd.OutOrStdout()

			fmt.Fprintf(w, "Checkpoint: %s\n", cp.ID.String()[:8])
			fmt.Fprintf(w, "Created:    %s\n", cp.CreatedAt.Format("2006-01-02T15:04:05Z"))

			if result.WorkflowName != "" {
				fmt.Fprintf(w, "Workflow:   %s (%s)\n", result.WorkflowName, cp.WorkflowID.String()[:8])
			} else {
				fmt.Fprintf(w, "Workflow:   (project-level)\n")
			}

			if cp.Note != "" {
				fmt.Fprintf(w, "\nState at checkpoint:\n  %s\n", cp.Note)
			}

			if !cp.WorkflowID.IsEmpty() {
				fmt.Fprintf(w, "\nTo resume the associated workflow:\n  context workflow resume %s\n", cp.WorkflowID.String()[:8])
			}

			return nil
		},
	}
}
