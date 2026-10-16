// Package template copies embedded starter site files for the documentation index.
package template

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mindovermachine-dev/gh-docs/internal/config"
)

// The all: prefix is needed to include files starting with "_".
//
//go:embed all:jekyll
var jekyllFS embed.FS

// Config holds the template settings. Zero values mean "not set".
type Config struct {
	Output string `yaml:"output"`
}

var fallback = Config{
	Output: "docs",
}

// LoadConfig reads the `template` node of the config file.
func LoadConfig(cfgPath string, required bool) (Config, error) {
	var cfg Config
	err := config.Section(cfgPath, required, "template", &cfg)
	return cfg, err
}

// Merge applies overrides on top of base, then fills gaps with fallbacks.
func Merge(base, overrides Config) Config {
	out := base
	if overrides.Output != "" {
		out.Output = overrides.Output
	}
	if out.Output == "" {
		out.Output = fallback.Output
	}
	return out
}

// Jekyll copies the embedded Jekyll files into cfg.Output and returns the written paths.
// Existing files are only replaced when force is true.
func Jekyll(cfg Config, force bool) ([]string, error) {
	var names []string
	err := fs.WalkDir(jekyllFS, "jekyll", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			names = append(names, strings.TrimPrefix(p, "jekyll/"))
		}
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded templates: %w", err)
	}

	if !force {
		var existing []string
		for _, n := range names {
			if _, err := os.Stat(filepath.Join(cfg.Output, n)); err == nil {
				existing = append(existing, filepath.Join(cfg.Output, n))
			}
		}
		if len(existing) > 0 {
			return nil, fmt.Errorf("%s already exist(s); use --force to overwrite", strings.Join(existing, ", "))
		}
	}

	var written []string
	for _, n := range names {
		data, err := jekyllFS.ReadFile(path.Join("jekyll", n))
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded %s: %w", n, err)
		}
		dst := filepath.Join(cfg.Output, n)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return nil, fmt.Errorf("failed to create directory for %s: %w", dst, err)
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", dst, err)
		}
		written = append(written, dst)
	}
	return written, nil
}
