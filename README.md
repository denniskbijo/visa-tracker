# visa-tracker

Open-source UK work visa tracker for tech and STEM professionals. Track salary thresholds, search 140k+ licensed sponsors, look up SOC codes, all from official gov.uk data.

Built with Go, SQLite (pure Go), and HTMX. No JavaScript frameworks. No CGO. Single binary.

## Features

- **Visa Route Dashboard**: compare 6 UK work visa routes side-by-side (Skilled Worker, Global Talent, HPI, Scale-up, ICT, Graduate) with salary thresholds, processing times, and eligibility at a glance
- **Eligibility Checker**: on each visa detail page, enter salary and SOC code to see green/amber/red vs the route threshold (`/visas/{slug}`)
- **Visa Wizard**: answer a few questions and get ranked route suggestions with plain-English reasons (`/wizard`)
- **Personal Timeline**: track visa expiry and ILR countdown in your browser, no account needed (`/timeline`)
- **Sponsor Search**: full-text search across 140k+ employers from the [gov.uk Register of Licensed Sponsors](https://www.gov.uk/government/publications/register-of-licensed-sponsors-workers), filterable by city and route
- **SOC Code Lookup**: find your Standard Occupational Classification code, its going-rate salary, and whether it's on the Immigration Salary List
- **JSON API**: programmatic access at `/api/v1/visas`, `/api/v1/sponsors?q=`, `/api/v1/soc?q=`
- **Auto-refresh**: sponsor data re-downloaded from gov.uk every 24 hours (configurable)
- **Nationality-agnostic**: designed for any nationality, any tech/STEM role

## Quick Start

**Prerequisites:** Go 1.22+ (no C compiler needed)

```bash
git clone https://github.com/denniskbijo/visa-tracker.git
cd visa-tracker
make run
```

Open http://localhost:8080. On first launch the server will:

1. Create `visa-tracker.db` and run schema migrations
2. Load 6 visa routes, salary thresholds, and curated tech/STEM SOC codes from YAML seed data
3. Download the latest licensed sponsors CSV from gov.uk (~140k records)

For a **step-by-step local run checklist** (ports, curl checks, reset DB), see [LOCAL.md](LOCAL.md).

### Docker

```bash
docker build -t visa-tracker .
docker run -p 8080:8080 visa-tracker
```

## Configuration

All settings via environment variables:

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `DB_PATH` | `visa-tracker.db` | SQLite database path |
| `DATA_DIR` | `data` | YAML seed data directory |
| `SPONSOR_CSV_URL` | *(gov.uk latest)* | Licensed sponsors CSV; host must be `assets.publishing.service.gov.uk` |
| `REFRESH_INTERVAL_HOURS` | `24` | Hours between sponsor refreshes |

## Project Layout

```
cmd/
  server/               Application entry point, flag parsing, server lifecycle

internal/
  config/               Environment-variable-based configuration with defaults
  models/               Domain types: visa routes, sponsors, SOC codes, thresholds
  store/                SQLite persistence layer (pure Go via modernc.org/sqlite)
  ingest/               YAML seed loader, gov.uk CSV downloader, refresh scheduler
  handlers/             HTTP handlers for HTML pages and /api/v1/ JSON endpoints
  wizard/               Visa wizard recommendation logic
  eligibility/          Salary vs threshold checker (green/amber/red)
  templates/            Go html/template files with HTMX partials
    partials/           HTMX partial templates for live search results

data/
  thresholds.yaml       Visa routes, salary thresholds, and processing times
  soc_codes.yaml        Curated tech/STEM SOC codes with going rates
  migrations/           SQL schema migrations (run automatically on startup)

static/
  style.css             Design system (dark theme, monospace accents)
  htmx.min.js           HTMX library (only JS dependency)
```

## Data Sources

| Source | What | Updated |
|---|---|---|
| [Register of Licensed Sponsors](https://www.gov.uk/government/publications/register-of-licensed-sponsors-workers) | 140k+ employer names, cities, routes, ratings | Weekly by Home Office |
| [SOC 2020](https://www.ons.gov.uk/methodology/classificationsandstandards/standardoccupationalclassificationsoc/soc2020) | Standard Occupational Classification codes | Stable |
| [Immigration Rules: Appendix Skilled Worker](https://www.gov.uk/guidance/immigration-rules/immigration-rules-appendix-skilled-worker) | Salary thresholds and going rates | On policy change |
| [Appendix Skilled Occupations](https://www.gov.uk/guidance/immigration-rules/immigration-rules-appendix-skilled-occupations) | SOC 2020 codes and going rates | 22 July 2025 snapshot in `data/` |

## API Examples

```bash
# all visa routes
curl http://localhost:8080/api/v1/visas

# search sponsors by name
curl "http://localhost:8080/api/v1/sponsors?q=google"

# filter sponsors by route and city
curl "http://localhost:8080/api/v1/sponsors?q=&route=Skilled+Worker&city=London"

# search SOC codes
curl "http://localhost:8080/api/v1/soc?q=security"
```

## Roadmap

### Recently shipped

| Feature | Description |
|---|---|
| Eligibility checker | Enter salary + SOC on a visa page; green/amber/red vs threshold, with ISL hints |
| Visa wizard | Ranked route suggestions from a short questionnaire (`/wizard`) |
| Personal timeline | Visa expiry and ILR countdown in localStorage (`/timeline`) |

### Next up (high value)

| Feature | Description |
|---|---|
| Route comparison table | Select 2–3 routes and compare sponsor, salary, duration, and ILR path side by side |
| Expand SOC codes | Grow beyond the initial 17 tech/STEM codes in `data/soc_codes.yaml` |
| Salary calculator | Job title → SOC match → minimum salary vs your offer |

### Data and automation

| Feature | Description |
|---|---|
| Processing times scraper | Auto-fetch gov.uk processing times instead of manual YAML snapshots |
| Policy change alerts | Banner when thresholds or rules change after ingest |
| Richer sponsor context | Explain ratings, link to gov.uk, suggest next steps after a search |

### Growth and polish

| Feature | Description |
|---|---|
| Smarter timeline | Multi-visa history, route switches, more accurate ILR rules |
| Country-specific guides | Community-contributed paths (e.g. coming from India, US, EU) |
| Onboarding tour | First-visit walkthrough: job code → salary → sponsors |
| Mobile polish | Improve tables and comparison views on small screens |
| Test coverage | Handlers, store, and ingest tests beyond wizard logic |
| Content-Security-Policy header | Extra XSS defence alongside template auto-escaping |

### Infrastructure

| Item | Description |
|---|---|
| Pin gosec to a release tag | CI currently uses `@master` |
| Upgrade CodeQL action to v4 | Deprecation warning in security workflow |

Work from [LEDGER.md](LEDGER.md): one item at a time. Suggested next: blog post on this data refresh, then eligibility-rule extras, then route comparison.

## Contributing

Contributions welcome. Pick an item from the [Roadmap](#roadmap) above, or open an issue to discuss something new. See [AGENTS.md](AGENTS.md) for repo conventions.

## License

[MIT](LICENSE)
