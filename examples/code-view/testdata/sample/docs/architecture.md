# Architecture

Beacon has three moving parts: a scheduler, a prober, and a store. The
dashboard is static HTML; the release job renders the latest round into it.

```text
            ┌──────────────┐        ┌──────────────┐
  tick ───▶ │  Scheduler   │ ─────▶ │    Prober    │ ───▶ target
            └──────┬───────┘        └──────────────┘
                   │ Result
                   ▼
            ┌──────────────┐        ┌──────────────┐
            │    Store     │ ◀───── │   Dashboard  │
            └──────────────┘  JSON  └──────────────┘
```

## Rules

1. **One writer.** Only the scheduler inserts rows. The API reads.
2. **One deadline per attempt.** A probe never outlives its context.
3. **Bounded bodies.** Beacon reads 64 KB and closes the connection; it never
   buffers a whole response.
4. **No retries on 4xx.** A 404 is an answer, not a transient failure.
5. **The UI is replaceable.** The dashboard uses the public JSON routes and no
   server-rendered state.

## Scheduling

`Scheduler.round` starts one goroutine per service and limits concurrency with
a four-slot channel. A slow endpoint cannot hold up the other seven.

```go
for _, svc := range s.services {
    wg.Add(1)
    go func(svc Service) {
        defer wg.Done()
        sem <- struct{}{}
        defer func() { <-sem }()
        s.store.Record(ctx, s.prober.Probe(ctx, svc))
    }(svc)
}
```

The round waits for every goroutine, so the next tick cannot overlap a round
that is still draining. Missing a tick is better than two probes of the same
endpoint racing each other.

## Probing

`Prober.once` attaches an `httptrace.ClientTrace` and records the first
response byte. That number is the one operators ask about when a page is
"slow": a large gap before the first byte points at DNS, TLS, or a queue; a
small gap with a large total points at the body or the client.

The state table:

| Condition | State | Stored error |
|---|---|---|
| 2xx or 3xx within the budget | `up` | empty |
| 2xx or 3xx over the budget | `slow` | empty |
| 4xx | `down` | status code |
| 5xx after three attempts | `down` | status code |
| Timeout or dial error | `down` | transport error |

## Storage

The schema is one table and two indexes. `at` is a Unix second; latency is
stored in microseconds so the report tools do not lose precision to a float.

```sql
CREATE TABLE probes (
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    service  TEXT NOT NULL,
    at       INTEGER NOT NULL,
    state    TEXT NOT NULL,
    status   INTEGER NOT NULL,
    total_us INTEGER NOT NULL,
    ttfb_us  INTEGER NOT NULL,
    err      TEXT NOT NULL DEFAULT '',
    attempts INTEGER NOT NULL DEFAULT 1
);
```

`Latest` joins each service to its maximum timestamp. That query is the one
the dashboard waits on, so the index order is `(service, at DESC)`.

## Dashboard

The dashboard is a single HTML file, two stylesheets, and no build step. The
release job bakes the last round into the table; the JSON routes stay the
source of truth for the admin build.

The detail drawer is a fixed aside that slides in over the board. It keeps the
list readable at 1280 px and below.

## Limits

- Single node. Two schedulers against one database write duplicate rows.
- The retention window is fixed at 24 hours; `Prune` exists but nothing calls
  it on a timer yet.
- Latency is measured from the client, so a slow local network looks like a
  slow service.
- TCP checks in the catalog (`tcp://`) are listed but the prober only speaks
  HTTP today. See `docs/deployment.md`.
