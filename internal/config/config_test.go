package config

import (
	"path/filepath"
	"testing"
)

type sample struct {
	Owner string `yaml:"owner"`
	Topic string `yaml:"topic"`
	File  string `yaml:"file"`
	Limit int    `yaml:"limit"`
}

func TestSection(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		required bool
		section  string
		want     sample
		wantErr  bool
	}{
		{"reads node", "full.yml", true, "index", sample{"acme", "docs", "out/list.json", 5}, false},
		{"other node is separate", "full.yml", true, "template", sample{}, false},
		{"absent node is empty", "full.yml", true, "nope", sample{}, false},
		{"missing optional file", "missing.yml", false, "index", sample{}, false},
		{"missing required file", "missing.yml", true, "index", sample{}, true},
		{"malformed yaml", "malformed.yml", true, "index", sample{}, true},
		{"wrong value type", "wrongtype.yml", true, "index", sample{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got sample
			err := Section(filepath.Join("testdata", tt.file), tt.required, tt.section, &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
