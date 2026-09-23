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

func TestValidateMissingFields(t *testing.T) {
	d := Validate(Invoice{})
	if len(d) < 2 {
		t.Fatalf("expected required-field diagnostics, got %v", d)
	}
}
