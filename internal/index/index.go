// Package index builds an index of documentation repositories for an owner.
package index

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"os"
	"path/filepath"
	"sort"
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
	PagesDir   string `yaml:"pages-dir"`
	BaseURL    string `yaml:"base-url"`
	BadgeURL   string `yaml:"badge-url"`
	Limit      int    `yaml:"limit"`
}

// Repo is a single entry in the generated index.
type Repo struct {
	Name          string       `yaml:"name" json:"name"`
	Description   string       `yaml:"description" json:"description"`
	HomepageURL   string       `yaml:"homepageUrl" json:"homepageUrl"`
	Owner         string       `yaml:"-" json:"-"`
	Title         string       `yaml:"-" json:"-"`
	Slug          string       `yaml:"-" json:"-"`
	RepoURL       string       `yaml:"-" json:"-"`
	Topics        []string     `yaml:"-" json:"-"`
	BadgeURL      string       `yaml:"-" json:"-"`
	DefaultBranch string       `yaml:"-" json:"-"`
	LatestCommit  LatestCommit `yaml:"-" json:"-"`
	LatestVersion string       `yaml:"-" json:"-"`
	Stars         int          `yaml:"-" json:"-"`
	License       string       `yaml:"-" json:"-"`
	UpdatedAt     string       `yaml:"-" json:"-"`
	SidebarURL    string       `yaml:"-" json:"-"`
}

// LatestCommit contains metadata for a repository's default-branch head commit.
type LatestCommit struct {
	SHA     string
	Date    string
	Message string
}

type ghRepo struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	HomepageURL      string `json:"homepageUrl"`
	DefaultBranchRef *struct {
		Name string `json:"name"`
	} `json:"defaultBranchRef"`
	StargazerCount int `json:"stargazerCount"`
	LicenseInfo    *struct {
		SPDXID string `json:"spdxId"`
	} `json:"licenseInfo"`
	UpdatedAt        string `json:"updatedAt"`
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
	PagesDir:   "docs/content/repo",
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
	if overrides.PagesDir != "" {
		out.PagesDir = overrides.PagesDir
	}
	if overrides.BaseURL != "" {
		out.BaseURL = overrides.BaseURL
	}
	if overrides.BadgeURL != "" {
		out.BadgeURL = overrides.BadgeURL
	}
	if out.Visibility == "" {
		out.Visibility = fallback.Visibility
	}
	if out.PagesDir == "" {
		out.PagesDir = fallback.PagesDir
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

func ghList(cfg Config) ([]byte, error) {
	fields := "name,homepageUrl,description,repositoryTopics"
	if cfg.Format == "sidebar" {
		fields += ",defaultBranchRef,stargazerCount,licenseInfo,updatedAt"
	}
	out, err := runGH("repo", "list", cfg.Owner,
		"--visibility", cfg.Visibility,
		"--limit", fmt.Sprint(cfg.Limit),
		"--json", fields)
	if err != nil {
		return nil, fmt.Errorf("gh repo list failed: %w", err)
	}
	return out, nil
}

// Collect lists repos and keeps those carrying the configured topic.
func Collect(cfg Config) ([]Repo, error) {
	raw, err := ghList(cfg)
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
				repo := Repo{
					Name: r.Name, Title: r.Name, Owner: cfg.Owner,
					Description: r.Description, HomepageURL: r.HomepageURL,
					RepoURL:  fmt.Sprintf("https://github.com/%s/%s", cfg.Owner, r.Name),
					BadgeURL: cfg.BadgeURL,
				}
				for _, topic := range r.RepositoryTopics {
					repo.Topics = append(repo.Topics, topic.Name)
				}
				if cfg.Format == "sidebar" {
					if r.DefaultBranchRef != nil {
						repo.DefaultBranch = r.DefaultBranchRef.Name
					}
					if r.LicenseInfo != nil {
						repo.License = r.LicenseInfo.SPDXID
					}
					repo.Stars, repo.UpdatedAt = r.StargazerCount, r.UpdatedAt
				}
				result = append(result, repo)
				break
			}
		}
	}
	if cfg.Format == "sidebar" {
		sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
		collectMetadata(cfg.Owner, result)
	}
	return result, nil
}

var htmlTmpl = htmltemplate.Must(htmltemplate.New("index").Parse(`<!DOCTYPE html>
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
			linkValue = pageURL("docs/content/repo", "", repo.Slug)
			if repo.Slug == "" {
				linkValue = pageURL("docs/content/repo", "", slug(repo.Name))
			}
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
	var pageTemplate pageTemplate
	if cfg.Format == "sidebar" {
		var err error
		pageTemplate, err = loadPageTemplate()
		if err != nil {
			return "", err
		}
	}
	repos, err := Collect(cfg)
	if err != nil {
		return "", err
	}
	if cfg.Format == "sidebar" {
		return writePagesAndSidebar(cfg, repos, pageTemplate)
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

// StarlightTemplate copies the maintained starter page template to destination.
func StarlightTemplate(destination string, force bool) error {
	return scaffoldStarlight(destination, force)
}

// PageURL returns the local Starlight route for a generated repository page.
func PageURL(pagesDir, baseURL, repoSlug string) string {
	return pageURL(pagesDir, baseURL, repoSlug)
}
