package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mindovermachine-dev/gh-docs/internal/testutil"
)

// run executes the real command tree and returns stdout and the error.
func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func TestIndexCommand(t *testing.T) {
	testutil.FakeGH(t, testutil.ReposFixture(), 0)
	dir := t.TempDir()
	cfg := testutil.WriteFile(t, dir, "cfg.yml", "index:\n  owner: acme\n  topic: Other\n  file: "+filepath.Join(dir, "from-config", "list.json")+"\n")

	t.Run("explicit missing config fails", func(t *testing.T) {
		if _, err := run(t, "index", "-c", filepath.Join(dir, "none.yml"), "-o", "acme"); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("flag overrides config file", func(t *testing.T) {
		out := filepath.Join(dir, "flags", "list.yml")
		stdout, err := run(t, "index", "-c", cfg, "-f", out)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(stdout, "Wrote "+out) {
			t.Errorf("stdout = %q", stdout)
		}
		b, err := os.ReadFile(out)
		if err != nil || !strings.Contains(string(b), "name: beta") || strings.Contains(string(b), "alpha") {
			t.Errorf("unexpected file: %s %v", b, err)
		}
	})

	t.Run("config supplies everything", func(t *testing.T) {
		if _, err := run(t, "index", "-c", cfg); err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(dir, "from-config", "list.json"))
		if err != nil || !strings.Contains(string(b), `"name": "beta"`) {
			t.Errorf("unexpected file: %s %v", b, err)
		}
	})

	t.Run("file and format are mutually exclusive", func(t *testing.T) {
		if _, err := run(t, "index", "-c", cfg, "--file", "x.json", "--format", "json"); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("missing owner", func(t *testing.T) {
		empty := testutil.WriteFile(t, dir, "empty.yml", "")
		if _, err := run(t, "index", "-c", empty); err == nil {
			t.Error("expected error")
		}
	})
}

func TestTemplateCommand(t *testing.T) {
	dir := t.TempDir()
	cfg := testutil.WriteFile(t, dir, "cfg.yml", "template:\n  output: "+filepath.Join(dir, "site")+"\n")

	if _, err := run(t, "template", "-c", cfg); err == nil {
		t.Error("expected error without --jekyll")
	}

	stdout, err := run(t, "template", "--jekyll", "-c", cfg)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(stdout, "Wrote ") != 2 {
		t.Errorf("stdout = %q", stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, "site", "index.md")); err != nil {
		t.Error(err)
	}

	if _, err := run(t, "template", "--jekyll", "-c", cfg); err == nil {
		t.Error("second run without --force should fail")
	}
	if _, err := run(t, "template", "--jekyll", "-c", cfg, "--force"); err != nil {
		t.Errorf("--force run failed: %v", err)
	}

	other := filepath.Join(dir, "other")
	if _, err := run(t, "template", "--jekyll", "-c", cfg, "-o", other); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(other, "_config.yml")); err != nil {
		t.Error("flag should override config output")
	}

	if _, err := run(t, "template", "--jekyll", "-c", filepath.Join(dir, "none.yml")); err == nil {
		t.Error("explicit missing config must fail")
	}
}

func TestVersionCommand(t *testing.T) {
	for _, k := range []string{"BUILD_REPO", "BUILD_REF", "BUILD_SHA", "GITHUB_REF_NAME", "GITHUB_SHA", "GITHUB_SERVER_URL", "GITHUB_REPOSITORY"} {
		t.Setenv(k, "")
	}
	t.Setenv("GITHUB_SERVER_URL", "https://github.com")
	t.Setenv("GITHUB_REPOSITORY", "acme/site")
	t.Setenv("GITHUB_REF_NAME", "main")
	t.Setenv("GITHUB_SHA", "abc123")

	dir := t.TempDir()
	cfg := testutil.WriteFile(t, dir, "cfg.yml", "version:\n  file: "+filepath.Join(dir, "from-config", "v.txt")+"\n")
	want := "Build from https://github.com/acme/site\nRef: main\nCommit: abc123\n"

	t.Run("config supplies file", func(t *testing.T) {
		stdout, err := run(t, "version", "-c", cfg)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(stdout, "Wrote ") {
			t.Errorf("stdout = %q", stdout)
		}
		b, err := os.ReadFile(filepath.Join(dir, "from-config", "v.txt"))
		if err != nil || string(b) != want {
			t.Errorf("got %q, %v", b, err)
		}
	})

	t.Run("flag overrides config", func(t *testing.T) {
		out := filepath.Join(dir, "flag", "v.txt")
		if _, err := run(t, "version", "-c", cfg, "-f", out); err != nil {
			t.Fatal(err)
		}
		if b, err := os.ReadFile(out); err != nil || string(b) != want {
			t.Errorf("got %q, %v", b, err)
		}
	})

	t.Run("default path", func(t *testing.T) {
		testutil.Chdir(t, t.TempDir())
		if _, err := run(t, "version", "-c", filepath.Join(dir, "none.yml")); err == nil {
			t.Fatal("explicit missing config must fail")
		}
		if _, err := run(t, "version"); err != nil {
			t.Fatal(err)
		}
		if b, err := os.ReadFile(filepath.Join("dist", "version.txt")); err != nil || string(b) != want {
			t.Errorf("got %q, %v", b, err)
		}
	})
}

func TestRootHelpListsSubcommands(t *testing.T) {
	out, err := run(t, "--help")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"index", "template", "version"} {
		if !strings.Contains(out, want) {
			t.Errorf("help lacks %q", want)
		}
	}
}
