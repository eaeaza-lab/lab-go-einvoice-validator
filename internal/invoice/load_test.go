package invoice

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "inv.json")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadOK(t *testing.T) {
	p := writeTemp(t, `{"id":"INV-0001","currency":"EUR","lines":[{"description":"Widget","quantity":2,"unit_price":1000,"tax_rate_bp":2000}],"net":2000,"tax":400,"gross":2400}`)
	inv, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if inv.ID != "INV-0001" || len(inv.Lines) != 1 || inv.Gross != 2400 {
		t.Fatalf("unexpected invoice: %+v", inv)
	}
	if d := Validate(inv); len(d) != 0 {
		t.Fatalf("expected valid, got %v", d)
	}
}

func TestLoadUnknownField(t *testing.T) {
	p := writeTemp(t, `{"id":"INV-1","bogus":1}`)
	_, err := Load(p)
	if err == nil || !strings.Contains(err.Error(), "bogus") {
		t.Fatalf("expected unknown-field error, got %v", err)
	}
}

func TestLoadMalformed(t *testing.T) {
	if _, err := Load(writeTemp(t, `{"id":`)); err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestLoadTrailingContent(t *testing.T) {
	if _, err := Load(writeTemp(t, `{"id":"A"} {"id":"B"}`)); err == nil {
		t.Fatal("expected error for trailing content")
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "nope.json")); err == nil {
		t.Fatal("expected error for missing file")
	}
}
