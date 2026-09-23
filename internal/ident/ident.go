// Package ident checks identifiers: invoice id format, currency code and a
// synthetic tax-id checksum. It is standalone and returns invoice.Diagnostic values.
package ident

import (
	"fmt"
	"regexp"
	"strings"

	"example.com/einvoice-validator/internal/invoice"
)

var invoiceIDRe = regexp.MustCompile(`^INV-[0-9]{4}-[0-9]{4,8}$`)

// currencies is a small ISO 4217 subset.
var currencies = map[string]bool{
	"EUR": true, "USD": true, "GBP": true, "CHF": true, "SEK": true, "NOK": true,
	"DKK": true, "PLN": true, "CZK": true, "JPY": true, "CAD": true, "AUD": true,
}

// TaxIDPrefix starts every synthetic tax id: "SY" + 8 digits + 1 check digit.
const TaxIDPrefix = "SY"

// CheckInvoiceID validates the invoice id format (INV-YYYY-NNNN..). An empty id is
// left to the required-field rule and yields no diagnostic here.
func CheckInvoiceID(id string) []invoice.Diagnostic {
	if strings.TrimSpace(id) == "" || invoiceIDRe.MatchString(id) {
		return nil
	}
	return []invoice.Diagnostic{{
		Code:    "ID_FORMAT",
		Path:    "/id",
		Message: fmt.Sprintf("invoice id %q does not match INV-YYYY-NNNN", id),
		Fix:     "use the form INV-<year>-<4 to 8 digit number>, for example \"INV-2026-0001\"",
	}}
}

// CheckCurrency validates the currency against the supported ISO 4217 subset.
// An empty code is left to the required-field rule.
func CheckCurrency(code string) []invoice.Diagnostic {
	if code == "" || currencies[code] {
		return nil
	}
	return []invoice.Diagnostic{{
		Code:    "CURRENCY_UNKNOWN",
		Path:    "/currency",
		Message: fmt.Sprintf("currency %q is not a supported ISO 4217 code", code),
		Fix:     "use an upper-case 3-letter code such as \"EUR\" or \"USD\"",
	}}
}

// TaxIDCheckDigit computes the check digit for the 8 payload digits: the sum of
// digit*(position+1) over positions 0..7, taken mod 10.
func TaxIDCheckDigit(payload string) (byte, bool) {
	if len(payload) != 8 {
		return 0, false
	}
	sum := 0
	for i := 0; i < 8; i++ {
		c := payload[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		sum += int(c-'0') * (i + 1)
	}
	return byte('0' + sum%10), true
}

// CheckTaxID validates the synthetic tax id "SY" + 8 digits + check digit.
// The path is supplied by the caller since tax ids sit on parties, not the invoice root.
func CheckTaxID(path, id string) []invoice.Diagnostic {
	if !strings.HasPrefix(id, TaxIDPrefix) || len(id) != len(TaxIDPrefix)+9 {
		return []invoice.Diagnostic{{
			Code:    "TAXID_FORMAT",
			Path:    path,
			Message: fmt.Sprintf("tax id %q must be %q followed by 9 digits", id, TaxIDPrefix),
			Fix:     "use the form SY########C (8 digits plus a check digit)",
		}}
	}
	body := id[len(TaxIDPrefix):]
	want, ok := TaxIDCheckDigit(body[:8])
	if !ok || body[8] < '0' || body[8] > '9' {
		return []invoice.Diagnostic{{
			Code:    "TAXID_FORMAT",
			Path:    path,
			Message: fmt.Sprintf("tax id %q must be %q followed by 9 digits", id, TaxIDPrefix),
			Fix:     "use the form SY########C (8 digits plus a check digit)",
		}}
	}
	if body[8] != want {
		return []invoice.Diagnostic{{
			Code:    "TAXID_CHECKSUM",
			Path:    path,
			Message: fmt.Sprintf("tax id %q has check digit %c but %c was expected", id, body[8], want),
			Fix:     fmt.Sprintf("replace the last digit with %c", want),
		}}
	}
	return nil
}
