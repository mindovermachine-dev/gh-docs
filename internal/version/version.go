// Package version writes a small build-provenance file.
package version

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mindovermachine-dev/gh-docs/internal/config"
)

// Config holds the version settings. Zero values mean "not set".
type Config struct {
	File string `yaml:"file"`
}

var fallback = Config{File: "dist/version.txt"}

// Info describes where a build came from.
type Info struct {
	Repo string
	Ref  string
	SHA  string
}

// LoadConfig reads the `version` node of the config file.
func LoadConfig(path string, required bool) (Config, error) {
	var cfg Config
	err := config.Section(path, required, "version", &cfg)
	return cfg, err
}

// Merge applies overrides on top of base, then fills gaps with fallbacks.
func Merge(base, overrides Config) Config {
	out := base
	if overrides.File != "" {
		out.File = overrides.File
	}
	if out.File == "" {
		out.File = fallback.File
	}
	return out
}

func git(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

// firstEnv returns the first non-empty environment variable.
func firstEnv(names ...string) string {
	for _, n := range names {
		if v := os.Getenv(n); v != "" {
			return v
		}
	}
	return ""
}

// Resolve reads BUILD_* variables first, then GitHub Actions variables, then git.
func Resolve() (Info, error) {
	info := Info{
		Repo: firstEnv("BUILD_REPO"),
		Ref:  firstEnv("BUILD_REF", "GITHUB_REF_NAME"),
		SHA:  firstEnv("BUILD_SHA", "GITHUB_SHA"),
	}
	if info.Repo == "" {
		if server, repo := os.Getenv("GITHUB_SERVER_URL"), os.Getenv("GITHUB_REPOSITORY"); server != "" && repo != "" {
			info.Repo = strings.TrimRight(server, "/") + "/" + repo
		}
	}

	var err error
	if info.Repo == "" {
		if info.Repo, err = git("remote", "get-url", "origin"); err != nil {
			return info, err
		}
	}
	if info.Ref == "" {
		if info.Ref, err = git("branch", "--show-current"); err != nil {
			return info, err
		}
		if info.Ref == "" {
			info.Ref = "detached"
		}
	}
	if info.SHA == "" {
		if info.SHA, err = git("rev-parse", "HEAD"); err != nil {
			return info, err
		}
	}
	return info, nil
}

// String renders the file content.
func (i Info) String() string {
	return fmt.Sprintf("Build from %s\nRef: %s\nCommit: %s\n", i.Repo, i.Ref, i.SHA)
}

// Run resolves the build info and writes it to cfg.File, returning the path.
func Run(cfg Config) (string, error) {
	info, err := Resolve()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(cfg.File), 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory for %s: %w", cfg.File, err)
	}
	if err := os.WriteFile(cfg.File, []byte(info.String()), 0o644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", cfg.File, err)
	}
	return cfg.File, nil
}
