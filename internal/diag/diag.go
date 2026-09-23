// Package diag is the catalogue of stable diagnostic codes.
package diag

import "sort"

// Entry describes one diagnostic code.
type Entry struct {
	Code    string
	Summary string
	Fix     string
}

var catalogue = []Entry{
	{"REQ_ID", "The invoice has no id.", "Set a non-empty id such as INV-2026-0001."},
	{"REQ_CURRENCY", "The invoice has no currency.", "Set a currency code such as EUR."},
	{"REQ_LINES", "The invoice has no lines.", "Add at least one line."},
	{"REQ_LINE_DESCRIPTION", "A line has no description.", "Give the line a non-empty description."},
	{"LINE_QUANTITY", "A line quantity is below 1.", "Use a quantity of at least 1."},
	{"LINE_UNIT_PRICE", "A line unit price is negative.", "Use a unit price of 0 or more (minor units)."},
	{"LINE_TAX_RATE", "A line tax rate is negative.", "Use a tax rate of 0 or more (basis points)."},
	{"TOTAL_NET", "The invoice net total differs from the sum of line nets.", "Set net to the computed sum of line nets."},
	{"TOTAL_TAX", "The invoice tax total differs from the sum of per-line taxes.", "Set tax to the computed sum (rounded half up per line)."},
	{"TOTAL_GROSS", "The invoice gross total is not net plus tax.", "Set gross to net + tax."},
	{"SCHEMA_TYPE", "A value has the wrong JSON type.", "Change the value to the type the schema requires."},
	{"SCHEMA_REQUIRED", "A required key is missing.", "Add the missing key."},
	{"SCHEMA_ENUM", "A value is not one of the allowed values.", "Use one of the values listed in the schema."},
	{"ID_FORMAT", "The invoice id does not match INV-YYYY-NNNN.", "Use the form INV-YYYY-NNNN with 4 to 8 digits."},
	{"CURRENCY_UNKNOWN", "The currency code is not in the supported ISO 4217 subset.", "Use a supported three-letter code."},
	{"TAXID_FORMAT", "The tax id is not SY followed by 9 digits.", "Use SY, 8 digits and a check digit."},
	{"TAXID_CHECKSUM", "The tax id check digit is wrong.", "Recompute the check digit: sum(digit*(pos+1)) mod 10."},
	{"XDOC_REF_MISSING", "A credit note does not reference an invoice.", "Set the referenced invoice id."},
	{"XDOC_REF_NOT_FOUND", "The referenced invoice is not in the batch.", "Add the invoice to the batch or fix the reference."},
	{"XDOC_CURRENCY", "The credit note currency differs from the invoice.", "Use the invoice's currency."},
	{"XDOC_PARTY", "The credit note party differs from the invoice.", "Use the invoice's party."},
	{"XDOC_EXCEEDS_NET", "Credited net exceeds the invoice net.", "Reduce the credited net."},
	{"XDOC_EXCEEDS_TAX", "Credited tax exceeds the invoice tax.", "Reduce the credited tax."},
	{"XDOC_EXCEEDS_GROSS", "Credited gross exceeds the invoice gross.", "Reduce the credited gross."},
}

// All returns the catalogue sorted by code.
func All() []Entry {
	out := append([]Entry(nil), catalogue...)
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	return out
}

// Lookup returns the entry for code.
func Lookup(code string) (Entry, bool) {
	for _, e := range catalogue {
		if e.Code == code {
			return e, true
		}
	}
	return Entry{}, false
}
