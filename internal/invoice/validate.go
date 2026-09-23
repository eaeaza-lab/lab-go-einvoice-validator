package invoice

import "fmt"

// Validate checks required fields and totals, returning diagnostics
// (empty when the invoice is valid).
func Validate(inv Invoice) []Diagnostic {
	var out []Diagnostic
	if inv.ID == "" {
		out = append(out, Diagnostic{"REQ_ID", "id", "invoice id is required"})
	}
	if len(inv.Lines) == 0 {
		out = append(out, Diagnostic{"REQ_LINES", "lines", "at least one line is required"})
	}

	var net, tax int64
	for _, l := range inv.Lines {
		ln := l.Quantity * l.UnitPrice
		net += ln
		tax += (ln*l.TaxRateBP + 5000) / 10000 // round half up
	}
	if inv.Net != net {
		out = append(out, Diagnostic{"TOTAL_NET", "net",
			fmt.Sprintf("net is %d but lines sum to %d", inv.Net, net)})
	}
	if inv.Tax != tax {
		out = append(out, Diagnostic{"TOTAL_TAX", "tax",
			fmt.Sprintf("tax is %d but lines compute to %d", inv.Tax, tax)})
	}
	if inv.Gross != inv.Net+inv.Tax {
		out = append(out, Diagnostic{"TOTAL_GROSS", "gross",
			fmt.Sprintf("gross is %d but net+tax is %d", inv.Gross, inv.Net+inv.Tax)})
	}
	return out
}
