package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files")

// TestGolden compares CLI output with files in testdata/golden. Run with
// `go test ./cmd/einvoice -run Golden -update` to regenerate them.
func TestGolden(t *testing.T) {
	const data = "../../testdata/"
	cases := []struct {
		name string
		args []string
		code int
	}{
		{"validate_ok.txt", []string{"validate", data + "valid.json"}, 0},
		{"validate_bad_totals.txt", []string{"validate", data + "bad_totals.json"}, 1},
		{"validate_bad_totals.json", []string{"validate", "--format", "json", data + "bad_totals.json"}, 1},
		{"validate_missing_id.txt", []string{"validate", data + "missing_id.json"}, 1},
		{"version.txt", []string{"version"}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, out, _ := runCLI(c.args...)
			if code != c.code {
				t.Fatalf("exit code %d, want %d\n%s", code, c.code, out)
			}
			// Paths are relative to the package directory; keep goldens location-independent.
			out = strings.ReplaceAll(out, data, "testdata/")
			path := filepath.Join("testdata", "golden", c.name)
			if *update {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if got := strings.ReplaceAll(string(want), "\r\n", "\n"); got != out {
				t.Fatalf("output differs from %s\n--- got\n%s\n--- want\n%s", path, out, got)
			}
		})
	}
}
