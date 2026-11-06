package index

import (
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"text/template"

	"github.com/mindovermachine-dev/gh-docs/internal/testutil"
)

var update = flag.Bool("update", false, "rewrite golden files")

func names(repos []Repo) []string {
	out := []string{}
	for _, r := range repos {
		out = append(out, r.Name)
	}
	return out
}

func TestCollect(t *testing.T) {
	testutil.FakeGH(t, testutil.ReposFixture(), 0)

	tests := []struct {
		topic string
		want  []string
	}{
		{"documentation", []string{"alpha", "delta"}},
		{"OTHER", []string{"beta", "delta"}},
		{"unused", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.topic, func(t *testing.T) {
			got, err := Collect(Config{Owner: "o", Visibility: "public", Topic: tt.topic, Limit: 10})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(names(got), tt.want) {
				t.Errorf("got %v, want %v", names(got), tt.want)
			}
		})
	}
}

func TestCollectErrors(t *testing.T) {
	t.Run("gh fails", func(t *testing.T) {
		testutil.FakeGH(t, "", 3)
		if _, err := Collect(Config{Owner: "o"}); err == nil {
			t.Error("expected error")
		}
	})
	t.Run("gh missing", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		if _, err := Collect(Config{Owner: "o"}); err == nil {
			t.Error("expected error")
		}
	})
	t.Run("invalid json", func(t *testing.T) {
		bad := testutil.WriteFile(t, t.TempDir(), "bad.json", "not json")
		testutil.FakeGH(t, bad, 0)
		if _, err := Collect(Config{Owner: "o"}); err == nil {
			t.Error("expected error")
		}
	})
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		name     string
		content  string // empty means the file does not exist
		required bool
		want     Config
		wantErr  bool
	}{
		{"index node", "index:\n  owner: acme\n  topic: docs\n  limit: 5\n", true, Config{Owner: "acme", Topic: "docs", Limit: 5}, false},
		{"page settings", "index:\n  pages-dir: site/docs/repositories\n  template: templates/page.mdx\n  base-url: /repositories\n", true, Config{PagesDir: "site/docs/repositories", Template: "templates/page.mdx", BaseURL: "/repositories"}, false},
		{"other nodes ignored", "template:\n  output: x\n", true, Config{}, false},
		{"missing optional", "", false, Config{}, false},
		{"missing required", "", true, Config{}, true},
		{"file and format conflict", "index:\n  file: a.json\n  format: json\n", true, Config{}, true},
		{"malformed", "index: [", true, Config{}, true},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(dir, "missing.yml")
			if tt.content != "" {
				path = testutil.WriteFile(t, dir, filepath.Join(string(rune('a'+i)), ".docs.yml"), tt.content)
			}
			got, err := LoadConfig(path, tt.required)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name         string
		base, over   Config
		wantFile     string
		wantFormat   string
		wantTopic    string
		wantLimit    int
		wantVisiblty string
	}{
		{"fallbacks", Config{}, Config{}, "docs-repos.yml", "yaml", "documentation", 1000, "public"},
		{"format picks default file ext", Config{}, Config{Format: "html"}, "docs-repos.html", "html", "documentation", 1000, "public"},
		{"file sets format", Config{}, Config{File: "out/x.json"}, "out/x.json", "json", "documentation", 1000, "public"},
		{"file flag beats config format", Config{Format: "html"}, Config{File: "x.yaml"}, "x.yaml", "yaml", "documentation", 1000, "public"},
		{"format flag beats config file", Config{File: "x.json"}, Config{Format: "html"}, "docs-repos.html", "html", "documentation", 1000, "public"},
		{"flags beat config", Config{Topic: "a", Limit: 1, Visibility: "private"}, Config{Topic: "b", Limit: 2, Visibility: "internal"}, "docs-repos.yml", "yaml", "b", 2, "internal"},
		{"config beats fallback", Config{Topic: "a", Limit: 1, Visibility: "private"}, Config{}, "docs-repos.yml", "yaml", "a", 1, "private"},
		{"unknown extension yields no format", Config{}, Config{File: "x.txt"}, "x.txt", "", "documentation", 1000, "public"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Merge(tt.base, tt.over)
			if got.File != tt.wantFile || got.Format != tt.wantFormat || got.Topic != tt.wantTopic ||
				got.Limit != tt.wantLimit || got.Visibility != tt.wantVisiblty {
				t.Errorf("unexpected: %+v", got)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	ok := Config{Owner: "o", Visibility: "public", Format: "yaml"}
	with := func(f func(*Config)) Config { c := ok; f(&c); return c }
	tests := []struct {
		name string
		cfg  Config
		ok   bool
	}{
		{"valid", ok, true},
		{"valid html file", with(func(c *Config) { c.Format, c.File = "html", "x.html" }), true},
		{"valid sidebar file", with(func(c *Config) { c.Format, c.File = "sidebar", "x.mjs" }), true},
		{"valid mdx template", with(func(c *Config) { c.Format, c.Template = "sidebar", "page.mdx" }), true},
		{"bad template extension", with(func(c *Config) { c.Format, c.Template = "sidebar", "page.txt" }), false},
		{"missing owner", with(func(c *Config) { c.Owner = "" }), false},
		{"bad visibility", with(func(c *Config) { c.Visibility = "secret" }), false},
		{"bad format", with(func(c *Config) { c.Format = "xml" }), false},
		{"bad extension", with(func(c *Config) { c.Format, c.File = "", "x.txt" }), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cfg.Validate(); (err == nil) != tt.ok {
				t.Errorf("err = %v, want ok=%v", err, tt.ok)
			}
		})
	}
}

// Golden files live in testdata/golden; refresh them with `go test ./internal/index -update`.
func TestRunGolden(t *testing.T) {
	testutil.FakeGH(t, testutil.ReposFixture(), 0)
	goldenDir, err := filepath.Abs(filepath.Join("testdata", "golden"))
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"index.yml", "index.yaml", "index.json", "index.html", "index.mjs"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			out := filepath.Join(dir, "nested", "dir", name)
			cfg := Merge(Config{}, Config{Owner: "o", File: out})
			cfg.PagesDir = filepath.Join(dir, "pages")
			cfg.BaseURL = "/repo"
			if name == "index.mjs" {
				testutil.Chdir(t, dir)
				testutil.WriteFile(t, dir, filepath.Join("docs", "repository-template.md"), "# [% .Name %]\n")
			}
			path, err := Run(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if path != out {
				t.Errorf("path = %s, want %s", path, out)
			}
			got, err := os.ReadFile(out)
			if err != nil {
				t.Fatal(err)
			}

			golden := filepath.Join(goldenDir, name)
			if *update {
				testutil.WriteFile(t, "", golden, string(got))
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Errorf("output differs from %s:\n%s", golden, got)
			}
		})
	}
}

func TestRunDefaultsToFormatExtension(t *testing.T) {
	testutil.FakeGH(t, testutil.ReposFixture(), 0)
	testutil.Chdir(t, t.TempDir())

	path, err := Run(Merge(Config{}, Config{Owner: "o", Format: "json"}))
	if err != nil {
		t.Fatal(err)
	}
	if path != "docs-repos.json" {
		t.Errorf("path = %s", path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Error(err)
	}
}

func TestRunSidebarFormatUsesPairedDefaults(t *testing.T) {
	testutil.FakeGH(t, testutil.ReposFixture(), 0)
	dir := t.TempDir()
	testutil.Chdir(t, dir)
	if err := StarlightTemplate(filepath.Join("docs", "repository-template.md"), false); err != nil {
		t.Fatal(err)
	}
	path, err := Run(Merge(Config{}, Config{Owner: "o", Format: "sidebar"}))
	if err != nil {
		t.Fatal(err)
	}
	if path != "docs-repos.mjs" {
		t.Fatalf("path = %s", path)
	}
	if _, err := os.Stat(filepath.Join("docs", "content", "repo", "alpha.md")); err != nil {
		t.Fatal(err)
	}
	alpha, err := os.ReadFile(filepath.Join("docs", "content", "repo", "alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	delta, err := os.ReadFile(filepath.Join("docs", "content", "repo", "delta.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(alpha), "[📚 **Documentation**](<https://example.github.io/alpha>)") ||
		!strings.Contains(string(alpha), "[💻 **GitHub repository**](<https://github.com/o/alpha>)") {
		t.Errorf("missing repository actions: %s", alpha)
	}
	if strings.Contains(string(delta), "**Documentation**") ||
		!strings.Contains(string(delta), "[💻 **GitHub repository**](<https://github.com/o/delta>)") {
		t.Errorf("unexpected actions without a homepage: %s", delta)
	}
}

func TestRunSidebarPageExtensionMatchesTemplate(t *testing.T) {
	for _, extension := range []string{".md", ".mdx"} {
		t.Run(extension, func(t *testing.T) {
			testutil.FakeGH(t, testutil.ReposFixture(), 0)
			dir := t.TempDir()
			testutil.Chdir(t, dir)
			testutil.WriteFile(t, dir, filepath.Join("docs", "repository-template"+extension), "# [% .Name %]\n")

			cfg := Merge(Config{}, Config{Owner: "o", Format: "sidebar"})
			if _, err := Run(cfg); err != nil {
				t.Fatal(err)
			}
			pagePath := filepath.Join(cfg.PagesDir, "alpha"+extension)
			if _, err := os.Stat(pagePath); err != nil {
				t.Fatalf("generated page %s: %v", pagePath, err)
			}
		})
	}
}

func TestRunSidebarUsesCustomTemplateOutsideDocs(t *testing.T) {
	testutil.FakeGH(t, testutil.ReposFixture(), 0)
	dir := t.TempDir()
	testutil.Chdir(t, dir)
	templatePath := testutil.WriteFile(t, dir, filepath.Join("templates", "repository.mdx"), "# [% .Name %]\n")
	cfg := Merge(Config{}, Config{
		Owner: "o", Format: "sidebar", PagesDir: filepath.Join(dir, "sample", "content", "repo"),
		Template: templatePath,
	})
	if _, err := Run(cfg); err != nil {
		t.Fatal(err)
	}
	pagePath := filepath.Join(cfg.PagesDir, "alpha.mdx")
	page, err := os.ReadFile(pagePath)
	if err != nil {
		t.Fatalf("generated MDX page %s: %v", pagePath, err)
	}
	if string(page) != "# alpha\n" {
		t.Errorf("page = %q", page)
	}
}

func TestRunSidebarCreatesPagesDirectoryWithNoMatches(t *testing.T) {
	testutil.FakeGH(t, testutil.ReposFixture(), 0)
	dir := t.TempDir()
	testutil.Chdir(t, dir)
	testutil.WriteFile(t, dir, filepath.Join("docs", "repository-template.md"), "# Repositories\n")
	pagesDir := filepath.Join(dir, "pages")
	cfg := Merge(Config{}, Config{
		Owner: "o", Topic: "no-matches", Format: "sidebar",
		File: filepath.Join(dir, "sidebar.mjs"), PagesDir: pagesDir,
	})
	if _, err := Run(cfg); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(pagesDir); err != nil || !info.IsDir() {
		t.Fatalf("pages directory was not created: %v", err)
	}
}

func TestRunSidebarGeneratesPagesAndLocalSidebar(t *testing.T) {
	testutil.FakeGH(t, testutil.ReposFixture(), 0)
	dir := t.TempDir()
	testutil.Chdir(t, dir)
	testutil.WriteFile(t, dir, filepath.Join("docs", "repository-template.md"), "# [% .Owner %]/[% .Name %]\n[% yaml .Description %]\n")
	cfg := Merge(Config{}, Config{
		Owner: "o", File: filepath.Join(dir, "sidebar.mjs"),
		PagesDir: filepath.Join(dir, "pages"),
		BaseURL:  "/repositories",
	})
	if _, err := Run(cfg); err != nil {
		t.Fatal(err)
	}
	page, err := os.ReadFile(filepath.Join(dir, "pages", "alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(page) != "# o/alpha\nAlpha docs\n" {
		t.Errorf("page = %q", page)
	}
	sidebar, err := os.ReadFile(cfg.File)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(sidebar), `{ label: "alpha", link: "/repositories/alpha/" }`) ||
		strings.Contains(string(sidebar), "https://") {
		t.Errorf("unexpected sidebar: %s", sidebar)
	}
}

func TestStarlightTemplate(t *testing.T) {
	dir := t.TempDir()
	destination := filepath.Join(dir, "nested", "template.md")
	if err := StarlightTemplate(destination, false); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("templates/starlight.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatal("scaffold differs from maintained template")
	}
	if err := StarlightTemplate(destination, false); err == nil {
		t.Fatal("expected existing destination error")
	}
	// #nosec G304 -- destination is inside the temporary test directory.
	afterRefusal, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(afterRefusal) != string(want) {
		t.Fatal("existing template was changed without --force")
	}
	if err := StarlightTemplate(destination, true); err != nil {
		t.Fatalf("force overwrite failed: %v", err)
	}
}

func TestSidebarTemplateErrorsIdentifyRepository(t *testing.T) {
	testutil.FakeGH(t, testutil.ReposFixture(), 0)
	dir := t.TempDir()
	testutil.Chdir(t, dir)
	testutil.WriteFile(t, dir, filepath.Join("docs", "repository-template.md"), "[% .UnknownField %]")
	cfg := Merge(Config{}, Config{
		Owner: "o", File: filepath.Join(dir, "sidebar.mjs"),
		PagesDir: filepath.Join(dir, "pages"),
	})
	_, err := Run(cfg)
	if err == nil || !strings.Contains(err.Error(), filepath.Join("docs", "repository-template.md")) || !strings.Contains(err.Error(), "repository alpha") {
		t.Fatalf("error = %v", err)
	}
	if _, err := os.Stat(cfg.File); !os.IsNotExist(err) {
		t.Fatalf("sidebar should not be written after render failure, stat error = %v", err)
	}
}

func TestSlugAndPageURL(t *testing.T) {
	if got := slug("../My Repo__"); got != "my-repo" {
		t.Errorf("slug = %q", got)
	}
	if got := PageURL("astro/src/content/docs/repositories", "", "alpha"); got != "/repositories/alpha/" {
		t.Errorf("page URL = %q", got)
	}
	if got := PageURL("docs/content/repo", "/catalog/", "alpha"); got != "/catalog/alpha/" {
		t.Errorf("configured page URL = %q", got)
	}
}

func TestSlugCollisionsFailBeforeWriting(t *testing.T) {
	dir := t.TempDir()
	testutil.Chdir(t, dir)
	testutil.WriteFile(t, dir, filepath.Join("docs", "repository-template.md"), "# [% .Name %]\n")
	tmpl, err := loadPageTemplate("")
	if err != nil {
		t.Fatal(err)
	}
	cfg := Merge(Config{}, Config{
		Format: "sidebar", File: filepath.Join(dir, "sidebar.mjs"),
		PagesDir: filepath.Join(dir, "pages"),
	})
	_, err = writePagesAndSidebar(cfg, []Repo{{Name: "repo.one"}, {Name: "repo-one"}}, tmpl)
	if err == nil || !strings.Contains(err.Error(), "same page slug") {
		t.Fatalf("error = %v", err)
	}
	if _, err := os.Stat(cfg.File); !os.IsNotExist(err) {
		t.Fatalf("sidebar should not be written, stat error = %v", err)
	}
}

func TestHighestStableVersion(t *testing.T) {
	tags := []struct {
		Name string `json:"name"`
	}{{"v1.9.0"}, {"v1.10.0"}, {"v2.0.0-rc.1"}, {"2.0.0+build.1"}, {"latest"}}
	if got := highestStableVersion(tags); got != "2.0.0+build.1" {
		t.Errorf("version = %q, want 2.0.0+build.1", got)
	}
}

func TestCollectRepoWorkflows(t *testing.T) {
	previousRunGH := runGH
	runGH = func(args ...string) ([]byte, error) {
		want := []string{"api", "--paginate", "--jq", `.workflows[] | select(.state == "active") | .path`, "repos/acme/widget/actions/workflows"}
		if !reflect.DeepEqual(args, want) {
			t.Errorf("gh args = %#v, want %#v", args, want)
		}
		return []byte(".github/workflows/release.yml\n.github/workflows/ci.yml\n"), nil
	}
	t.Cleanup(func() { runGH = previousRunGH })

	repo := Repo{Name: "widget"}
	collectRepoWorkflows("acme", &repo)
	if !reflect.DeepEqual(repo.Workflows, map[string]bool{"release.yml": true, "ci.yml": true}) {
		t.Errorf("workflows = %#v", repo.Workflows)
	}
}

func TestWorkflowTemplateCondition(t *testing.T) {
	tmpl, err := template.New("workflow").Delims("[%", "%]").Parse(`[% if index .Workflows "release.yml" %]release badge[% end %]`)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name      string
		workflows map[string]bool
		want      string
	}{
		{"workflow exists", map[string]bool{"release.yml": true}, "release badge"},
		{"workflow missing", map[string]bool{"ci.yml": true}, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var rendered strings.Builder
			if err := tmpl.Execute(&rendered, Repo{Workflows: tt.workflows}); err != nil {
				t.Fatal(err)
			}
			if rendered.String() != tt.want {
				t.Errorf("rendered = %q, want %q", rendered.String(), tt.want)
			}
		})
	}
}

func TestConditionalTrimMarkersWithCustomDelimiters(t *testing.T) {
	tmpl, err := template.New("trim").Delims("[%", "%]").Parse("before\n[% if .Present -%]\noptional\n[% end -%]\nafter")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name    string
		present bool
		want    string
	}{
		{"false branch", false, "before\nafter"},
		{"true branch", true, "before\noptional\nafter"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var rendered strings.Builder
			if err := tmpl.Execute(&rendered, map[string]bool{"Present": tt.present}); err != nil {
				t.Fatal(err)
			}
			if rendered.String() != tt.want {
				t.Errorf("rendered = %q, want %q", rendered.String(), tt.want)
			}
		})
	}
}

func TestMergePageOptions(t *testing.T) {
	base := Config{PagesDir: "configured/pages", BaseURL: "/configured"}
	overrides := Config{PagesDir: "flag/pages", BaseURL: "/flag"}
	got := Merge(base, overrides)
	if got.PagesDir != "flag/pages" || got.BaseURL != "/flag" {
		t.Errorf("flag options did not override config: %+v", got)
	}
	base.Template = "configured/repository.md"
	overrides.Template = "flag/repository.mdx"
	got = Merge(base, overrides)
	if got.Template != "flag/repository.mdx" {
		t.Errorf("template flag did not override config: %+v", got)
	}
	got = Merge(Config{}, Config{})
	if got.PagesDir != "docs/content/repo" {
		t.Errorf("default pages directory = %q", got.PagesDir)
	}
}

func TestSingleFileFormatsDoNotFetchSidebarMetadata(t *testing.T) {
	dir := t.TempDir()
	scriptDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(scriptDir, 0o755); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(dir, "calls")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$GH_CALL_LOG\"\ncat \"$GH_FIXTURE\"\n"
	scriptPath := filepath.Join(scriptDir, "gh")
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", scriptDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GH_CALL_LOG", logPath)
	t.Setenv("GH_FIXTURE", testutil.ReposFixture())

	for _, format := range []string{"yaml", "json", "html"} {
		t.Run(format, func(t *testing.T) {
			if err := os.WriteFile(logPath, nil, 0o600); err != nil {
				t.Fatal(err)
			}
			cfg := Merge(Config{}, Config{
				Owner: "o", Format: format, File: filepath.Join(dir, "out."+extForFormat(format)),
			})
			if _, err := Run(cfg); err != nil {
				t.Fatal(err)
			}
			calls, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatal(err)
			}
			if lines := strings.Split(strings.TrimSpace(string(calls)), "\n"); len(lines) != 1 || !strings.HasPrefix(lines[0], "repo list ") {
				t.Errorf("unexpected gh calls: %q", calls)
			}
		})
	}
}

func TestRunErrors(t *testing.T) {
	testutil.FakeGH(t, testutil.ReposFixture(), 0)
	dir := t.TempDir()
	blocker := testutil.WriteFile(t, dir, "blocker", "i am a file")

	tests := []struct {
		name string
		cfg  Config
	}{
		{"invalid config", Merge(Config{}, Config{})},
		{"parent is a file", Merge(Config{}, Config{Owner: "o", File: filepath.Join(blocker, "x.yml")})},
		{"target is a directory", Merge(Config{}, Config{Owner: "o", File: dir + ".yml"})},
	}
	if err := os.Mkdir(dir+".yml", 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(dir + ".yml") })

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Run(tt.cfg); err == nil {
				t.Error("expected error")
			}
		})
	}
}
