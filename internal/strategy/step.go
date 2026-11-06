package strategy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/WitPxxl/worktree/internal/shell"
)

// ExecContext carries the per-invocation state passed to every Step.
type ExecContext struct {
	SourceDir string            // repo containing .worktree.yaml
	TargetDir string            // newly-created worktree directory
	Params    map[string]string // resolved params + built-ins, already expanded into commands by caller
}

// Step is the Strategy interface: each kind of action implements it.
// New step types (git, http, template, ...) can be added without touching
// the runner.
type Step interface {
	Describe() string
	Execute(ctx ExecContext) error
}

// RunStep executes a shell command via `sh -c` in the configured directory.
// The Dir field selects which directory: "source" or "target".
type RunStep struct {
	Command string
	Dir     string // "source" or "target"
}

func (s RunStep) Describe() string { return fmt.Sprintf("[%s] $ %s", s.Dir, s.Command) }

func (s RunStep) Execute(ctx ExecContext) error {
	workdir := ctx.TargetDir
	if s.Dir == "source" {
		workdir = ctx.SourceDir
	}
	return shell.RunEnv(workdir, ctx.Params, "sh", "-c", s.Command)
}

// CopyStep copies a single file from source repo to the new worktree,
// preserving the relative path.
type CopyStep struct {
	RelPath string
}

func (s CopyStep) Describe() string { return fmt.Sprintf("copy %s -> <target>/%s", s.RelPath, s.RelPath) }

func (s CopyStep) Execute(ctx ExecContext) error {
	if !filepath.IsLocal(s.RelPath) {
		return fmt.Errorf("copy path escapes directory: %s", s.RelPath)
	}

	src := filepath.Join(ctx.SourceDir, s.RelPath)
	dst := filepath.Join(ctx.TargetDir, s.RelPath)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("mkdir for copy: %w", err)
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
