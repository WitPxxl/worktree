package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Create project-specific git worktrees and bootstrap their containers",
	Long: "worktree creates a git worktree for the current repo under " +
		"~/project/worktree/<branch> and runs the workflow declared in " +
		".worktree.yaml (copy / pre_create / post_create / pre_remove).",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute is the entry point used by main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
