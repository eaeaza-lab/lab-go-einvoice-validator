package invoice

import (
	"fmt"
	"strings"
)

// Validate checks required fields and totals, returning diagnostics
// (empty when the invoice is valid). Paths are JSON-pointer-like
// (for example "/lines/0/quantity") and every diagnostic carries a suggested fix.
func Validate(inv Invoice) []Diagnostic {
	var out []Diagnostic
	add := func(code, path, msg, fix string) {
		out = append(out, Diagnostic{Code: code, Path: path, Message: msg, Fix: fix})
	}

	if strings.TrimSpace(inv.ID) == "" {
		add("REQ_ID", "/id", "invoice id is required", "set \"id\" to a non-empty invoice number")
	}
	if strings.TrimSpace(inv.Currency) == "" {
		add("REQ_CURRENCY", "/currency", "currency is required", "set \"currency\" to a 3-letter code such as \"EUR\"")
	}
	if len(inv.Lines) == 0 {
		add("REQ_LINES", "/lines", "at least one line is required", "add at least one entry to \"lines\"")
	}

	var net, tax int64
	for i, l := range inv.Lines {
		p := fmt.Sprintf("/lines/%d", i)
		if strings.TrimSpace(l.Description) == "" {
			add("REQ_LINE_DESCRIPTION", p+"/description", "line description is required",
				"describe the goods or service on this line")
		}
		if l.Quantity <= 0 {
			add("LINE_QUANTITY", p+"/quantity", fmt.Sprintf("quantity is %d but must be positive", l.Quantity),
				"set \"quantity\" to a whole number of at least 1")
		}
		if l.UnitPrice < 0 {
			add("LINE_UNIT_PRICE", p+"/unit_price", fmt.Sprintf("unit price is %d but must not be negative", l.UnitPrice),
				"use a non-negative unit price in minor units")
		}
		if l.TaxRateBP < 0 {
			add("LINE_TAX_RATE", p+"/tax_rate_bp", fmt.Sprintf("tax rate is %d bp but must not be negative", l.TaxRateBP),
				"use a non-negative rate in basis points (2000 = 20%)")
		}
		ln := l.Quantity * l.UnitPrice
		net += ln
		tax += (ln*l.TaxRateBP + 5000) / 10000 // round half up
	}
	if inv.Net != net {
		add("TOTAL_NET", "/net", fmt.Sprintf("net is %d but lines sum to %d", inv.Net, net),
			fmt.Sprintf("set \"net\" to %d", net))
	}
	if inv.Tax != tax {
		add("TOTAL_TAX", "/tax", fmt.Sprintf("tax is %d but lines compute to %d", inv.Tax, tax),
			fmt.Sprintf("set \"tax\" to %d", tax))
	}
	if inv.Gross != inv.Net+inv.Tax {
		add("TOTAL_GROSS", "/gross", fmt.Sprintf("gross is %d but net+tax is %d", inv.Gross, inv.Net+inv.Tax),
			fmt.Sprintf("set \"gross\" to %d", inv.Net+inv.Tax))
	}
	return out
}
