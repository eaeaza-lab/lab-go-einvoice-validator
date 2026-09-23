package invoice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Load reads an invoice from a file: XML when the extension is .xml
// (case-insensitive), JSON otherwise. Unknown fields and trailing content
// are rejected in both formats.
func Load(path string) (Invoice, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Invoice{}, fmt.Errorf("read %s: %w", path, err)
	}
	parse := ParseJSON
	if strings.EqualFold(filepath.Ext(path), ".xml") {
		parse = ParseXML
	}
	inv, err := parse(data)
	if err != nil {
		return Invoice{}, fmt.Errorf("%s: %w", path, err)
	}
	return inv, nil
}

// ParseJSON decodes a JSON invoice document, rejecting unknown fields.
func ParseJSON(data []byte) (Invoice, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var inv Invoice
	if err := dec.Decode(&inv); err != nil {
		return Invoice{}, fmt.Errorf("invalid JSON invoice: %w", err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return Invoice{}, fmt.Errorf("invalid JSON invoice: unexpected content after document")
	}
	return inv, nil
}
