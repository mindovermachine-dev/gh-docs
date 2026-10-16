// Package testutil holds shared fixtures and helpers for tests.
package testutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// ReposFixture returns the path of the shared `gh repo list` JSON fixture.
func ReposFixture() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", "repos.json")
}

// FakeGH puts a `gh` executable first on PATH. It prints the fixture file when
// exitCode is 0, otherwise it fails with that exit code.
func FakeGH(t *testing.T, fixture string, exitCode int) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake gh script requires a POSIX shell")
	}
	script := fmt.Sprintf("#!/bin/sh\ncat '%s'\n", fixture)
	if exitCode != 0 {
		script = fmt.Sprintf("#!/bin/sh\necho boom >&2\nexit %d\n", exitCode)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "gh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// Chdir changes the working directory for the test and restores it afterwards.
func Chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
}

// WriteFile writes content to dir/name, creating parents, and returns the path.
func WriteFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}
