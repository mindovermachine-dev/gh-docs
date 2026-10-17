// Package index builds an index of documentation repositories for an owner.
package index

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mindovermachine-dev/gh-docs/internal/config"
	"gopkg.in/yaml.v3"
)

// Config holds the index settings. Zero values mean "not set".
type Config struct {
	Owner      string `yaml:"owner"`
	Visibility string `yaml:"visibility"`
	Topic      string `yaml:"topic"`
	Format     string `yaml:"format"`
	File       string `yaml:"file"`
	Limit      int    `yaml:"limit"`
}

// Repo is a single entry in the generated index.
type Repo struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	HomepageURL string `yaml:"homepageUrl" json:"homepageUrl"`
	SidebarURL  string `yaml:"-" json:"-"`
}

type ghRepo struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	HomepageURL      string `json:"homepageUrl"`
	RepositoryTopics []struct {
		Name string `json:"name"`
	} `json:"repositoryTopics"`
}

// Fallback values used when neither a flag nor the config file sets them.
var fallback = Config{
	Visibility: "public",
	Topic:      "documentation",
	Format:     "yaml",
	File:       "docs-repos", // extension is appended from the format
	Limit:      1000,
}

func formatFromExt(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yml", ".yaml":
		return "yaml"
	case ".json":
		return "json"
	case ".html":
		return "html"
	case ".mjs":
		return "sidebar"
	}
	return ""
}

func extForFormat(format string) string {
	switch format {
	case "json":
		return ".json"
	case "html":
		return ".html"
	case "sidebar":
		return ".mjs"
	}
	return ".yml"
}

// LoadConfig reads a YAML config file. A missing file is an error only when required is true.
func LoadConfig(path string, required bool) (Config, error) {
	var cfg Config
	if err := config.Section(path, required, "index", &cfg); err != nil {
		return cfg, err
	}
	if cfg.File != "" && cfg.Format != "" {
		return cfg, fmt.Errorf("config %s: 'file' and 'format' cannot be used together", path)
	}
	return cfg, nil
}

// Merge returns overrides on top of base, then fills remaining gaps with fallbacks.
func Merge(base, overrides Config) Config {
	out := base
	if overrides.Owner != "" {
		out.Owner = overrides.Owner
	}
	if overrides.Visibility != "" {
		out.Visibility = overrides.Visibility
	}
	if overrides.Topic != "" {
		out.Topic = overrides.Topic
	}
	// File and Format are mutually exclusive; the higher-precedence source wins.
	if overrides.Format != "" {
		out.Format = overrides.Format
		out.File = ""
	}
	if overrides.File != "" {
		out.File = overrides.File
		out.Format = ""
	}
	if overrides.Limit != 0 {
		out.Limit = overrides.Limit
	}
	if out.Visibility == "" {
		out.Visibility = fallback.Visibility
	}
	if out.Topic == "" {
		out.Topic = fallback.Topic
	}
	if out.File != "" {
		out.Format = formatFromExt(out.File)
	} else {
		if out.Format == "" {
			out.Format = fallback.Format
		}
		out.File = fallback.File + extForFormat(out.Format)
	}
	if out.Limit == 0 {
		out.Limit = fallback.Limit
	}
	return out
}

// Validate checks that the config is usable.
func (c Config) Validate() error {
	if c.Owner == "" {
		return errors.New("owner is required (--owner or 'owner' in config)")
	}
	switch c.Visibility {
	case "public", "private", "internal":
	default:
		return fmt.Errorf("invalid visibility %q: must be public, private or internal", c.Visibility)
	}
	if c.File != "" && formatFromExt(c.File) == "" {
		return fmt.Errorf("invalid file extension %q: must be .yml, .yaml, .json, .html or .mjs", filepath.Ext(c.File))
	}
	switch c.Format {
	case "yaml", "json", "html", "sidebar":
	default:
		return fmt.Errorf("invalid format %q: must be yaml, json, html or sidebar", c.Format)
	}
	return nil
}

func ghList(owner, visibility string, limit int) ([]byte, error) {
	out, err := exec.Command("gh", "repo", "list", owner,
		"--visibility", visibility,
		"--limit", fmt.Sprint(limit),
		"--json", "name,homepageUrl,description,repositoryTopics").Output()
	if err != nil {
		return nil, fmt.Errorf("gh repo list failed: %w", err)
	}
	return out, nil
}

// Collect lists repos and keeps those carrying the configured topic.
func Collect(cfg Config) ([]Repo, error) {
	raw, err := ghList(cfg.Owner, cfg.Visibility, cfg.Limit)
	if err != nil {
		return nil, err
	}
	var repos []ghRepo
	if err := json.Unmarshal(raw, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse gh output: %w", err)
	}
	result := []Repo{}
	for _, r := range repos {
		for _, t := range r.RepositoryTopics {
			if strings.EqualFold(t.Name, cfg.Topic) {
				link := r.HomepageURL
				if link == "" {
					link = fmt.Sprintf("https://github.com/%s/%s", cfg.Owner, r.Name)
				}
				result = append(result, Repo{Name: r.Name, Description: r.Description, HomepageURL: r.HomepageURL, SidebarURL: link})
				break
			}
		}
	}
	return result, nil
}

var htmlTmpl = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>Documentation</title>
</head>
<body>
<h1>Documentation</h1>
<ul>
{{- range .}}
<li>{{if .HomepageURL}}<a href="{{.HomepageURL}}">{{.Name}}</a>{{else}}{{.Name}}{{end}}{{if .Description}} - {{.Description}}{{end}}</li>
{{- end}}
</ul>
</body>
</html>
`))

func renderSidebar(repos []Repo) ([]byte, error) {
	var out strings.Builder
	out.WriteString("export const sidebar = [\n  {\n    label: \"Documentation\",\n    items: [\n")
	for _, repo := range repos {
		label, err := json.Marshal(repo.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to encode sidebar label: %w", err)
		}
		linkValue := repo.SidebarURL
		if linkValue == "" {
			linkValue = repo.HomepageURL
		}
		link, err := json.Marshal(linkValue)
		if err != nil {
			return nil, fmt.Errorf("failed to encode sidebar link: %w", err)
		}
		out.WriteString("      { label: ")
		out.Write(label)
		out.WriteString(", link: ")
		out.Write(link)
		out.WriteString(" },\n")
	}
	out.WriteString("    ],\n  },\n];\n")
	return []byte(out.String()), nil
}

// Render serializes repos in the given format.
func Render(repos []Repo, format string) ([]byte, error) {
	if format == "sidebar" {
		return renderSidebar(repos)
	}
	if format == "html" {
		var buf bytes.Buffer
		if err := htmlTmpl.Execute(&buf, repos); err != nil {
			return nil, fmt.Errorf("failed to render html: %w", err)
		}
		return buf.Bytes(), nil
	}
	if format == "json" {
		data, err := json.MarshalIndent(repos, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json: %w", err)
		}
		return append(data, '\n'), nil
	}
	data, err := yaml.Marshal(repos)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal yaml: %w", err)
	}
	return data, nil
}

// Run collects repos and writes the index file, returning its path.
func Run(cfg Config) (string, error) {
	if err := cfg.Validate(); err != nil {
		return "", err
	}
	repos, err := Collect(cfg)
	if err != nil {
		return "", err
	}
	data, err := Render(repos, cfg.Format)
	if err != nil {
		return "", err
	}
	path := cfg.File
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory for %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", path, err)
	}
	return path, nil
}
