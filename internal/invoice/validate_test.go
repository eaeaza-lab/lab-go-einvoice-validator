package invoice

import "testing"

func good() Invoice {
	return Invoice{
		ID: "INV-0001", Currency: "EUR",
		Lines: []Line{{Description: "Widget", Quantity: 2, UnitPrice: 1000, TaxRateBP: 2000}},
		Net:   2000, Tax: 400, Gross: 2400,
	}
}

func TestValidateOK(t *testing.T) {
	if d := Validate(good()); len(d) != 0 {
		t.Fatalf("expected no diagnostics, got %v", d)
	}
}

func TestValidateBadTotals(t *testing.T) {
	inv := good()
	inv.Gross = 9999
	d := Validate(inv)
	if len(d) != 1 || d[0].Code != "TOTAL_GROSS" {
		t.Fatalf("expected TOTAL_GROSS, got %v", d)
	}
}

func TestValidateLineDiagnostics(t *testing.T) {
	inv := good()
	inv.Lines = append(inv.Lines, Line{Quantity: 0, UnitPrice: -5, TaxRateBP: -1})
	inv.Net, inv.Tax, inv.Gross = 2000, 400, 2400
	d := Validate(inv)
	want := map[string]string{
		"REQ_LINE_DESCRIPTION": "/lines/1/description",
		"LINE_QUANTITY":        "/lines/1/quantity",
		"LINE_UNIT_PRICE":      "/lines/1/unit_price",
		"LINE_TAX_RATE":        "/lines/1/tax_rate_bp",
	}
	got := map[string]string{}
	for _, x := range d {
		got[x.Code] = x.Path
		if x.Fix == "" {
			t.Errorf("%s has no suggested fix", x.Code)
		}
	}
	for code, path := range want {
		if got[code] != path {
			t.Errorf("%s: path %q, want %q (all: %v)", code, got[code], path, d)
		}
	}
}

func TestValidateTotalsFix(t *testing.T) {
	inv := good()
	inv.Net = 1
	d := Validate(inv)
	for _, x := range d {
		if x.Code == "TOTAL_NET" {
			if x.Fix != "set \"net\" to 2000" {
				t.Fatalf("unexpected fix %q", x.Fix)
			}
			return
		}
	}
	t.Fatalf("expected TOTAL_NET, got %v", d)
}

func TestValidateMissingCurrency(t *testing.T) {
	inv := good()
	inv.Currency = ""
	d := Validate(inv)
	if len(d) != 1 || d[0].Code != "REQ_CURRENCY" {
		t.Fatalf("expected REQ_CURRENCY, got %v", d)
	}
}

func TestValidateMissingFields(t *testing.T) {
	d := Validate(Invoice{})
	if len(d) < 2 {
		t.Fatalf("expected required-field diagnostics, got %v", d)
	}
}
