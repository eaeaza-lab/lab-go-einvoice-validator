// Package invoice defines the synthetic invoice model and its validation.
package invoice

import "encoding/xml"

// Line is a single invoice line. Money amounts are integer minor units (cents).
type Line struct {
	Description string `json:"description" xml:"description"`
	Quantity    int64  `json:"quantity" xml:"quantity"`
	UnitPrice   int64  `json:"unit_price" xml:"unit_price"`
	TaxRateBP   int64  `json:"tax_rate_bp" xml:"tax_rate_bp"` // tax rate in basis points (2000 = 20%)
}

// Invoice is a synthetic e-invoice document.
type Invoice struct {
	XMLName  xml.Name `json:"-" xml:"invoice"` // only used while decoding XML; cleared afterwards
	ID       string   `json:"id" xml:"id"`
	Currency string   `json:"currency" xml:"currency"`
	Lines    []Line   `json:"lines" xml:"lines>line"`
	Net      int64    `json:"net" xml:"net"`
	Tax      int64    `json:"tax" xml:"tax"`
	Gross    int64    `json:"gross" xml:"gross"`
}

// Diagnostic is one actionable finding.
type Diagnostic struct {
	Code    string `json:"code"`
	Path    string `json:"path"`
	Message string `json:"message"`
	Fix     string `json:"fix,omitempty"` // suggested correction
}
