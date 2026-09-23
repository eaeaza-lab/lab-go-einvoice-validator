# SPEC — Go E-Invoice Validator

## Problem
Teams building or testing e-invoicing pipelines need to know *why* an invoice document is wrong
before it is submitted anywhere. Errors are usually scattered: malformed structure, missing
fields, arithmetic that does not add up, bad identifiers, or a credit note that disagrees with
the invoice it references. Existing checks tend to be opaque or online-only.

## Target user
A developer or QA engineer who works with e-invoice documents (JSON and XML) and wants a fast,
offline, scriptable CLI that returns precise, actionable diagnostics. All data in this project
is synthetic.

## MVP scope
- CLI `einvoice` (Cobra) with `validate <file...>` and `version`.
- Input formats: JSON and a simple XML representation of the same synthetic model.
- Checks:
  1. Schema (embedded JSON Schema subset) — types, required keys.
  2. Required fields.
  3. Totals: line net, tax per line (integer minor units, basis-point rates, round half up),
     invoice net/tax/gross.
  4. Identifiers: invoice id format, currency code (ISO 4217 subset), tax id checksum (synthetic scheme).
  5. Cross-document consistency: a credit note's referenced invoice must exist in the same
     batch and share currency/party; totals must not exceed the original.
- Diagnostics: stable code, JSON-pointer-like path, message, suggested fix. Output as text or JSON (`--format`).
- Exit code 0 = all valid, 1 = diagnostics found, 2 = usage/IO error.
- SQLite store of validation runs (`einvoice history`), local file only.
- Embedded synthetic fixtures (`embed`) used by tests and `einvoice demo`.

## Non-goals
- No real-world legal/tax compliance claims or country-specific rule packs.
- No network access at runtime, no submission to any tax authority or network.
- No real company, person or account data; no PDF/image parsing; no signing/encryption.
- No web server or GUI in the MVP (the "saas" goal is a later, separate concern).

## Acceptance criteria (each checkable by command)
1. `go build ./...` exits 0.
2. `go vet ./...` exits 0.
3. `go test ./...` exits 0.
4. `go run ./cmd/einvoice version` prints a version line and exits 0.
5. `go run ./cmd/einvoice validate testdata/valid.json` exits 0.
6. `go run ./cmd/einvoice validate testdata/bad_totals.json` exits 1 and output contains `TOTAL_GROSS`.
7. `go run ./cmd/einvoice validate --format json testdata/bad_totals.json` outputs valid JSON with a `diagnostics` array.
8. `go run ./cmd/einvoice validate testdata/valid.xml` exits 0 (XML parity).
9. `go run ./cmd/einvoice validate testdata/missing_id.json` exits 1 and output contains `REQ_ID`.
10. `go test ./internal/crossdoc/...` passes, covering credit-note/invoice mismatch.
11. `go test ./internal/store/...` passes, covering persisting and listing runs in SQLite.
12. `go test ./... -run Fixtures` validates every embedded fixture against its expected diagnostics.
13. No non-stdlib dependency other than Cobra and a pure-Go SQLite driver appears in `go.mod`.
