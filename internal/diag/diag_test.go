package diag

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestLookup(t *testing.T) {
	e, ok := Lookup("TOTAL_GROSS")
	if !ok || e.Summary == "" || e.Fix == "" {
		t.Fatalf("Lookup(TOTAL_GROSS) = %+v, %v", e, ok)
	}
	if _, ok := Lookup("NOPE"); ok {
		t.Error("unknown code found")
	}
}

func TestAllSortedUnique(t *testing.T) {
	all := All()
	for i := 1; i < len(all); i++ {
		if all[i-1].Code >= all[i].Code {
			t.Errorf("not sorted/unique at %q, %q", all[i-1].Code, all[i].Code)
		}
	}
}

// Every code literal used by the packages must be catalogued.
func TestCatalogueCoversSource(t *testing.T) {
	re := regexp.MustCompile(`"([A-Z]+(?:_[A-Z]+)+)"`)
	root := ".."
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == "diag" {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, m := range re.FindAllStringSubmatch(string(b), -1) {
			if _, ok := Lookup(m[1]); !ok {
				t.Errorf("%s: code %s missing from catalogue", path, m[1])
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"XDOC_EXCEEDS_NET", "XDOC_EXCEEDS_TAX", "XDOC_EXCEEDS_GROSS"} {
		if _, ok := Lookup(f); !ok {
			t.Errorf("%s missing", f)
		}
	}
}
