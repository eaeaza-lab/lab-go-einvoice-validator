# Go E-Invoice Validator

A local validator for synthetic e-invoice documents. It checks schemas, totals, tax
calculations, identifiers, required fields and cross-document consistency, then returns
actionable diagnostics. Runs fully offline; all sample data is synthetic.

**Status: work in progress.** See [SPEC.md](SPEC.md) and [PLANS.md](PLANS.md).

## Run

```
go test ./...
go run ./cmd/einvoice
```

Requires Go 1.22+.

## Library

`invoice.Load(path)` reads a JSON invoice (unknown fields are rejected) and
`invoice.Validate` returns diagnostics for required fields, per-line checks and totals.
Each diagnostic has a stable `code`, a JSON-pointer-like `path` (e.g. `/lines/0/quantity`),
a `message` and a suggested `fix`.

Built by a supervised autonomous agent pipeline (nightshift).
