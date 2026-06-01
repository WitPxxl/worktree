package runner

import (
	"os"
	"testing"
)

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
