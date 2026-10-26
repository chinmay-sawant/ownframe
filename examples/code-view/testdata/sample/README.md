# Beacon

Beacon checks the HTTP endpoints you care about every 30 seconds and keeps the
last day of results in SQLite. The dashboard is one HTML page; the API is Go.
The release job renders the last round into the page, so the board works
without a script tag.

> **Status:** internal preview. The scheduler is single-node and the history
> window is fixed at 24 hours.

## Why

Most uptime tools answer one question: is the URL up? Beacon records the status
code, the total latency, and the time to first byte, so a slow endpoint looks
different from a broken one.

## Layout

| Path | What it holds |
|---|---|
| `backend/` | Go API, scheduler, and probe loop |
| `web/` | Dashboard markup, styles, and the fetch client |
| `tools/` | Python reports and a backfill script |
| `docs/` | Architecture and deployment notes |
| `config/` | Service catalog and package metadata |

## Run it

```sh
cd backend
go run . -db /tmp/beacon.db -config ../config/services.yaml
```

Then open <http://127.0.0.1:8099>. The board is the last rendered round; use
`/api/checks` for the live JSON.

## Checks

Each check runs on its own goroutine with a per-request deadline. The default
timeout is 4 seconds, and a check that fails three times in a row is marked down
until it answers twice in a row.

- `up` - HTTP 2xx or 3xx, within the timeout
- `slow` - answered, but slower than the latency budget
- `down` - timeout, connection error, or a 4xx/5xx code

## Notes

- No accounts. The API binds loopback and trusts the caller.
- The report tools read the same SQLite file, so stop the server before a
  backfill or point them at a copy.
- Latency is stored in microseconds; the UI formats milliseconds to one decimal.
