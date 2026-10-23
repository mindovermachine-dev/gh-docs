package version

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mindovermachine-dev/gh-docs/internal/testutil"
)

// clearEnv blanks every variable Resolve reads so the host environment can't leak in.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"BUILD_REPO", "BUILD_REF", "BUILD_SHA", "GITHUB_REF_NAME", "GITHUB_SHA", "GITHUB_SERVER_URL", "GITHUB_REPOSITORY"} {
		t.Setenv(k, "")
	}
}

// newRepo creates a real git repository with one commit and returns its HEAD SHA.
func newRepo(t *testing.T, branch, origin string) string {
	t.Helper()
	dir := t.TempDir()
	testutil.Chdir(t, dir)
	run := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("init", "-q", "-b", branch)
	run("remote", "add", "origin", origin)
	run("commit", "-q", "--allow-empty", "-m", "init")
	return run("rev-parse", "HEAD")
}

func TestResolveFromGit(t *testing.T) {
	clearEnv(t)
	sha := newRepo(t, "feature", "https://example.com/o/r.git")

	got, err := Resolve()
	if err != nil {
		t.Fatal(err)
	}
	want := Info{Repo: "https://example.com/o/r.git", Ref: "feature", SHA: sha}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestResolveDetachedHead(t *testing.T) {
	clearEnv(t)
	sha := newRepo(t, "main", "https://example.com/o/r.git")
	if out, err := exec.Command("git", "checkout", "-q", "--detach", sha).CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}

	got, err := Resolve()
	if err != nil || got.Ref != "detached" {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestResolveEnvPrecedence(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want Info
	}{
		{
			"github actions",
			map[string]string{"GITHUB_SERVER_URL": "https://github.com/", "GITHUB_REPOSITORY": "o/r", "GITHUB_REF_NAME": "v1", "GITHUB_SHA": "abc"},
			Info{Repo: "https://github.com/o/r", Ref: "v1", SHA: "abc"},
		},
		{
			"build vars beat github vars",
			map[string]string{"BUILD_REPO": "r1", "BUILD_REF": "b1", "BUILD_SHA": "s1", "GITHUB_SERVER_URL": "https://github.com", "GITHUB_REPOSITORY": "o/r", "GITHUB_REF_NAME": "v1", "GITHUB_SHA": "abc"},
			Info{Repo: "r1", Ref: "b1", SHA: "s1"},
		},
		{
			"mix of env and git",
			map[string]string{"BUILD_REF": "custom"},
			Info{Repo: "https://example.com/o/r.git", Ref: "custom"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			sha := newRepo(t, "main", "https://example.com/o/r.git")
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			if tt.want.SHA == "" {
				tt.want.SHA = sha
			}
			got, err := Resolve()
			if err != nil || got != tt.want {
				t.Errorf("got %+v (%v), want %+v", got, err, tt.want)
			}
		})
	}
}

func TestResolveFailsOutsideGitWithoutEnv(t *testing.T) {
	clearEnv(t)
	testutil.Chdir(t, t.TempDir())
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(t.TempDir()))
	if _, err := Resolve(); err == nil {
		t.Error("expected error")
	}
}

func TestMerge(t *testing.T) {
	if got := Merge(Config{}, Config{}).File; got != "dist/version.txt" {
		t.Errorf("fallback = %s", got)
	}
	if got := Merge(Config{File: "a.txt"}, Config{}).File; got != "a.txt" {
		t.Errorf("config = %s", got)
	}
	if got := Merge(Config{File: "a.txt"}, Config{File: "b.txt"}).File; got != "b.txt" {
		t.Errorf("flag = %s", got)
	}
}

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	good := testutil.WriteFile(t, dir, "good.yml", "version:\n  file: out/v.txt\n")
	if got, err := LoadConfig(good, true); err != nil || got.File != "out/v.txt" {
		t.Errorf("got %+v, %v", got, err)
	}
	if _, err := LoadConfig(filepath.Join(dir, "none.yml"), true); err == nil {
		t.Error("missing required should fail")
	}
	if got, err := LoadConfig(filepath.Join(dir, "none.yml"), false); err != nil || got.File != "" {
		t.Errorf("missing optional: %+v, %v", got, err)
	}
}

func TestRunWritesFile(t *testing.T) {
	clearEnv(t)
	t.Setenv("BUILD_REPO", "https://example.com/o/r")
	t.Setenv("BUILD_REF", "main")
	t.Setenv("BUILD_SHA", "abc123")
	out := filepath.Join(t.TempDir(), "nested", "version.txt")

	path, err := Run(Config{File: out})
	if err != nil || path != out {
		t.Fatalf("path %s, err %v", path, err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	want := "Build from https://example.com/o/r\nRef: main\nCommit: abc123\n"
	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRunErrors(t *testing.T) {
	clearEnv(t)
	t.Setenv("BUILD_REPO", "r")
	t.Setenv("BUILD_REF", "b")
	t.Setenv("BUILD_SHA", "s")
	blocker := testutil.WriteFile(t, t.TempDir(), "blocker", "file")

	if _, err := Run(Config{File: filepath.Join(blocker, "v.txt")}); err == nil {
		t.Error("expected error when parent is a file")
	}

	t.Setenv("BUILD_REPO", "")
	t.Setenv("BUILD_REF", "")
	t.Setenv("BUILD_SHA", "")
	testutil.Chdir(t, t.TempDir())
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(t.TempDir()))
	if _, err := Run(Config{File: "v.txt"}); err == nil {
		t.Error("expected error when nothing can be resolved")
	}
}
