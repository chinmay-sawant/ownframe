#!/usr/bin/env python3
"""Summarise Beacon probe history from SQLite.

The report reads the same database the server writes. Use --copy when the
server is running so a long query never holds the write lock.

Examples:
    python3 report.py --db /tmp/beacon.db
    python3 report.py --db /tmp/beacon.db --since 6h --format csv
"""

from __future__ import annotations

import argparse
import csv
import sqlite3
import sys
from dataclasses import dataclass
from datetime import UTC, datetime, timedelta
from pathlib import Path

WINDOWS = {
    "15m": timedelta(minutes=15),
    "1h": timedelta(hours=1),
    "6h": timedelta(hours=6),
    "24h": timedelta(hours=24),
    "7d": timedelta(days=7),
}


@dataclass(frozen=True)
class Row:
    service: str
    checks: int
    up: int
    slow: int
    down: int
    p50_ms: float
    p95_ms: float
    last_error: str

    @property
    def uptime(self) -> float:
        if self.checks == 0:
            return 0.0
        return (self.up + self.slow) / self.checks * 100.0


def parse_since(value: str) -> datetime:
    """Turn 6h, 30m, or an ISO timestamp into a UTC start time."""
    now = datetime.now(UTC)

    if value in WINDOWS:
        return now - WINDOWS[value]

    try:
        parsed = datetime.fromisoformat(value)
    except ValueError as error:
        raise argparse.ArgumentTypeError(f"bad --since {value!r}") from error

    if parsed.tzinfo is None:
        return parsed.replace(tzinfo=UTC)
    return parsed.astimezone(UTC)


def percentile(values: list[int], fraction: float) -> float:
    """Return the nearest-rank percentile in milliseconds."""
    if not values:
        return 0.0
    ordered = sorted(values)
    index = min(len(ordered) - 1, round(fraction * (len(ordered) - 1)))
    return ordered[index] / 1000.0


def summarise(conn: sqlite3.Connection, since: datetime, tag: str | None) -> list[Row]:
    """Aggregate one window into a row per service."""
    query = """
        SELECT service, state, status, total_us, err
        FROM probes
        WHERE at >= ?
        ORDER BY service, at
    """
    services: dict[str, dict] = {}

    for service, state, status, total_us, err in conn.execute(query, (int(since.timestamp()),)):
        bucket = services.setdefault(
            service,
            {"checks": 0, "up": 0, "slow": 0, "down": 0, "latency": [], "error": ""},
        )
        bucket["checks"] += 1
        bucket[state] = bucket.get(state, 0) + 1
        bucket["latency"].append(int(total_us))

        if err:
            bucket["error"] = f"{status or '—'} {err}"[:80]

    rows = [
        Row(
            service=service,
            checks=data["checks"],
            up=data["up"],
            slow=data["slow"],
            down=data["down"],
            p50_ms=percentile(data["latency"], 0.50),
            p95_ms=percentile(data["latency"], 0.95),
            last_error=data["error"],
        )
        for service, data in services.items()
    ]

    if tag:
        rows = [row for row in rows if tag in row.service]
    return sorted(rows, key=lambda row: row.p95_ms, reverse=True)


def render_table(rows: list[Row]) -> str:
    """Render a fixed-width table for a terminal."""
    header = f"{'SERVICE':<24} {'CHECKS':>6} {'UP%':>6} {'P50':>8} {'P95':>8}  ERROR"
    lines = [header, "-" * len(header)]
    for row in rows:
        lines.append(
            f"{row.service:<24} {row.checks:>6} {row.uptime:>6.1f} "
            f"{row.p50_ms:>7.1f} {row.p95_ms:>7.1f}  {row.last_error}"
        )
    return "\n".join(lines)


def render_csv(rows: list[Row], stream) -> None:
    """Write the same columns as CSV."""
    writer = csv.writer(stream)
    writer.writerow(
        ["service", "checks", "up", "slow", "down", "uptime_pct", "p50_ms", "p95_ms", "last_error"]
    )
    for row in rows:
        writer.writerow(
            [
                row.service,
                row.checks,
                row.up,
                row.slow,
                row.down,
                f"{row.uptime:.1f}",
                f"{row.p50_ms:.1f}",
                f"{row.p95_ms:.1f}",
                row.last_error,
            ]
        )


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--db", type=Path, default=Path("beacon.db"), help="SQLite database")
    parser.add_argument("--since", default="24h", help="15m, 1h, 6h, 24h, 7d, or ISO time")
    parser.add_argument("--tag", help="only services whose name contains this text")
    parser.add_argument("--format", choices=("table", "csv"), default="table")
    parser.add_argument("--limit", type=int, default=0, help="keep the N slowest rows")
    args = parser.parse_args(argv)

    if not args.db.exists():
        parser.error(f"no database at {args.db}")

    since = parse_since(args.since)
    connection = sqlite3.connect(f"file:{args.db}?mode=ro", uri=True)

    try:
        rows = summarise(connection, since, args.tag)
    except sqlite3.Error as error:
        print(f"query failed: {error}", file=sys.stderr)
        return 1
    finally:
        connection.close()

    if args.limit > 0:
        rows = rows[: args.limit]

    if not rows:
        print(f"no checks since {since.isoformat(timespec='seconds')}", file=sys.stderr)
        return 0

    if args.format == "csv":
        render_csv(rows, sys.stdout)
    else:
        print(render_table(rows))

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
