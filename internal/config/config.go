// Package config reads per-subcommand sections from the shared config file.
package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Section decodes the top-level node `name` of the YAML file at path into out.
// A missing file is an error only when required is true.
func Section(path string, required bool, name string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && !required {
			return nil
		}
		return fmt.Errorf("failed to read config %s: %w", path, err)
	}
	var nodes map[string]yaml.Node
	if err := yaml.Unmarshal(data, &nodes); err != nil {
		return fmt.Errorf("failed to parse config %s: %w", path, err)
	}
	node, ok := nodes[name]
	if !ok {
		return nil
	}
	if err := node.Decode(out); err != nil {
		return fmt.Errorf("failed to parse '%s' in config %s: %w", name, path, err)
	}
	return nil
}
