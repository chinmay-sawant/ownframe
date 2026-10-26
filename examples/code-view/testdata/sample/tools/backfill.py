#!/usr/bin/env python3
"""Write synthetic probe rows for a dashboard demo.

The rows are plausible, not random noise: each service has a baseline latency,
an outage window, and a slow minute. Point --db at a copy so a demo never
touches production history.
"""

from __future__ import annotations

import argparse
import math
import random
import sqlite3
from datetime import UTC, datetime, timedelta
from pathlib import Path

SERVICES = {
    "checkout-api": {"base": 42, "jitter": 18, "tags": ["api"]},
    "catalog": {"base": 28, "jitter": 9, "tags": ["api"]},
    "search": {"base": 66, "jitter": 24, "tags": ["api"]},
    "dashboard": {"base": 95, "jitter": 30, "tags": ["web"]},
    "docs": {"base": 18, "jitter": 6, "tags": ["web"]},
    "postgres": {"base": 12, "jitter": 4, "tags": ["infra"]},
    "redis": {"base": 3, "jitter": 2, "tags": ["infra"]},
    "mailer": {"base": 51, "jitter": 15, "tags": ["internal"]},
}


def ensure_schema(connection: sqlite3.Connection) -> None:
    """Create the table when the demo database is empty."""
    connection.executescript(
        """
        CREATE TABLE IF NOT EXISTS probes (
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
        CREATE INDEX IF NOT EXISTS probes_service_at ON probes (service, at DESC);
        """
    )


def state_for(minute: int, latency_ms: float, budget_ms: int = 750) -> tuple[str, int, str]:
    """Pick a state from the outage window and the latency."""
    if 180 <= minute <= 192 and minute % 3 == 0:
        return "down", 0, "dial tcp: i/o timeout"
    if latency_ms > budget_ms:
        return "slow", 200, ""
    if minute % 97 == 0:
        return "down", 502, "upstream connect error"
    return "up", 200, ""


def backfill(connection: sqlite3.Connection, hours: int, seed: int) -> int:
    """Insert one row per service per minute."""
    rng = random.Random(seed)
    end = datetime.now(UTC).replace(second=0, microsecond=0)
    start = end - timedelta(hours=hours)
    inserted = 0

    rows = []
    for service, profile in SERVICES.items():
        for minute in range(hours * 60):
            at = start + timedelta(minutes=minute)
            wave = math.sin(minute / 19.0) * 4.0
            latency = profile["base"] + wave + rng.gauss(0, profile["jitter"])
            state, status, error = state_for(minute, latency)
            total_us = max(1, int(latency * 1000))
            rows.append(
                (
                    service,
                    int(at.timestamp()),
                    state,
                    status,
                    total_us,
                    max(1, int(total_us * 0.62)),
                    error,
                    1 if state == "up" else 3,
                )
            )

    connection.executemany(
        """
        INSERT INTO probes (service, at, state, status, total_us, ttfb_us, err, attempts)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        """,
        rows,
    )
    inserted = len(rows)
    connection.commit()
    return inserted


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--db", type=Path, default=Path("beacon.db"))
    parser.add_argument("--hours", type=int, default=6, choices=range(1, 25))
    parser.add_argument("--seed", type=int, default=7)
    args = parser.parse_args()

    connection = sqlite3.connect(args.db)
    try:
        ensure_schema(connection)
        count = backfill(connection, args.hours, args.seed)
    finally:
        connection.close()

    print(f"wrote {count} rows to {args.db} ({args.hours}h, seed {args.seed})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
