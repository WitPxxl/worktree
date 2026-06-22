package shell

import (
	"fmt"
	"os"
	"os/exec"
)

// Run executes a command, streaming stdout/stderr to the user's terminal.
// If dir is non-empty the command runs in that working directory.
// If env is non-nil, its variables are appended to the command's environment.
func Run(dir string, env map[string]string, name string, args ...string) error {
	c := exec.Command(name, args...)
	if dir != "" {
		c.Dir = dir
	}
	if len(env) > 0 {
		c.Env = os.Environ()
		for k, v := range env {
			c.Env = append(c.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	if err := c.Run(); err != nil {
		return fmt.Errorf("%s %v: %w", name, args, err)
	}
	return nil
}
