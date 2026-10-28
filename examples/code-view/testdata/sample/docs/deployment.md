# Deployment

Beacon runs as one process and one SQLite file. There is no queue, no worker
pool, and no shared cache.

## Build

```sh
cd backend
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /usr/local/bin/beacon .
```

The dashboard files are served from `../web` relative to the working
directory. Install both:

```sh
install -d /srv/beacon/web /var/lib/beacon
install -m 0644 web/*.css web/*.html web/*.js /srv/beacon/web/
```

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `BEACON_ADDR` | `127.0.0.1:8099` | Listen address |
| `BEACON_DB` | `beacon.db` | SQLite path |
| `BEACON_CONFIG` | `config/services.yaml` | Service catalog |
| `BEACON_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |

Durations in the catalog use Go syntax: `30s`, `750ms`, `4s`. A zero interval
becomes 30 seconds, and a zero timeout becomes 4 seconds.

## systemd

```ini
[Unit]
Description=Beacon endpoint health
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=beacon
Group=beacon
WorkingDirectory=/srv/beacon
Environment=BEACON_DB=/var/lib/beacon/beacon.db
ExecStart=/usr/local/bin/beacon -config /srv/beacon/config/services.yaml
Restart=on-failure
RestartSec=5s

# The SQLite file and its journal live here.
ReadWritePaths=/var/lib/beacon

[Install]
WantedBy=multi-user.target
```

## Reverse proxy

Terminate TLS at the proxy and keep Beacon on loopback. It sends no
authentication of its own.

```nginx
location / {
    proxy_pass http://127.0.0.1:8099;
    proxy_set_header Host $host;
    proxy_read_timeout 10s;
}
```

Add a header check at the proxy if the dashboard leaves the machine:

```nginx
if ($http_x_beacon_token != "replace-me") {
    return 401;
}
```

## Retention

History grows by one row per service per interval. Eight services at 30
seconds is about 23,000 rows a day, under 8 MB with the indexes.

A daily prune keeps the window bounded:

```sh
sqlite3 /var/lib/beacon/beacon.db \
  "DELETE FROM probes WHERE at < strftime('%s','now') - 86400;"
```

Run it from a timer, not from the request path.

## Upgrades

1. Stop the service.
2. Back up the database file and any `-wal` and `-shm` siblings.
3. Replace the binary and the `web/` files.
4. Start the service and check `/api/health`.

The schema is additive, so an older binary against a newer file is the only
direction that breaks.
