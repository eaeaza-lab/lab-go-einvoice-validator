package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validDoc = `{"id":"INV-1","currency":"EUR","lines":[{"description":"Widget","quantity":2,"unit_price":1000,"tax_rate_bp":2000}],"net":2000,"tax":400,"gross":2400}`

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func runCLI(args ...string) (int, string, string) {
	var out, errb bytes.Buffer
	code := run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func TestVersion(t *testing.T) {
	code, out, _ := runCLI("version")
	if code != 0 || !strings.Contains(out, "einvoice "+Version) {
		t.Fatalf("code=%d out=%q", code, out)
	}
}

func TestValidateOK(t *testing.T) {
	p := writeTemp(t, "ok.json", validDoc)
	code, out, _ := runCLI("validate", p)
	if code != 0 || !strings.Contains(out, "OK") {
		t.Fatalf("code=%d out=%q", code, out)
	}
}

func TestValidateDiagnostics(t *testing.T) {
	bad := strings.Replace(validDoc, `"gross":2400`, `"gross":9999`, 1)
	p := writeTemp(t, "bad.json", bad)
	code, out, _ := runCLI("validate", p)
	if code != 1 || !strings.Contains(out, "TOTAL_GROSS") {
		t.Fatalf("code=%d out=%q", code, out)
	}
}

func TestValidateMultipleFilesReportsAll(t *testing.T) {
	ok := writeTemp(t, "ok.json", validDoc)
	bad := writeTemp(t, "bad.json", strings.Replace(validDoc, `"id":"INV-1"`, `"id":""`, 1))
	code, out, _ := runCLI("validate", bad, ok)
	if code != 1 || !strings.Contains(out, "REQ_ID") || !strings.Contains(out, "OK") {
		t.Fatalf("code=%d out=%q", code, out)
	}
}

func TestUsageAndIOErrorsExitTwo(t *testing.T) {
	if code, _, _ := runCLI("validate"); code != 2 {
		t.Fatalf("no args: code=%d, want 2", code)
	}
	if code, _, _ := runCLI("validate", filepath.Join(t.TempDir(), "missing.json")); code != 2 {
		t.Fatalf("missing file: code=%d, want 2", code)
	}
	if code, _, _ := runCLI("validate", writeTemp(t, "junk.json", "{")); code != 2 {
		t.Fatalf("malformed: code=%d, want 2", code)
	}
	if code, _, _ := runCLI("nope"); code != 2 {
		t.Fatalf("unknown command: code=%d, want 2", code)
	}
}
