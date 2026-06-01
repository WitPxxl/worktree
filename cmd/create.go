package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/WitPxxl/worktree/internal/runner"
)

func init() {
	var branch string
	var rawParams []string

	c := &cobra.Command{
		Use:     "create",
		Aliases: []string{"add", "new"},
		Short:   "Create a worktree and run the workflow from .worktree.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			params, err := parseParams(rawParams)
			if err != nil {
				return err
			}
			return runner.Create(runner.CreateOptions{Branch: branch, Params: params})
		},
	}
	c.Flags().StringVarP(&branch, "branch", "b", "", "branch name to create (required)")
	c.Flags().StringArrayVarP(&rawParams, "param", "p", nil, "param in KEY=VALUE form (repeatable)")
	_ = c.MarkFlagRequired("branch")
	rootCmd.AddCommand(c)
}

// parseParams turns ["KEY=VAL", ...] flags into a map, rejecting malformed entries.
func parseParams(raw []string) (map[string]string, error) {
	out := make(map[string]string, len(raw))
	for _, p := range raw {
		idx := strings.IndexByte(p, '=')
		if idx <= 0 {
			return nil, fmt.Errorf("invalid --param %q, expected KEY=VALUE", p)
		}
		out[p[:idx]] = p[idx+1:]
	}
	return out, nil
}
