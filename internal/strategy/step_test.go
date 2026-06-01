package strategy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyStep_CopiesFileWithMode(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	if err := os.WriteFile(filepath.Join(src, ".env"), []byte("KEY=value"), 0o600); err != nil {
		t.Fatalf("write src: %v", err)
	}

	step := CopyStep{RelPath: ".env"}
	if err := step.Execute(ExecContext{SourceDir: src, TargetDir: dst}); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dst, ".env"))
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != "KEY=value" {
		t.Errorf("contents = %q", got)
	}

	info, err := os.Stat(filepath.Join(dst, ".env"))
	if err != nil {
		t.Fatalf("stat dst: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %v, want 0600", info.Mode().Perm())
	}
}

func TestCopyStep_CreatesIntermediateDirs(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	nested := filepath.Join(src, "config")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nested, "app.yaml"), []byte("x: 1"), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	step := CopyStep{RelPath: "config/app.yaml"}
	if err := step.Execute(ExecContext{SourceDir: src, TargetDir: dst}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "config", "app.yaml")); err != nil {
		t.Errorf("expected nested dst file: %v", err)
	}
}

func TestCopyStep_MissingSource(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	step := CopyStep{RelPath: "missing.txt"}
	if err := step.Execute(ExecContext{SourceDir: src, TargetDir: dst}); err == nil {
		t.Fatal("expected error for missing source file")
	}
}

func TestRunStep_Describe(t *testing.T) {
	s := RunStep{Command: "echo hi", Dir: "target"}
	if got := s.Describe(); got != "[target] $ echo hi" {
		t.Errorf("Describe = %q", got)
	}
}

func TestRunStep_ExecuteInDir(t *testing.T) {
	dir := t.TempDir()
	// Write a file from inside the step's working directory, then verify it landed there.
	s := RunStep{Command: "touch marker", Dir: "target"}
	if err := s.Execute(ExecContext{TargetDir: dir}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "marker")); err != nil {
		t.Errorf("expected marker in target dir: %v", err)
	}
}
