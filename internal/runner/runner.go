package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/WitPxxl/worktree/internal/config"
	"github.com/WitPxxl/worktree/internal/shell"
	"github.com/WitPxxl/worktree/internal/strategy"
)

// localBranchExists reports whether a local branch with the given name exists
// in the repo at dir. Output is suppressed; only the exit code matters.
func localBranchExists(dir, branch string) bool {
	c := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	c.Dir = dir
	return c.Run() == nil
}

// CreateOptions / RemoveOptions are the user-facing inputs for the two flows.
type CreateOptions struct {
	Branch string
	Params map[string]string
}

type RemoveOptions struct {
	Branch string
	Force  bool
}

// worktreeDirName converts a branch name into a safe single-segment directory
// name (e.g. "feature/foo" -> "feature_foo") so we never create nested dirs.
func worktreeDirName(branch string) string {
	return strings.ReplaceAll(branch, "/", "_")
}

// defaultWorktreeBase is used when the config doesn't declare worktree_dir.
const defaultWorktreeBase = "~/project/worktree"

// expandPath expands a leading "~" and ${ENV_VAR} references in a path.
func expandPath(p string) (string, error) {
	p = os.ExpandEnv(p)
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		if p == "~" {
			p = home
		} else {
			p = filepath.Join(home, p[2:])
		}
	}
	return p, nil
}

// resolveTarget computes the absolute worktree path for a branch given the
// (optional) base directory declared in the config.
func resolveTarget(base, branch string) (string, error) {
	if base == "" {
		base = defaultWorktreeBase
	}
	base, err := expandPath(base)
	if err != nil {
		return "", err
	}
	return filepath.Join(base, worktreeDirName(branch)), nil
}

// expander returns an os.Expand mapper for the given params.
func expander(params map[string]string) func(string) string {
	return func(name string) string {
		if v, ok := params[name]; ok {
			return v
		}
		// Fall back to process env so users can reference $HOME etc.
		return os.Getenv(name)
	}
}

// Create runs the create workflow defined in .worktree.yaml.
func Create(opts CreateOptions) error {
	if opts.Branch == "" {
		return fmt.Errorf("branch is required")
	}
	source, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}
	cfg, err := config.Load(source)
	if err != nil {
		return err
	}
	target, err := resolveTarget(cfg.WorktreeDir, opts.Branch)
	if err != nil {
		return err
	}
	params, err := cfg.ResolveParams(opts.Params)
	if err != nil {
		return err
	}
	// Inject built-ins (after resolve so user can't shadow them silently).
	params["BRANCH"] = opts.Branch
	params["TARGET"] = target
	params["SOURCE"] = source

	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("target worktree directory already exists: %s", target)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat target: %w", err)
	}

	ctx := strategy.ExecContext{SourceDir: source, TargetDir: target, Params: params}
	expand := expander(params)

	// 1. pre_create steps run in the source repo (e.g. git pull).
	for _, raw := range cfg.PreCreate {
		step := strategy.RunStep{Command: os.Expand(raw, expand), Dir: "source"}
		fmt.Printf("==> %s\n", step.Describe())
		if err := step.Execute(ctx); err != nil {
			return err
		}
	}

	// 2. Create the worktree. If the branch already exists locally, check it
	// out as-is; otherwise create it from the current HEAD.
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create parent directory: %w", err)
	}
	var gitArgs []string
	if localBranchExists(source, opts.Branch) {
		fmt.Printf("==> branch %q already exists, checking it out into %s\n", opts.Branch, target)
		gitArgs = []string{"worktree", "add", target, opts.Branch}
	} else {
		fmt.Printf("==> creating branch %q from HEAD and worktree at %s\n", opts.Branch, target)
		gitArgs = []string{"worktree", "add", "-b", opts.Branch, target}
	}
	if err := shell.Run(source, "git", gitArgs...); err != nil {
		return err
	}

	// 3. copy steps.
	for _, raw := range cfg.Copy {
		step := strategy.CopyStep{RelPath: os.Expand(raw, expand)}
		fmt.Printf("==> %s\n", step.Describe())
		if err := step.Execute(ctx); err != nil {
			return err
		}
	}

	// 4. post_create steps run inside the worktree.
	for _, raw := range cfg.PostCreate {
		step := strategy.RunStep{Command: os.Expand(raw, expand), Dir: "target"}
		fmt.Printf("==> %s\n", step.Describe())
		if err := step.Execute(ctx); err != nil {
			return err
		}
	}

	fmt.Printf("\nWorktree ready: %s\n", target)
	return nil
}

// Remove runs the teardown workflow defined in .worktree.yaml.
func Remove(opts RemoveOptions) error {
	if opts.Branch == "" {
		return fmt.Errorf("branch is required")
	}
	source, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}
	cfg, err := config.Load(source)
	if err != nil {
		return err
	}
	target, err := resolveTarget(cfg.WorktreeDir, opts.Branch)
	if err != nil {
		return err
	}

	if _, err := os.Stat(target); os.IsNotExist(err) {
		return fmt.Errorf("worktree directory does not exist: %s", target)
	} else if err != nil {
		return fmt.Errorf("stat target: %w", err)
	}

	// For pre_remove we only inject built-ins; user params aren't required for teardown.
	params := map[string]string{
		"BRANCH": opts.Branch,
		"TARGET": target,
		"SOURCE": source,
	}
	ctx := strategy.ExecContext{SourceDir: source, TargetDir: target, Params: params}
	expand := expander(params)

	for _, raw := range cfg.PreRemove {
		step := strategy.RunStep{Command: os.Expand(raw, expand), Dir: "target"}
		fmt.Printf("==> %s\n", step.Describe())
		if err := step.Execute(ctx); err != nil {
			// Non-fatal: containers/resources may already be gone.
			fmt.Fprintf(os.Stderr, "warning: %v (continuing)\n", err)
		}
	}

	args := []string{"worktree", "remove"}
	if opts.Force {
		args = append(args, "--force")
	}
	args = append(args, target)
	fmt.Printf("==> git %s\n", strings.Join(args, " "))
	if err := shell.Run(source, "git", args...); err != nil {
		return err
	}

	fmt.Printf("\nWorktree removed: %s\n", target)
	return nil
}
