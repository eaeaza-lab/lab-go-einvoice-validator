package schema

import (
	"os"
	"strings"
	"testing"
)

func codes(ds []Diagnostic) []string {
	var out []string
	for _, d := range ds {
		out = append(out, d.Code+" "+d.Path)
	}
	return out
}

func mustInvoice(t *testing.T) *Schema {
	t.Helper()
	s, err := Invoice()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestInvoiceSchemaAcceptsValidFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/valid.json")
	if err != nil {
		t.Fatal(err)
	}
	ds, err := mustInvoice(t).Validate(data)
	if err != nil || len(ds) != 0 {
		t.Fatalf("want no diagnostics, got %v (err %v)", ds, err)
	}
}

func TestRequiredMissing(t *testing.T) {
	ds, err := mustInvoice(t).Validate([]byte(`{"currency":"EUR","lines":[],"net":0,"tax":0,"gross":0}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(codes(ds), ","); got != "SCHEMA_REQUIRED /id" {
		t.Fatalf("got %s", got)
	}
}

func TestTypeMismatchAndNestedPaths(t *testing.T) {
	doc := `{"id":"A","currency":"EUR","net":1.5,"tax":0,"gross":0,
		"lines":[{"description":"x","quantity":"2","unit_price":1,"tax_rate_bp":0},{"description":"y"}]}`
	ds, err := mustInvoice(t).Validate([]byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	want := "SCHEMA_TYPE /lines/0/quantity,SCHEMA_REQUIRED /lines/1/quantity," +
		"SCHEMA_REQUIRED /lines/1/unit_price,SCHEMA_REQUIRED /lines/1/tax_rate_bp,SCHEMA_TYPE /net"
	if got := strings.Join(codes(ds), ","); got != want {
		t.Fatalf("got  %s\nwant %s", got, want)
	}
}

func TestRootTypeMismatch(t *testing.T) {
	ds, err := mustInvoice(t).Validate([]byte(`[]`))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(codes(ds), ","); got != "SCHEMA_TYPE /" {
		t.Fatalf("got %s", got)
	}
}

func TestEnum(t *testing.T) {
	s, err := Parse([]byte(`{"type":"object","properties":{
		"cur":{"type":"string","enum":["EUR","USD"]},
		"n":{"type":"integer","enum":[1,2]}}}`))
	if err != nil {
		t.Fatal(err)
	}
	ds, err := s.Validate([]byte(`{"cur":"XXX","n":3}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(codes(ds), ","); got != "SCHEMA_ENUM /cur,SCHEMA_ENUM /n" {
		t.Fatalf("got %s", got)
	}
	ds, err = s.Validate([]byte(`{"cur":"USD","n":2}`))
	if err != nil || len(ds) != 0 {
		t.Fatalf("want valid, got %v (err %v)", ds, err)
	}
}

func TestParseRejectsBadSchemas(t *testing.T) {
	for _, src := range []string{
		`{"type":"blob"}`,
		`{"pattern":"x"}`,
		`{"properties":{"a":null}}`,
		`{`,
	} {
		if _, err := Parse([]byte(src)); err == nil {
			t.Errorf("Parse(%s) succeeded, want error", src)
		}
	}
}

func TestValidateMalformedDocument(t *testing.T) {
	s := mustInvoice(t)
	for _, doc := range []string{`{`, `{} {}`, ``} {
		if _, err := s.Validate([]byte(doc)); err == nil {
			t.Errorf("Validate(%q) succeeded, want error", doc)
		}
	}
}
