package index

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed templates/starlight.md
var starterFS embed.FS

var safeName = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
var stableVersion = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

const defaultPageTemplate = "docs/repository-template.md"

func runGH(args ...string) ([]byte, error) {
	return exec.Command("gh", args...).Output()
}

func slug(name string) string {
	var out strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(name) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			out.WriteRune(r)
			lastDash = false
		} else if !lastDash {
			out.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(out.String(), "-")
}

func pageURL(pagesDir, baseURL, repoSlug string) string {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		clean := filepath.ToSlash(filepath.Clean(pagesDir))
		parts := strings.Split(clean, "/")
		for i := len(parts) - 2; i >= 0; i-- {
			if parts[i] == "content" && parts[i+1] == "docs" {
				base = strings.Join(parts[i+2:], "/")
				break
			}
		}
		if base == "" {
			base = filepath.Base(clean)
		}
	}
	base = strings.Trim(base, "/")
	if base == "" || base == "." {
		return "/" + repoSlug + "/"
	}
	return "/" + base + "/" + repoSlug + "/"
}

func loadPageTemplate() (*template.Template, error) {
	source, err := os.ReadFile(defaultPageTemplate)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("template %s is missing; create it with `gh docs template --starlight --output docs`", defaultPageTemplate)
		}
		return nil, fmt.Errorf("failed to read template %s: %w", defaultPageTemplate, err)
	}
	tmpl, err := template.New(defaultPageTemplate).Delims("[%", "%]").Funcs(template.FuncMap{
		"yaml": func(value any) (string, error) {
			data, err := yaml.Marshal(value)
			return strings.TrimSuffix(string(data), "\n"), err
		},
	}).Parse(string(source))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", defaultPageTemplate, err)
	}
	return tmpl, nil
}

func scaffoldStarlight(destination string, force bool) error {
	if _, err := os.Lstat(destination); err == nil && !force {
		return fmt.Errorf("template destination %s already exists; use --force to overwrite", destination)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to check template destination %s: %w", destination, err)
	}
	source, err := starterFS.ReadFile("templates/starlight.md")
	if err != nil {
		return fmt.Errorf("failed to read embedded Starlight template: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", destination, err)
	}
	if force {
		if err := os.WriteFile(destination, source, 0o644); err != nil {
			return fmt.Errorf("failed to write template %s: %w", destination, err)
		}
		return nil
	}
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write template %s: %w", destination, err)
	}
	if _, err := file.Write(source); err != nil {
		_ = file.Close()
		_ = os.Remove(destination)
		return fmt.Errorf("failed to write template %s: %w", destination, err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(destination)
		return fmt.Errorf("failed to close template %s: %w", destination, err)
	}
	return nil
}

func writePagesAndSidebar(cfg Config, repos []Repo, tmpl *template.Template) (string, error) {
	var err error
	pagesDir, err := filepath.Abs(cfg.PagesDir)
	if err != nil {
		return "", fmt.Errorf("invalid pages directory %s: %w", cfg.PagesDir, err)
	}
	type page struct {
		path string
		data []byte
	}
	pages := make([]page, 0, len(repos))
	seen := make(map[string]string, len(repos))
	for i := range repos {
		repo := &repos[i]
		repo.Slug = slug(repo.Name)
		if repo.Slug == "" {
			return "", fmt.Errorf("repository %q has no safe page slug", repo.Name)
		}
		if previous, ok := seen[repo.Slug]; ok {
			return "", fmt.Errorf("repositories %q and %q produce the same page slug %q", previous, repo.Name, repo.Slug)
		}
		seen[repo.Slug] = repo.Name
		target := filepath.Join(pagesDir, repo.Slug+".md")
		relative, err := filepath.Rel(pagesDir, target)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("page path for repository %q escapes pages directory", repo.Name)
		}
		if info, err := os.Lstat(target); err == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("page path %s is a symlink", target)
		} else if err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("failed to inspect page path %s: %w", target, err)
		}
		repo.SidebarURL = pageURL(cfg.PagesDir, cfg.BaseURL, repo.Slug)
		var rendered bytes.Buffer
		if err := tmpl.Execute(&rendered, *repo); err != nil {
			return "", fmt.Errorf("failed to render template %s for repository %s: %w", defaultPageTemplate, repo.Name, err)
		}
		pages = append(pages, page{path: target, data: rendered.Bytes()})
	}
	sort.Slice(repos, func(i, j int) bool { return repos[i].Name < repos[j].Name })
	sidebar, err := renderSidebar(repos)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(pagesDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create pages directory %s: %w", pagesDir, err)
	}
	for _, p := range pages {
		if err := os.MkdirAll(filepath.Dir(p.path), 0o755); err != nil {
			return "", fmt.Errorf("failed to create directory for %s: %w", p.path, err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(cfg.File), 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory for %s: %w", cfg.File, err)
	}
	for _, p := range pages {
		if err := os.WriteFile(p.path, p.data, 0o644); err != nil {
			return "", fmt.Errorf("failed to write %s: %w", p.path, err)
		}
	}
	if err := os.WriteFile(cfg.File, sidebar, 0o644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", cfg.File, err)
	}
	return cfg.File, nil
}

func collectMetadata(owner string, repos []Repo) {
	jobs := make(chan int)
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for i := range jobs {
				repo := &repos[i]
				if !safeName.MatchString(owner) || !safeName.MatchString(repo.Name) {
					log.Printf("warning: skipping GitHub metadata for invalid repository path %q/%q", owner, repo.Name)
					continue
				}
				collectRepoMetadata(owner, repo)
			}
		}()
	}
	for i := range repos {
		jobs <- i
	}
	close(jobs)
	workers.Wait()
}

func collectRepoMetadata(owner string, repo *Repo) {
	endpoint := "repos/" + owner + "/" + repo.Name
	commitQuery := "per_page=1"
	if repo.DefaultBranch != "" {
		commitQuery = "sha=" + url.QueryEscape(repo.DefaultBranch) + "&" + commitQuery
	}
	commitData, err := runGH("api", endpoint+"/commits?"+commitQuery)
	if err != nil {
		log.Printf("warning: failed to fetch latest commit for %s/%s: %v", owner, repo.Name, err)
	} else {
		var commits []struct {
			SHA    string `json:"sha"`
			Commit struct {
				Author struct {
					Date string `json:"date"`
				} `json:"author"`
				Message string `json:"message"`
			} `json:"commit"`
		}
		if err := json.Unmarshal(commitData, &commits); err != nil {
			log.Printf("warning: failed to parse latest commit for %s/%s: %v", owner, repo.Name, err)
		} else if len(commits) > 0 {
			sha := commits[0].SHA
			if len(sha) > 7 {
				sha = sha[:7]
			}
			date := ""
			if parsed, err := time.Parse(time.RFC3339, commits[0].Commit.Author.Date); err == nil {
				date = parsed.Format(time.RFC3339)
			}
			message := strings.SplitN(commits[0].Commit.Message, "\n", 2)[0]
			repo.LatestCommit = LatestCommit{SHA: sha, Date: date, Message: message}
		}
	}

	tagsData, err := runGH("api", endpoint+"/tags?per_page=100")
	if err != nil {
		log.Printf("warning: failed to fetch tags for %s/%s: %v", owner, repo.Name, err)
		return
	}
	var tags []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(tagsData, &tags); err != nil {
		log.Printf("warning: failed to parse tags for %s/%s: %v", owner, repo.Name, err)
		return
	}
	repo.LatestVersion = highestStableVersion(tags)
}

func highestStableVersion(tags []struct {
	Name string `json:"name"`
}) string {
	best := ""
	var bestParts [3]string
	for _, tag := range tags {
		match := stableVersion.FindStringSubmatch(tag.Name)
		if match == nil {
			continue
		}
		current := [3]string{match[1], match[2], match[3]}
		if best == "" || greaterVersion(current, bestParts) {
			best, bestParts = tag.Name, current
		}
	}
	return best
}

func greaterVersion(a, b [3]string) bool {
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return len(a[i]) > len(b[i])
		}
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}
