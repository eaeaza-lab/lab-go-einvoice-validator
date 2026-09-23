package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestHistoryRecordsRuns(t *testing.T) {
	db := filepath.Join(t.TempDir(), "h.db")
	ok := writeTemp(t, "ok.json", validDoc)
	bad := writeTemp(t, "bad.json", strings.Replace(validDoc, `"gross":2400`, `"gross":9999`, 1))

	if code, _, errs := runCLI("validate", "--db", db, ok); code != 0 {
		t.Fatalf("code=%d err=%q", code, errs)
	}
	if code, _, errs := runCLI("validate", "--db", db, bad); code != 1 {
		t.Fatalf("code=%d err=%q", code, errs)
	}

	code, out, errs := runCLI("history", "--db", db)
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, errs)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %q", out)
	}
	if !strings.HasPrefix(lines[0], "#2 ") || !strings.Contains(lines[0], "FAIL") || !strings.Contains(lines[0], bad) {
		t.Errorf("newest run wrong: %q", lines[0])
	}
	if !strings.Contains(lines[1], " OK ") || !strings.Contains(lines[1], "diagnostics=0") {
		t.Errorf("oldest run wrong: %q", lines[1])
	}

	_, out, _ = runCLI("history", "--db", db, "--limit", "1")
	if strings.Count(strings.TrimSpace(out), "\n") != 0 {
		t.Errorf("limit not applied: %q", out)
	}
}

func TestHistoryEmptyAndBadLimit(t *testing.T) {
	db := filepath.Join(t.TempDir(), "h.db")
	code, out, _ := runCLI("history", "--db", db)
	if code != 0 || !strings.Contains(out, "no runs recorded") {
		t.Fatalf("code=%d out=%q", code, out)
	}
	if code, _, _ := runCLI("history", "--db", db, "--limit", "-1"); code != 2 {
		t.Fatalf("negative limit: code=%d", code)
	}
}
