// Package invoice defines the synthetic invoice model and its validation.
package invoice

// Line is a single invoice line. Money amounts are integer minor units (cents).
type Line struct {
	Description string `json:"description"`
	Quantity    int64  `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
	TaxRateBP   int64  `json:"tax_rate_bp"` // tax rate in basis points (2000 = 20%)
}

// Invoice is a synthetic e-invoice document.
type Invoice struct {
	ID       string `json:"id"`
	Currency string `json:"currency"`
	Lines    []Line `json:"lines"`
	Net      int64  `json:"net"`
	Tax      int64  `json:"tax"`
	Gross    int64  `json:"gross"`
}

// Diagnostic is one actionable finding.
type Diagnostic struct {
	Code    string `json:"code"`
	Path    string `json:"path"`
	Message string `json:"message"`
	Fix     string `json:"fix,omitempty"` // suggested correction
}
