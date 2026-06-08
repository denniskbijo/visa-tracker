# AGENTS.md

Guidance for AI agents working on **visa-tracker**.

## Workspace layout

- Parent workspace: `opensource/` (may hold multiple projects).
- **This repo** lives at `opensource/visa-tracker/`. Always edit files here.
- Do **not** move the git root to `opensource/` or flatten the repo into the workspace.

## Project summary

Open-source UK work visa tracker for tech/STEM professionals. Compare visa routes, search licensed sponsors, look up SOC codes, run a visa wizard, and track a personal timeline. Data from gov.uk; deployed on Oracle Cloud Free Tier with Caddy + GitHub Actions.

## Stack

- **Go 1.26.3+** (pure Go, no CGO)
- **SQLite** via `modernc.org/sqlite`
- **HTML templates + HTMX** for UI (minimal JS in `static/app.js` only where needed: wizard steps, timeline localStorage)
- Single binary: `cmd/server`

## Local development

```bash
cd visa-tracker
make run
```

Open http://localhost:8080. First run creates `visa-tracker.db` and ingests sponsor CSV.

Go binary path: `~/.local/go/bin` (add to PATH if `go` not found).

## Coding conventions

- Match existing style: small focused diffs, no over-engineering.
- Reuse handlers, store, and template patterns; do not introduce JS frameworks.
- User-facing copy: **plain English**, glossary tooltips for jargon (SOC, ISL, going rate).
- **No em dashes** (`—` or markdown ` -- ` as punctuation). Use commas, colons, periods, or semicolons instead.
- Security: SSRF protection on sponsor CSV URL, rate-limited API, security headers middleware.
- Personal data stays in **localStorage** (timeline); no auth unless explicitly requested.

## Key paths

| Path | Purpose |
|------|---------|
| `cmd/server/` | Entry point |
| `internal/handlers/` | HTTP handlers (HTML + `/api/v1/*`) |
| `internal/wizard/` | Visa wizard recommendation logic |
| `internal/eligibility/` | Salary vs threshold checker on visa detail pages |
| `internal/templates/` | Go html/template pages and HTMX partials |
| `internal/store/` | SQLite persistence |
| `internal/ingest/` | YAML seed + gov.uk CSV ingest |
| `data/` | YAML seed data and SQL migrations |
| `static/` | CSS, HTMX, `app.js` |

## Deploy

- Manual: `./deploy.sh <VM_IP>` from repo root (targets `/opt/visa-tracker` on Oracle VM).
- CI: push to `main` triggers `.github/workflows/deploy.yml`.
- HTTPS/Caddy: one-time setup via `setup-vm.sh` on the VM. See `DEPLOY.md`.

## Git

- **Commit** only when the user explicitly asks.
- **Push** only when the user explicitly asks. Never push proactively after a commit.
- Follow existing commit message style: short summary, optional body explaining why.
- Never mention Cursor (or any AI tool) in commit messages. No `Co-authored-by` trailers for agents.

## UX preferences (established in this repo)

- Mobile nav hamburger; plain-English nav labels.
- Page intros on major routes.
- Live search (HTMX debounced) on sponsor/SOC pages.
- Helpful empty states with example chips.
- Skilled Worker threshold from DB, not hardcoded in templates.
- Visa wizard at `/wizard`; personal timeline at `/timeline`; eligibility checker on `/visas/{slug}`.

## Tests

```bash
go test ./...
govulncheck ./...
gosec ./...
```

Keep `govulncheck` clean; bump Go/stdlib deps when security CI flags issues.

See [README.md](README.md#roadmap) for planned features and build order.
