package main

import (
	"strings"
	"testing"
)

func TestExplainCode(t *testing.T) {
	code, out, errs := runCLI("explain", "total_gross")
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, errs)
	}
	if !strings.Contains(out, "TOTAL_GROSS") || !strings.Contains(out, "fix:") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestExplainList(t *testing.T) {
	code, out, _ := runCLI("explain")
	if code != 0 || !strings.Contains(out, "REQ_ID:") || !strings.Contains(out, "XDOC_PARTY:") {
		t.Fatalf("code=%d out=%q", code, out)
	}
}

func TestExplainUnknown(t *testing.T) {
	if code, _, _ := runCLI("explain", "NOPE"); code != 2 {
		t.Fatalf("code=%d, want 2", code)
	}
}
