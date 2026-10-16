package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mindovermachine-dev/gh-docs/internal/testutil"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	good := testutil.WriteFile(t, dir, "good.yml", "template:\n  output: site\nindex:\n  owner: x\n")
	bad := testutil.WriteFile(t, dir, "bad.yml", "template: [")

	got, err := LoadConfig(good, true)
	if err != nil || got.Output != "site" {
		t.Errorf("got %+v, %v", got, err)
	}
	if got, err := LoadConfig(filepath.Join(dir, "none.yml"), false); err != nil || got != (Config{}) {
		t.Errorf("missing optional: %+v, %v", got, err)
	}
	if _, err := LoadConfig(filepath.Join(dir, "none.yml"), true); err == nil {
		t.Error("missing required should fail")
	}
	if _, err := LoadConfig(bad, true); err == nil {
		t.Error("malformed should fail")
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name       string
		base, over Config
		want       string
	}{
		{"fallback", Config{}, Config{}, "docs"},
		{"config beats fallback", Config{Output: "a"}, Config{}, "a"},
		{"flag beats config", Config{Output: "a"}, Config{Output: "b"}, "b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Merge(tt.base, tt.over).Output; got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

// The embedded sources in ./jekyll are the single source of truth for the expected output.
func TestJekyllCopiesEmbeddedFiles(t *testing.T) {
	out := filepath.Join(t.TempDir(), "deep", "docz")
	written, err := Jekyll(Config{Output: out}, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(written) != 2 {
		t.Fatalf("written = %v", written)
	}
	for _, name := range []string{"index.md", "_config.yml"} {
		want, err := os.ReadFile(filepath.Join("jekyll", name))
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join(out, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Errorf("%s differs from template source", name)
		}
	}
}

func TestJekyllExistingFiles(t *testing.T) {
	tests := []struct {
		name     string
		existing []string
		force    bool
		wantErr  bool
	}{
		{"none exist", nil, false, false},
		{"index exists", []string{"index.md"}, false, true},
		{"config exists", []string{"_config.yml"}, false, true},
		{"both exist, forced", []string{"index.md", "_config.yml"}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := t.TempDir()
			for _, f := range tt.existing {
				testutil.WriteFile(t, out, f, "mine")
			}
			_, err := Jekyll(Config{Output: out}, tt.force)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}

			idx, _ := os.ReadFile(filepath.Join(out, "index.md"))
			if tt.wantErr {
				if err != nil && !strings.Contains(err.Error(), "--force") {
					t.Errorf("error should mention --force: %v", err)
				}
				// Nothing is overwritten or partially written.
				for _, f := range []string{"index.md", "_config.yml"} {
					b, rerr := os.ReadFile(filepath.Join(out, f))
					exists := false
					for _, e := range tt.existing {
						exists = exists || e == f
					}
					if exists && string(b) != "mine" {
						t.Errorf("%s was overwritten", f)
					}
					if !exists && rerr == nil {
						t.Errorf("%s was written despite conflict", f)
					}
				}
			} else if string(idx) == "mine" {
				t.Error("index.md not overwritten")
			}
		})
	}
}

func TestJekyllErrors(t *testing.T) {
	blocker := testutil.WriteFile(t, t.TempDir(), "blocker", "file")
	if _, err := Jekyll(Config{Output: filepath.Join(blocker, "sub")}, false); err == nil {
		t.Error("expected error when output parent is a file")
	}
}
