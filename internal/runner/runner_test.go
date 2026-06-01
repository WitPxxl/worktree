package runner

import (
	"os"
	"os/exec"
	"testing"
)

// initGitRepo creates a bare-minimum repo with one commit on `main` so that
// HEAD exists. Skips the test if git isn't installed.
func initGitRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	run := func(args ...string) {
		c := exec.Command("git", args...)
		c.Dir = dir
		c.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
		)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-b", "main")
	if err := os.WriteFile(dir+"/README", []byte("hi"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	run("add", ".")
	run("commit", "-m", "init")
	return dir
}

func TestLocalBranchExists(t *testing.T) {
	dir := initGitRepo(t)

	if !localBranchExists(dir, "main") {
		t.Error("expected main to exist")
	}
	if localBranchExists(dir, "feature/new") {
		t.Error("did not expect feature/new to exist yet")
	}

	c := exec.Command("git", "branch", "feature/new")
	c.Dir = dir
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("create branch: %v\n%s", err, out)
	}
	if !localBranchExists(dir, "feature/new") {
		t.Error("expected feature/new to exist after creation")
	}
}

func TestResolveTarget_DefaultBase(t *testing.T) {
	t.Setenv("HOME", "/tmp/fake-home")
	got, err := resolveTarget("", "feature/foo")
	if err != nil {
		t.Fatalf("resolveTarget: %v", err)
	}
	want := "/tmp/fake-home/project/worktree/feature_foo"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveTarget_ExpandsTildeAndEnv(t *testing.T) {
	t.Setenv("HOME", "/tmp/fake-home")
	t.Setenv("WORKTREE_TEST_BASE", "/srv/worktrees")
	got, err := resolveTarget("${WORKTREE_TEST_BASE}/team", "main")
	if err != nil {
		t.Fatalf("resolveTarget: %v", err)
	}
	if got != "/srv/worktrees/team/main" {
		t.Errorf("got %q", got)
	}

	got, err = resolveTarget("~/wt", "main")
	if err != nil {
		t.Fatalf("resolveTarget: %v", err)
	}
	if got != "/tmp/fake-home/wt/main" {
		t.Errorf("got %q", got)
	}
}

func TestResolveTarget_AbsolutePath(t *testing.T) {
	got, err := resolveTarget("/var/wt", "feature/x")
	if err != nil {
		t.Fatalf("resolveTarget: %v", err)
	}
	if got != "/var/wt/feature_x" {
		t.Errorf("got %q", got)
	}
}

func TestWorktreeDirName(t *testing.T) {
	cases := map[string]string{
		"main":              "main",
		"feature/foo":       "feature_foo",
		"feature/foo/bar":   "feature_foo_bar",
		"release/2026/may":  "release_2026_may",
		"plain":             "plain",
	}
	for in, want := range cases {
		if got := worktreeDirName(in); got != want {
			t.Errorf("worktreeDirName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestExpander_KnownParam(t *testing.T) {
	m := map[string]string{"TOKEN": "abc"}
	expand := expander(m)
	if got := os.Expand("token=${TOKEN}", expand); got != "token=abc" {
		t.Errorf("got %q", got)
	}
}

func TestExpander_FallsBackToEnv(t *testing.T) {
	t.Setenv("WORKTREE_TEST_VAR", "from-env")
	expand := expander(map[string]string{})
	if got := os.Expand("v=${WORKTREE_TEST_VAR}", expand); got != "v=from-env" {
		t.Errorf("got %q", got)
	}
}

func TestExpander_ParamShadowsEnv(t *testing.T) {
	t.Setenv("WORKTREE_TEST_VAR", "from-env")
	expand := expander(map[string]string{"WORKTREE_TEST_VAR": "from-param"})
	if got := os.Expand("v=${WORKTREE_TEST_VAR}", expand); got != "v=from-param" {
		t.Errorf("got %q", got)
	}
}

func TestExpander_UnknownReturnsEmpty(t *testing.T) {
	expand := expander(map[string]string{})
	if got := os.Expand("v=${NOPE_NOT_DEFINED_XYZ}", expand); got != "v=" {
		t.Errorf("got %q", got)
	}
}
