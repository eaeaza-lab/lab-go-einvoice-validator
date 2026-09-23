package invoice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Load reads an invoice from a JSON file. Unknown fields and trailing
// content are rejected.
func Load(path string) (Invoice, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Invoice{}, fmt.Errorf("read %s: %w", path, err)
	}
	inv, err := ParseJSON(data)
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
