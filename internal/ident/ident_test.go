package ident

import "testing"

func TestCheckInvoiceID(t *testing.T) {
	for _, id := range []string{"INV-2026-0001", "INV-1999-12345678", ""} {
		if d := CheckInvoiceID(id); len(d) != 0 {
			t.Errorf("%q: unexpected %v", id, d)
		}
	}
	for _, id := range []string{"inv-2026-0001", "INV-26-0001", "INV-2026-1", "2026-0001", "INV-2026-0001 "} {
		d := CheckInvoiceID(id)
		if len(d) != 1 || d[0].Code != "ID_FORMAT" || d[0].Path != "/id" || d[0].Fix == "" {
			t.Errorf("%q: got %v", id, d)
		}
	}
}

func TestCheckCurrency(t *testing.T) {
	for _, c := range []string{"EUR", "USD", ""} {
		if d := CheckCurrency(c); len(d) != 0 {
			t.Errorf("%q: unexpected %v", c, d)
		}
	}
	for _, c := range []string{"eur", "XXX", "EURO", "€"} {
		d := CheckCurrency(c)
		if len(d) != 1 || d[0].Code != "CURRENCY_UNKNOWN" || d[0].Path != "/currency" {
			t.Errorf("%q: got %v", c, d)
		}
	}
}

func TestTaxIDCheckDigit(t *testing.T) {
	// 1*1+2*2+3*3+4*4+5*5+6*6+7*7+8*8 = 204 -> 4
	if d, ok := TaxIDCheckDigit("12345678"); !ok || d != '4' {
		t.Errorf("got %c %v", d, ok)
	}
	for _, p := range []string{"1234567", "1234567a", ""} {
		if _, ok := TaxIDCheckDigit(p); ok {
			t.Errorf("%q should be rejected", p)
		}
	}
}

func TestCheckTaxID(t *testing.T) {
	if d := CheckTaxID("/seller/tax_id", "SY123456784"); len(d) != 0 {
		t.Errorf("valid id rejected: %v", d)
	}
	d := CheckTaxID("/seller/tax_id", "SY123456785")
	if len(d) != 1 || d[0].Code != "TAXID_CHECKSUM" || d[0].Fix != "replace the last digit with 4" {
		t.Errorf("checksum: got %v", d)
	}
	for _, id := range []string{"", "XX123456784", "SY12345678", "SY1234567844", "SY1234567x4", "SY12345678x"} {
		d := CheckTaxID("/seller/tax_id", id)
		if len(d) != 1 || d[0].Code != "TAXID_FORMAT" || d[0].Path != "/seller/tax_id" {
			t.Errorf("%q: got %v", id, d)
		}
	}
}
