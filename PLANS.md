# PLANS — execution plan

Each milestone fits one ~30-minute session. Acceptance command must exit 0.

## Milestones

- [x] **M0 setup** (mvp) — spec, plans, skeleton, one passing test. Accept: `go test ./...`
- [x] **M1 model + JSON loader** (mvp) — `invoice.Load(path)` for JSON, unknown-field errors. Accept: `go test ./internal/invoice/...`
- [x] **M2 required-field + totals rules polish** (mvp) — per-line diagnostics with paths and suggested fixes. Accept: `go test ./internal/invoice/...`
- [x] **M3 Cobra CLI: validate + version** (mvp) — add Cobra, exit codes 0/1/2, text output. Accept: `go test ./cmd/...`
- [ ] **M4 JSON output format** (mvp) — `--format json`. Accept: `go test ./cmd/...`
- [ ] **M5 embedded fixtures** (mvp) — `testdata` + `embed`, fixture-driven test `-run Fixtures`. Accept: `go test ./... -run Fixtures`
- [ ] **M6 JSON Schema subset validator** (mvp) — embedded schema, stdlib-only checker for type/required/enum. Accept: `go test ./internal/schema/...`
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

## Decision log

- D1: Money is integer minor units; tax rates in basis points; rounding is half up per line. Avoids floats.
- D2: Skeleton is stdlib-only so `go test` works offline; Cobra and a pure-Go SQLite driver (modernc.org/sqlite) are added in M3/M10 (the check sandbox has network access).
- D3: JSON Schema support is a small in-repo subset checker rather than a large dependency.
- D5: JSON loader uses `DisallowUnknownFields` and rejects any content after the first document, so typos in field names surface as errors rather than silently validating.
- D6: Diagnostic paths use a leading slash (`/lines/0/quantity`); `Fix` is `omitempty` in JSON. Quantity must be >= 1, unit price and tax rate >= 0 (zero-priced lines allowed). Totals fixes quote the computed value.
- D7: `run(args, stdout, stderr) int` holds the CLI logic so tests call it directly; Cobra's own usage/error printing is silenced and errors are mapped to exit codes centrally. Load/parse failures (missing or malformed file) are exit 2, not 1; a run over several files keeps going through diagnostics but aborts on the first I/O error.
- D4: `.nightshift.json` runs only `go vet` and `go test`, both on the allowlist.
