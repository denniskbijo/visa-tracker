# visa-tracker

Open-source UK work visa tracker for tech and STEM professionals. Track salary thresholds, search 140k+ licensed sponsors, look up SOC codes -- all from official gov.uk data.

Built with Go, SQLite (pure Go), and HTMX. No JavaScript frameworks. No CGO. Single binary.

## Features

- **Visa Route Dashboard** -- compare 6 UK work visa routes side-by-side (Skilled Worker, Global Talent, HPI, Scale-up, ICT, Graduate) with salary thresholds, processing times, and eligibility at a glance
- **Sponsor Search** -- full-text search across 140k+ employers from the [gov.uk Register of Licensed Sponsors](https://www.gov.uk/government/publications/register-of-licensed-sponsors-workers), filterable by city and route
- **SOC Code Lookup** -- find your Standard Occupational Classification code, its going-rate salary, and whether it's on the Immigration Salary List
- **JSON API** -- programmatic access at `/api/v1/visas`, `/api/v1/sponsors?q=`, `/api/v1/soc?q=`
- **Auto-refresh** -- sponsor data re-downloaded from gov.uk every 24 hours (configurable)
- **Nationality-agnostic** -- designed for any nationality, any tech/STEM role

## Quick Start

**Prerequisites:** Go 1.22+ (no C compiler needed)

```bash
git clone https://github.com/denniskbijo/visa-tracker.git
cd visa-tracker
make run
```

Open http://localhost:8080. On first launch the server will:

1. Create `visa-tracker.db` and run schema migrations
2. Load 6 visa routes, salary thresholds, and 17 SOC codes from YAML seed data
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
| `SPONSOR_CSV_URL` | *(gov.uk latest)* | Licensed sponsors CSV URL |
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

## Contributing

Contributions welcome. Areas that need help:

- Expanding SOC codes beyond the initial tech/STEM set
- Automated gov.uk processing times scraper
- Community-contributed country-specific guides
- ILR countdown / personal visa timeline tracker
- Mobile-responsive design improvements
- Test coverage

## License

[MIT](LICENSE)
