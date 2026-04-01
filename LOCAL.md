# Running visa-tracker locally

Use this checklist to run the app on your machine and confirm pages, API, and static assets work before deploying.

## Prerequisites

- **Go 1.22+** (`go version`)
- Network access to download the gov.uk sponsor CSV on first run (HTTPS)

## Steps

### 1. Clone and enter the repo

```bash
git clone https://github.com/denniskbijo/visa-tracker.git
cd visa-tracker
```

### 2. Build (optional)

```bash
go build -o visa-tracker ./cmd/server
```

Or use the Makefile:

```bash
make build
```

### 3. Run the server

Default port is **8080**.

```bash
make run
```

Or without Make:

```bash
go run ./cmd/server
```

Or run the binary you built:

```bash
./visa-tracker
```

### 4. Use a different port (optional)

If **8080** is busy:

```bash
PORT=8090 go run ./cmd/server
```

Then open **http://localhost:8090** in your browser.

### 5. What happens on first launch

1. Creates **`visa-tracker.db`** in the current working directory
2. Runs SQL migrations
3. Loads YAML seed data (visa routes, thresholds, SOC codes)
4. Starts a background job to **download the gov.uk sponsor CSV** (~140k rows). This can take **1–3 minutes** on a slow link; the HTTP server is up before this finishes
5. After sponsors load, sponsor search and `/api/v1/sponsors` return full results

Subsequent runs reuse the existing database and skip a full re-download until the scheduled refresh (default: every 24 hours).

### 6. Quick verification (browser)

Open:

| URL | What to expect |
|-----|----------------|
| http://localhost:8080/ | Home page with visa stats |
| http://localhost:8080/visas | Visa dashboard |
| http://localhost:8080/sponsors | Sponsor search (empty until CSV loaded) |
| http://localhost:8080/soc | SOC code table |

### 7. Quick verification (terminal)

```bash
# Home page
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/

# JSON API
curl -s http://localhost:8080/api/v1/visas | head -c 200
echo

# Static CSS
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/static/style.css
```

Expect **200** for each.

### 8. Stop the server

Press **Ctrl+C** in the terminal where it is running.

### 9. Reset local data (optional)

To start from a clean database:

```bash
rm -f visa-tracker.db visa-tracker.db-shm visa-tracker.db-wal
go run ./cmd/server
```

## Environment variables (local)

| Variable | Default | Purpose |
|----------|---------|---------|
| `PORT` | `8080` | HTTP listen port |
| `DB_PATH` | `visa-tracker.db` | SQLite file path |
| `DATA_DIR` | `data` | YAML seed data directory |
| `SPONSOR_CSV_URL` | gov.uk URL | Sponsor CSV (override for testing) |
| `REFRESH_INTERVAL_HOURS` | `24` | Hours between sponsor refreshes |

Example:

```bash
PORT=3000 DB_PATH=/tmp/vt.db go run ./cmd/server
```

## Troubleshooting

- **Port already in use:** set `PORT=` to another value (see above).
- **`index ... already exists` / migration errors on second `make run`:** fixed in current code via `schema_migrations`. If you still see this on an old checkout, delete `visa-tracker.db` (and `-wal`/`-shm` if present) and start again, or `git pull` and rebuild.
- **Sponsor download fails:** check outbound HTTPS; you can still use `/visas` and `/soc` from seed data. Fix the URL or network, then restart.
- **Wrong working directory:** run from the repo root so `data/` and `static/` resolve correctly, or set `DATA_DIR`, `STATIC_DIR`, `TEMPLATES_DIR` to absolute paths.

## Related docs

- [README.md](README.md) -- project overview and API
