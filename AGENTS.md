# AGENTS

## Commands
- Build: `go build ./...`
- Vet: `go vet ./...`
- Test: `go test ./...`
- Run: `go run ./cmd/einvoice`

## Rules
- Read SPEC.md and PLANS.md first; do one milestone per session, then tick it and add a Progress log line.
- Record design choices in the Decision log in PLANS.md.
- Synthetic data only: no real company, person, marketplace or account names; no secrets.
- No network at runtime. Dependencies: stdlib, Cobra, and a pure-Go SQLite driver only.
- Money is integer minor units; never use floats for amounts.
- Keep `.nightshift.json` commands limited to the runner allowlist (`go vet`, `go test`); no shell wrappers, `&&`, pipes or redirection.
- Every change keeps `go vet ./...` and `go test ./...` green. Add tests with each feature.
- Do not commit unless the session instructions say so.
