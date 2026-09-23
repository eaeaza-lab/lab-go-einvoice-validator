package main

import (
	"strings"
	"testing"
)

func TestDemo(t *testing.T) {
	code, out, errs := runCLI("demo")
	if code != 0 {
		t.Fatalf("code=%d out=%q err=%q", code, out, errs)
	}
	for _, want := range []string{"valid.json: OK", "bad_totals.json: TOTAL_GROSS", "missing_id.json: REQ_ID"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "UNEXPECTED") {
		t.Errorf("unexpected fixture result:\n%s", out)
	}
}

func TestDemoRejectsArgs(t *testing.T) {
	if code, _, _ := runCLI("demo", "extra"); code != 2 {
		t.Fatalf("code=%d, want 2", code)
	}
}
