#!/bin/sh
# Measure three window examples: build, open, sample RSS and CPU for a while,
# close. Then optionally run a -web pass and pull a heap profile.
#
# Usage: scripts/perf-apps.sh [--web] [--seconds N] [--out DIR] [--keep-web]
#   --web       after each window pass, start -web -perf and curl the heap
#   --seconds N sample length per app, default 20
#   --out DIR   write raw samples and a summary under DIR
#   --keep-web  leave the web server running (not for normal runs)
#
# Requires a Linux box with DISPLAY set for the window pass. Builds land in
# /tmp/opencode/perf-apps. Kills only the pids it started.
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
WORK=${TMPDIR:-/tmp}/opencode/perf-apps
DUR=20
WEB=0
KEEP_WEB=0
OUT=
WEB_WAIT=30

while [ $# -gt 0 ]; do
	case "$1" in
		--web) WEB=1; shift;;
		--keep-web) KEEP_WEB=1; shift;;
		--seconds) DUR="$2"; shift 2;;
		--out) OUT="$2"; shift 2;;
		-h|--help) sed -n '2,12p' "$0"; exit 0;;
		*) echo "perf-apps: unknown flag $1" >&2; exit 1;;
	esac
done

if [ "$WEB" -eq 0 ] && [ -z "${DISPLAY:-}" ]; then
	echo "perf-apps: DISPLAY is not set; the window pass needs an X or XWayland session" >&2
	exit 1
fi

mkdir -p "$WORK"
SUMMARY=
if [ -n "$OUT" ]; then
	mkdir -p "$OUT"
	SUMMARY="$OUT/perf-apps-summary.txt"
	: >"$SUMMARY"
fi

# One app per line: name, example dir, window title substring, web port.
APPS="departure-board examples/departure-board Departure 8137
spotify-player examples/spotify-player Spotify 8119
wispr-flow-dashboard examples/wispr-flow-dashboard Flow 8117"

KILL_PIDS=
DEAD=0
BUILD_OK=1
cleanup() {
	for p in $KILL_PIDS; do
		kill "$p" 2>/dev/null || true
	done
	sleep 1
	for p in $KILL_PIDS; do
		kill -9 "$p" 2>/dev/null || true
	done
	KILL_PIDS=
}
trap cleanup EXIT INT TERM HUP

# emit writes one line to stdout and to the summary file when --out is set.
emit() {
	if [ -n "$SUMMARY" ]; then
		printf '%s\n' "$1" | tee -a "$SUMMARY"
	else
		printf '%s\n' "$1"
	fi
}

PAGESIZE=$(getconf PAGESIZE 2>/dev/null || echo 4096)

rss_kb() {
	# /proc statm field 2 is resident pages. Fall back to ps.
	if [ -r "/proc/$1/statm" ]; then
		awk -v ps="$PAGESIZE" '{printf "%d\n", $2*ps/1024}' "/proc/$1/statm" 2>/dev/null && return 0
	fi
	ps -o rss= -p "$1" 2>/dev/null | tr -d ' ' || echo 0
}

cpu_ticks() {
	awk '{print $14+$15}' "/proc/$1/stat" 2>/dev/null || echo 0
}

wait_window() {
	pid=$1
	title=$2
	waited=0
	while [ "$waited" -lt 30 ]; do
		if ! kill -0 "$pid" 2>/dev/null; then
			return 1
		fi
		if command -v xwininfo >/dev/null 2>&1 &&
			xwininfo -root -tree 2>/dev/null | grep -F "$title" >/dev/null 2>&1; then
			return 0
		fi
		sleep 1
		waited=$((waited+1))
	done
	# No X tool or a native Wayland window: accept a live process after a
	# longer settle, and say so.
	echo "perf-apps: no X window named '$title' after 30 s; using the live process"
	kill -0 "$pid" 2>/dev/null
}

stop_pid() {
	pid=$1
	kill "$pid" 2>/dev/null || true
	n=0
	while [ "$n" -lt 5 ] && kill -0 "$pid" 2>/dev/null; do
		sleep 1
		n=$((n+1))
	done
	kill -9 "$pid" 2>/dev/null || true
	wait "$pid" 2>/dev/null || true
	KILL_PIDS=$(printf '%s\n' "$KILL_PIDS" | grep -v -x "$pid" || true)
}

tail_log() {
	echo "---- log tail: $1"
	tail -n 15 "$1" 2>/dev/null || true
	echo "---- end log"
}

# sample_window prints one row per second and sets DEAD when the process
# exits inside the window.
sample_window() {
	name=$1
	pid=$2
	raw=$3
	printf 't_seconds,rss_kb,cpu_percent\n' >"$raw"
	DEAD=0
	prev=$(cpu_ticks "$pid")
	i=1
	while [ "$i" -le "$DUR" ]; do
		sleep 1
		if ! kill -0 "$pid" 2>/dev/null; then
			DEAD=1
			return 0
		fi
		now=$(cpu_ticks "$pid")
		rss=$(rss_kb "$pid")
		cpu=$(awk -v d="$((now - prev))" -v h=100 'BEGIN {printf "%.1f", d/h*100}')
		prev=$now
		printf '%-24s %4s  %9s  %7s\n' "$name" "$i" "$rss" "$cpu"
		printf '%s,%s,%s\n' "$i" "$rss" "$cpu" >>"$raw"
		i=$((i+1))
	done
}

summarize() {
	name=$1
	raw=$2
	if [ ! -s "$raw" ]; then
		return
	fi
	awk -F, -v name="$name" '
		NR > 1 {
			rss += $2; cpu += $3
			if (NR == 2 || $2 > rssmax) rssmax = $2
			if (NR == 2 || $3 > cpumax) cpumax = $3
		}
		END {
			if (NR > 1)
				printf "%-24s %7d %10d %10d %12.1f %11.1f\n",
					name, NR-1, rss/(NR-1), rssmax, cpu/(NR-1), cpumax
		}' "$raw"
}

run_window() {
	name=$1
	dir=$2
	title=$3
	bin="$WORK/$name"
	log="$WORK/$name.log"
	raw="$WORK/$name-samples.csv"
	if [ -n "$OUT" ]; then
		raw="$OUT/$name-samples.csv"
	fi

	echo "== $name"
	BUILD_OK=1
	if ! (cd "$ROOT" && go build -p 1 -o "$bin" "./$dir") >"$log" 2>&1; then
		echo "$name: build failed"
		BUILD_OK=0
		tail_log "$log"
		return 0
	fi

	"$bin" -perf >>"$log" 2>&1 &
	pid=$!
	KILL_PIDS="$KILL_PIDS $pid"
	if ! wait_window "$pid" "$title"; then
		echo "$name: window did not appear or the process exited"
		tail_log "$log"
		stop_pid "$pid"
		return 0
	fi

	echo "sampling $name for $DUR s at 1 s intervals"
	sample_window "$name" "$pid" "$raw"
	if [ "$DEAD" -eq 1 ]; then
		echo "$name: process exited during sampling"
		tail_log "$log"
	fi
	stop_pid "$pid"
}

run_web() {
	name=$1
	port=$2
	bin="$WORK/$name"
	log="$WORK/$name-web.log"
	heap="$WORK/$name-heap.pprof"
	if [ -n "$OUT" ]; then
		heap="$OUT/$name-heap.pprof"
	fi

	"$bin" -web -perf -addr "127.0.0.1:$port" >>"$log" 2>&1 &
	pid=$!
	KILL_PIDS="$KILL_PIDS $pid"
	up=0
	n=0
	while [ "$n" -lt "$WEB_WAIT" ]; do
		if ! kill -0 "$pid" 2>/dev/null; then
			break
		fi
		if curl -sS -o /dev/null "http://127.0.0.1:$port/" 2>/dev/null; then
			up=1
			break
		fi
		sleep 1
		n=$((n+1))
	done
	if [ "$up" -eq 0 ]; then
		echo "$name web: server did not come up on port $port"
		tail_log "$log"
		stop_pid "$pid"
		return 0
	fi
	if curl -sS "http://127.0.0.1:$port/debug/pprof/heap" -o "$heap" 2>>"$log"; then
		size=$(wc -c <"$heap" | tr -d ' ')
		echo "$name web: heap profile $size bytes -> $heap"
	else
		echo "$name web: heap fetch failed"
		tail_log "$log"
	fi
	if [ "$KEEP_WEB" -eq 0 ]; then
		stop_pid "$pid"
	fi
}

emit "perf-apps $(date -u '+%Y-%m-%dT%H:%M:%SZ') duration=${DUR}s display=${DISPLAY:-none}"
printf '%-24s %4s  %9s  %7s\n' app t_s rss_kb cpu_pct

while IFS=' ' read -r name dir title port; do
	[ -n "$name" ] || continue
	run_window "$name" "$dir" "$title"
	if [ "$WEB" -eq 1 ] && [ "$BUILD_OK" -eq 1 ]; then
		run_web "$name" "$port"
	fi
done <<EOF
$APPS
EOF

emit ""
emit "summary: samples, mean and max per app"
ROW=$(printf '%-24s %7s %10s %10s %12s %11s' app samples rss_mean rss_max cpu_mean% cpu_max%)
emit "$ROW"
for name in departure-board spotify-player wispr-flow-dashboard; do
	raw="$WORK/$name-samples.csv"
	if [ -n "$OUT" ]; then
		raw="$OUT/$name-samples.csv"
	fi
	row=$(summarize "$name" "$raw")
	[ -n "$row" ] && emit "$row"
done
echo "perf-apps: done"
