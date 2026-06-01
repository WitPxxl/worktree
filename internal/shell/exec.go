package shell

import (
	"fmt"
	"os"
	"os/exec"
)

// Run executes a command, streaming stdout/stderr to the user's terminal.
// If dir is non-empty the command runs in that working directory.
func Run(dir, name string, args ...string) error {
	c := exec.Command(name, args...)
	if dir != "" {
		c.Dir = dir
	}
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	if err := c.Run(); err != nil {
		return fmt.Errorf("%s %v: %w", name, args, err)
	}
	return nil
}
