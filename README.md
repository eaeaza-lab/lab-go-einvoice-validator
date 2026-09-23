# Go E-Invoice Validator

A local validator for synthetic e-invoice documents. It checks schemas, totals, tax
calculations, identifiers, required fields and cross-document consistency, then returns
actionable diagnostics. Runs fully offline; all sample data is synthetic.

**Status: work in progress.** See [SPEC.md](SPEC.md) and [PLANS.md](PLANS.md).

## Run

```
go test ./...
go run ./cmd/einvoice version
go run ./cmd/einvoice validate testdata/valid.json
go run ./cmd/einvoice validate testdata/bad_totals.json
go run ./cmd/einvoice validate testdata/valid.xml
```

`validate <file>...` checks JSON invoices (or XML, for files ending in `.xml`) and prints one line per diagnostic
(`file: CODE /path: message`, followed by a `fix:` hint). Exit codes: `0` all valid,
`1` diagnostics found, `2` usage or I/O error (missing/malformed file, bad arguments).

`validate --format json <file>...` prints one JSON document instead:
`{"valid": bool, "files": [...], "diagnostics": [{"file", "code", "path", "message", "fix"}]}`
(`diagnostics` is `[]` when everything is valid). Exit codes are the same. `--format`
accepts `text` (default) or `json`; anything else is a usage error (exit `2`).

Requires Go 1.22+.

## Library

`invoice.Load(path)` reads a JSON invoice, or an XML one for `.xml` paths (unknown fields,
elements and attributes are rejected; see `testdata/valid.xml` for the element layout), and
`invoice.Validate` returns diagnostics for required fields, per-line checks and totals.
Each diagnostic has a stable `code`, a JSON-pointer-like `path` (e.g. `/lines/0/quantity`),
a `message` and a suggested `fix`.

## Schema subset

`internal/schema` checks a JSON document against an embedded schema
(`invoice.schema.json`) supporting `type`, `required`, `properties`, `items` and `enum`.
Violations are reported as `SCHEMA_TYPE`, `SCHEMA_REQUIRED` or `SCHEMA_ENUM` with a
JSON-pointer-like path. It is a library only for now; the CLI does not call it yet.

## Identifier checks

`internal/ident` validates the invoice id (`INV-YYYY-NNNN`, code `ID_FORMAT`), the currency
(ISO 4217 subset, `CURRENCY_UNKNOWN`) and a synthetic tax id `SY` + 8 digits + check digit
(`TAXID_FORMAT`, `TAXID_CHECKSUM`). Library only for now; the CLI does not call it yet.

## Cross-document checks

`internal/crossdoc` checks a batch: each credit note must reference an invoice in the batch
(`XDOC_REF_MISSING`, `XDOC_REF_NOT_FOUND`), share its currency and party (`XDOC_CURRENCY`,
`XDOC_PARTY`), and not credit more than the invoice net, tax or gross in total
(`XDOC_EXCEEDS_NET`, `XDOC_EXCEEDS_TAX`, `XDOC_EXCEEDS_GROSS`). Library only for now; the CLI
does not call it yet.

## Fixtures

`internal/fixtures` embeds synthetic sample invoices (`internal/fixtures/data/*.json`) and
`expected.json`, which lists the diagnostic codes each fixture must produce.
`go test ./... -run Fixtures` validates every fixture against its expectation; add a new
fixture by dropping a JSON file in `data/` and an entry in `expected.json`.

Built by a supervised autonomous agent pipeline (nightshift).
