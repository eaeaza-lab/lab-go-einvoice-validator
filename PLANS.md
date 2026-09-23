# PLANS — execution plan

Each milestone fits one ~30-minute session. Acceptance command must exit 0.

## Milestones

- [x] **M0 setup** (mvp) — spec, plans, skeleton, one passing test. Accept: `go test ./...`
- [x] **M1 model + JSON loader** (mvp) — `invoice.Load(path)` for JSON, unknown-field errors. Accept: `go test ./internal/invoice/...`
- [x] **M2 required-field + totals rules polish** (mvp) — per-line diagnostics with paths and suggested fixes. Accept: `go test ./internal/invoice/...`
- [x] **M3 Cobra CLI: validate + version** (mvp) — add Cobra, exit codes 0/1/2, text output. Accept: `go test ./cmd/...`
- [x] **M4 JSON output format** (mvp) — `--format json`. Accept: `go test ./cmd/...`
- [x] **M5 embedded fixtures** (mvp) — `testdata` + `embed`, fixture-driven test `-run Fixtures`. Accept: `go test ./... -run Fixtures`
- [x] **M6 JSON Schema subset validator** (mvp) — embedded schema, stdlib-only checker for type/required/enum. Accept: `go test ./internal/schema/...`
- [ ] **M7 XML input** (mvp) — `encoding/xml` loader, parity with JSON fixtures. Accept: `go test ./internal/invoice/...`
- [ ] **M8 identifier checks** (mvp) — invoice id format, currency, synthetic tax-id checksum. Accept: `go test ./internal/ident/...`
- [ ] **M9 cross-document consistency** (mvp) — credit note vs. invoice batch. Accept: `go test ./internal/crossdoc/...`
- [ ] **M10 SQLite run history** (mvp) — pure-Go driver, `history` command. Accept: `go test ./internal/store/...`
- [ ] **M11 demo command + docs** (polish) — `einvoice demo` from embedded fixtures; README usage. Accept: `go test ./...`
- [ ] **M12 diagnostics polish** (polish) — stable code catalogue, `einvoice explain <code>`. Accept: `go test ./...`
- [ ] **M13 golden-output tests + lint pass** (polish) — golden files, `go vet` clean. Accept: `go vet ./...`

## Progress log

- M0: spec, plans, README, AGENTS, skeleton with totals validation and 3 passing tests written. Not committed.
- M1 (2026-09-24): `invoice.Load(path)` / `ParseJSON` with unknown-field and trailing-content errors, 5 tests. Not committed.
- M2 (2026-09-24): per-line diagnostics (description, quantity, unit price, tax rate), REQ_CURRENCY, JSON-pointer paths and a `fix` field on every diagnostic; 4 new tests. Not committed.
- M3 (2026-09-24): Cobra CLI with `validate <file>...` and `version`, exit codes 0/1/2, text output, 5 tests in `cmd/einvoice`; added `testdata/valid.json` and `bad_totals.json`; cobra declared in go.mod (go.sum left to the runner's `go mod tidy`). Not committed.
- M4 (2026-09-24): `validate --format text|json`; JSON is one report object (`valid`, `files`, `diagnostics` with per-entry `file`); 3 new tests in `cmd/einvoice`. Not committed.
- M5 (2026-09-24): new `internal/fixtures` package embeds `data/*.json` (valid, bad_totals, missing_id, bad_line) plus `expected.json` mapping each fixture to its expected diagnostic codes; `TestFixtures` validates all of them. Added `testdata/missing_id.json` for the CLI. Not committed.
- M6 (2026-09-24): new `internal/schema` package: subset checker (type/required/properties/items/enum) with embedded `invoice.schema.json`, codes `SCHEMA_TYPE`/`SCHEMA_REQUIRED`/`SCHEMA_ENUM`, 7 tests. Not wired into the CLI yet. Not committed.

## Decision log

- D1: Money is integer minor units; tax rates in basis points; rounding is half up per line. Avoids floats.
- D2: Skeleton is stdlib-only so `go test` works offline; Cobra and a pure-Go SQLite driver (modernc.org/sqlite) are added in M3/M10 (the check sandbox has network access).
- D3: JSON Schema support is a small in-repo subset checker rather than a large dependency.
- D5: JSON loader uses `DisallowUnknownFields` and rejects any content after the first document, so typos in field names surface as errors rather than silently validating.
- D6: Diagnostic paths use a leading slash (`/lines/0/quantity`); `Fix` is `omitempty` in JSON. Quantity must be >= 1, unit price and tax rate >= 0 (zero-priced lines allowed). Totals fixes quote the computed value.
- D7: `run(args, stdout, stderr) int` holds the CLI logic so tests call it directly; Cobra's own usage/error printing is silenced and errors are mapped to exit codes centrally. Load/parse failures (missing or malformed file) are exit 2, not 1; a run over several files keeps going through diagnostics but aborts on the first I/O error.
- D8: JSON output is a single report object for the whole run (not one document per file) so it stays parseable with several inputs; `diagnostics` is always an array (never null) and each entry carries its `file`. Unknown `--format` values are a usage error (exit 2). I/O errors still abort with exit 2 and print nothing on stdout in JSON mode.
- D9: `go:embed` cannot reach `../testdata` and `testdata` dirs are skipped by `./...`, so embedded fixtures live in `internal/fixtures/data/` (separate from the CLI's on-disk `testdata/`). Expectations are an ordered list of diagnostic codes per fixture in `expected.json`; a fixture without an entry (or vice versa) is an error.
- D10: Schema checker is standalone and not yet part of `validate`: wiring it in would add SCHEMA_* codes on top of the REQ_* ones and change fixture expectations, so that is left for a later milestone. Schemas are parsed strictly (unknown keywords and unsupported types are errors); numbers are decoded as `json.Number` so "integer" means an int64-parsable literal (`1.5` fails, `2` passes); enum compares canonical JSON renderings; a type mismatch stops descent into that value.
- D4: `.nightshift.json` runs only `go vet` and `go test`, both on the allowlist.
