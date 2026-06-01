package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FileName is the configuration file expected at the root of every repo
// managed by worktree.
const FileName = ".worktree.yaml"

// Param describes a parameter declared in the config file.
type Param struct {
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     string `yaml:"default"`
}

// Config is the deserialized representation of .worktree.yaml.
type Config struct {
	Params      map[string]Param `yaml:"params"`
	Copy        []string         `yaml:"copy"`
	PreCreate   []string         `yaml:"pre_create"`
	PostCreate  []string         `yaml:"post_create"`
	PreRemove   []string         `yaml:"pre_remove"`
}

// Load reads .worktree.yaml from the given directory.
func Load(dir string) (*Config, error) {
	path := filepath.Join(dir, FileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no %s found in %s", FileName, dir)
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &c, nil
}

// ResolveParams merges user-supplied values with the declared params,
// applying defaults and validating required ones.
func (c *Config) ResolveParams(provided map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(c.Params)+len(provided))

	// Apply declared defaults first.
	for name, p := range c.Params {
		if p.Default != "" {
			out[name] = p.Default
		}
	}
	// User overrides.
	for k, v := range provided {
		out[k] = v
	}
	// Validate required.
	var missing []string
	for name, p := range c.Params {
		if p.Required {
			if _, ok := out[name]; !ok {
				missing = append(missing, name)
			}
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required params: %v (pass with -p KEY=VAL)", missing)
	}
	return out, nil
}
