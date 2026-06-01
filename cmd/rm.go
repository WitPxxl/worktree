package cmd

import (
	"github.com/spf13/cobra"

	"github.com/WitPxxl/worktree/internal/runner"
)

func init() {
	var branch string
	var force bool

	c := &cobra.Command{
		Use:     "rm",
		Aliases: []string{"remove", "delete"},
		Short:   "Run pre_remove steps, then remove the worktree",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.Remove(runner.RemoveOptions{Branch: branch, Force: force})
		},
	}
	c.Flags().StringVarP(&branch, "branch", "b", "", "branch name whose worktree should be removed (required)")
	c.Flags().BoolVarP(&force, "force", "f", false, "force removal even if the worktree has local changes")
	_ = c.MarkFlagRequired("branch")
	rootCmd.AddCommand(c)
}
