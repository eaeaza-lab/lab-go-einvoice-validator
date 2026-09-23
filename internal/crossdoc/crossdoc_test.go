package crossdoc

import (
	"reflect"
	"testing"

	"example.com/einvoice-validator/internal/invoice"
)

func inv(id, cur string, net, tax, gross int64) invoice.Invoice {
	return invoice.Invoice{ID: id, Currency: cur, Net: net, Tax: tax, Gross: gross}
}

func codes(ds []invoice.Diagnostic) []string {
	out := []string{}
	for _, d := range ds {
		out = append(out, d.Code)
	}
	return out
}

func base() Document {
	return Document{Name: "a.json", Kind: KindInvoice, Party: "SY000000000", Invoice: inv("INV-2026-0001", "EUR", 1000, 200, 1200)}
}

func TestValidCreditNote(t *testing.T) {
	cn := Document{Name: "c.json", Kind: KindCreditNote, Party: "SY000000000", Ref: "INV-2026-0001", Invoice: inv("INV-2026-0002", "EUR", 500, 100, 600)}
	if got := Check([]Document{base(), cn}); len(got) != 0 {
		t.Fatalf("want none, got %v", got)
	}
}

func TestMissingAndUnknownRef(t *testing.T) {
	a := Document{Name: "c1", Kind: KindCreditNote}
	b := Document{Name: "c2", Kind: KindCreditNote, Ref: "INV-2026-9999"}
	got := codes(Check([]Document{base(), a, b}))
	want := []string{"XDOC_REF_MISSING", "XDOC_REF_NOT_FOUND"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestCurrencyAndPartyMismatch(t *testing.T) {
	cn := Document{Name: "c", Kind: KindCreditNote, Party: "SY111111111", Ref: "INV-2026-0001", Invoice: inv("INV-2026-0002", "USD", 1, 0, 1)}
	got := codes(Check([]Document{base(), cn}))
	want := []string{"XDOC_CURRENCY", "XDOC_PARTY"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestExceedsOriginal(t *testing.T) {
	cn := Document{Name: "c", Kind: KindCreditNote, Party: "SY000000000", Ref: "INV-2026-0001", Invoice: inv("INV-2026-0002", "EUR", 1001, 201, 1202)}
	got := codes(Check([]Document{base(), cn}))
	want := []string{"XDOC_EXCEEDS_NET", "XDOC_EXCEEDS_TAX", "XDOC_EXCEEDS_GROSS"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestCumulativeCredits(t *testing.T) {
	mk := func(name string) Document {
		return Document{Name: name, Kind: KindCreditNote, Party: "SY000000000", Ref: "INV-2026-0001", Invoice: inv(name, "EUR", 600, 100, 700)}
	}
	got := Check([]Document{base(), mk("c1"), mk("c2")})
	if len(got) == 0 {
		t.Fatal("second credit note should exceed the invoice")
	}
	for _, d := range got {
		if d.Path[:3] != "/c2" {
			t.Fatalf("only c2 should be reported, got %v", d)
		}
	}
}
