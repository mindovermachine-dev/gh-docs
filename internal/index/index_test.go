package index

import (
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"

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

	for _, name := range []string{"index.yml", "index.yaml", "index.json", "index.html", "index.mjs"} {
		t.Run(name, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "nested", "dir", name)
			cfg := Merge(Config{}, Config{Owner: "o", File: out})
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

			golden := filepath.Join("testdata", "golden", name)
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
