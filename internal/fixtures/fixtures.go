// Package fixtures embeds the synthetic sample invoices and the diagnostic
// codes each one is expected to produce.
package fixtures

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
)

//go:embed data/*.json
var files embed.FS

const expectedFile = "expected.json"

// Fixture is one embedded document with the diagnostic codes Validate should emit.
type Fixture struct {
	Name     string
	Data     []byte
	Expected []string // diagnostic codes in emission order; empty means valid
}

// All returns every embedded invoice fixture, sorted by name.
func All() ([]Fixture, error) {
	raw, err := files.ReadFile("data/" + expectedFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", expectedFile, err)
	}
	var expected map[string][]string
	if err := json.Unmarshal(raw, &expected); err != nil {
		return nil, fmt.Errorf("parse %s: %w", expectedFile, err)
	}
	entries, err := files.ReadDir("data")
	if err != nil {
		return nil, err
	}
	var out []Fixture
	for _, e := range entries {
		name := e.Name()
		if name == expectedFile {
			continue
		}
		codes, ok := expected[name]
		if !ok {
			return nil, fmt.Errorf("fixture %s has no entry in %s", name, expectedFile)
		}
		data, err := files.ReadFile("data/" + name)
		if err != nil {
			return nil, err
		}
		out = append(out, Fixture{Name: name, Data: data, Expected: codes})
	}
	for name := range expected {
		if _, err := files.ReadFile("data/" + name); err != nil {
			return nil, fmt.Errorf("%s lists %s but the file is not embedded", expectedFile, name)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
