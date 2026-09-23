// Package schema is a small stdlib-only checker for a JSON Schema subset:
// type, required, properties, items and enum.
package schema

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
)

//go:embed invoice.schema.json
var invoiceSchema []byte

// Schema is the supported subset of a JSON Schema document.
type Schema struct {
	Type       string             `json:"type"`
	Required   []string           `json:"required"`
	Properties map[string]*Schema `json:"properties"`
	Items      *Schema            `json:"items"`
	Enum       []any              `json:"enum"`
}

// Diagnostic is one schema violation.
type Diagnostic struct {
	Code    string `json:"code"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

// Parse reads a schema document, rejecting unknown keywords and unsupported types.
func Parse(data []byte) (*Schema, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	dec.UseNumber() // enum numbers stay json.Number, matching document values
	var s Schema
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}
	if err := s.check("#"); err != nil {
		return nil, err
	}
	return &s, nil
}

// Invoice returns the embedded invoice schema.
func Invoice() (*Schema, error) { return Parse(invoiceSchema) }

func (s *Schema) check(at string) error {
	switch s.Type {
	case "", "object", "array", "string", "integer", "number", "boolean", "null":
	default:
		return fmt.Errorf("schema %s: unsupported type %q", at, s.Type)
	}
	for k, p := range s.Properties {
		if p == nil {
			return fmt.Errorf("schema %s/properties/%s: null schema", at, k)
		}
		if err := p.check(at + "/properties/" + k); err != nil {
			return err
		}
	}
	if s.Items != nil {
		return s.Items.check(at + "/items")
	}
	return nil
}

// Validate checks a JSON document against the schema. The error is non-nil only
// when data is not parseable JSON; violations are returned as diagnostics.
func (s *Schema) Validate(data []byte) ([]Diagnostic, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("parse document: %w", err)
	}
	if dec.More() {
		return nil, fmt.Errorf("parse document: unexpected content after JSON value")
	}
	var out []Diagnostic
	s.validate(v, "", &out)
	return out, nil
}

func (s *Schema) validate(v any, path string, out *[]Diagnostic) {
	report := func(code, msg string) {
		p := path
		if p == "" {
			p = "/"
		}
		*out = append(*out, Diagnostic{Code: code, Path: p, Message: msg})
	}
	if s.Type != "" && !matches(s.Type, v) {
		report("SCHEMA_TYPE", fmt.Sprintf("expected %s, got %s", s.Type, kind(v)))
		return
	}
	if len(s.Enum) > 0 && !inEnum(s.Enum, v) {
		report("SCHEMA_ENUM", fmt.Sprintf("value %s is not one of the allowed values", render(v)))
	}
	switch t := v.(type) {
	case map[string]any:
		for _, k := range s.Required {
			if _, ok := t[k]; !ok {
				*out = append(*out, Diagnostic{
					Code: "SCHEMA_REQUIRED", Path: path + "/" + k,
					Message: fmt.Sprintf("required key %q is missing", k),
				})
			}
		}
		for _, k := range sortedKeys(t) {
			if p := s.Properties[k]; p != nil {
				p.validate(t[k], path+"/"+k, out)
			}
		}
	case []any:
		if s.Items != nil {
			for i, e := range t {
				s.Items.validate(e, fmt.Sprintf("%s/%d", path, i), out)
			}
		}
	}
}

func matches(typ string, v any) bool {
	switch typ {
	case "object":
		_, ok := v.(map[string]any)
		return ok
	case "array":
		_, ok := v.([]any)
		return ok
	case "string":
		_, ok := v.(string)
		return ok
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "null":
		return v == nil
	case "number":
		_, ok := v.(json.Number)
		return ok
	case "integer":
		n, ok := v.(json.Number)
		if !ok {
			return false
		}
		_, err := n.Int64()
		return err == nil
	}
	return false
}

func kind(v any) string {
	switch t := v.(type) {
	case map[string]any:
		return "object"
	case []any:
		return "array"
	case string:
		return "string"
	case bool:
		return "boolean"
	case nil:
		return "null"
	case json.Number:
		if _, err := t.Int64(); err == nil {
			return "integer"
		}
		return "number"
	}
	return "unknown"
}

// inEnum compares scalars by their canonical JSON rendering.
func inEnum(enum []any, v any) bool {
	want := render(v)
	for _, e := range enum {
		if render(e) == want {
			return true
		}
	}
	return false
}

func render(v any) string {
	switch t := v.(type) {
	case json.Number:
		return t.String()
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
