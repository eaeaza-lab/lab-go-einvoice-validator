// Package crossdoc checks consistency between the documents of one batch: a
// credit note must reference an invoice in the batch, agree with it on currency
// and party, and not credit more than the invoice total. It is standalone and
// returns invoice.Diagnostic values.
package crossdoc

import (
	"fmt"

	"example.com/einvoice-validator/internal/invoice"
)

// Document kinds.
const (
	KindInvoice    = "invoice"
	KindCreditNote = "credit_note"
)

// Document is one member of a batch. The invoice model carries no party or
// reference fields, so the caller supplies them here.
type Document struct {
	Name    string          // label used in diagnostic paths, usually the file name
	Kind    string          // KindInvoice or KindCreditNote
	Party   string          // synthetic party identifier (e.g. a tax id)
	Ref     string          // credit notes only: id of the referenced invoice
	Invoice invoice.Invoice // credit notes carry positive amounts
}

// Check returns diagnostics for every credit note in the batch, in batch order.
// Amounts of several credit notes against the same invoice are accumulated, so
// the note that pushes the running total over the invoice is the one reported.
func Check(docs []Document) []invoice.Diagnostic {
	byID := map[string]Document{}
	for _, d := range docs {
		if d.Kind == KindInvoice {
			byID[d.Invoice.ID] = d
		}
	}

	var out []invoice.Diagnostic
	type total struct{ net, tax, gross int64 }
	credited := map[string]total{}

	for _, d := range docs {
		if d.Kind != KindCreditNote {
			continue
		}
		base := "/" + d.Name
		if d.Ref == "" {
			out = append(out, invoice.Diagnostic{
				Code:    "XDOC_REF_MISSING",
				Path:    base + "/ref",
				Message: "credit note does not reference an invoice",
				Fix:     "set the id of the invoice being credited",
			})
			continue
		}
		orig, ok := byID[d.Ref]
		if !ok {
			out = append(out, invoice.Diagnostic{
				Code:    "XDOC_REF_NOT_FOUND",
				Path:    base + "/ref",
				Message: fmt.Sprintf("referenced invoice %q is not in the batch", d.Ref),
				Fix:     "add the invoice to the batch or correct the reference",
			})
			continue
		}
		if d.Invoice.Currency != orig.Invoice.Currency {
			out = append(out, invoice.Diagnostic{
				Code:    "XDOC_CURRENCY",
				Path:    base + "/currency",
				Message: fmt.Sprintf("credit note currency %q differs from invoice currency %q", d.Invoice.Currency, orig.Invoice.Currency),
				Fix:     fmt.Sprintf("use currency %q", orig.Invoice.Currency),
			})
		}
		if d.Party != orig.Party {
			out = append(out, invoice.Diagnostic{
				Code:    "XDOC_PARTY",
				Path:    base + "/party",
				Message: fmt.Sprintf("credit note party %q differs from invoice party %q", d.Party, orig.Party),
				Fix:     fmt.Sprintf("use party %q", orig.Party),
			})
		}

		t := credited[d.Ref]
		t.net += d.Invoice.Net
		t.tax += d.Invoice.Tax
		t.gross += d.Invoice.Gross
		credited[d.Ref] = t
		for _, c := range []struct {
			field      string
			got, limit int64
		}{
			{"net", t.net, orig.Invoice.Net},
			{"tax", t.tax, orig.Invoice.Tax},
			{"gross", t.gross, orig.Invoice.Gross},
		} {
			if c.got > c.limit {
				out = append(out, invoice.Diagnostic{
					Code:    "XDOC_EXCEEDS_" + upper(c.field),
					Path:    base + "/" + c.field,
					Message: fmt.Sprintf("credited %s %d exceeds invoice %s %d", c.field, c.got, c.field, c.limit),
					Fix:     fmt.Sprintf("reduce the credit note %s so the total credited is at most %d", c.field, c.limit),
				})
			}
		}
	}
	return out
}

func upper(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 'a' + 'A'
		}
	}
	return string(b)
}
