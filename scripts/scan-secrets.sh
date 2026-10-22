#!/bin/sh
# GEN-42: release secret scan. API keys stay server side: never bake them
# into -ldflags -X, //go:embed, or the wasm bundle, all of which are
# trivially extractable. This script scans tracked source with gitleaks
# (when installed) plus a high-confidence pattern grep, and scans built
# artifacts in dist/ with `strings` for the same patterns.
#
# Usage: sh scripts/scan-secrets.sh
set -eu

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$root"

fail=0

# 1. Source scan: gitleaks when present, pattern grep always.
if command -v gitleaks >/dev/null 2>&1; then
	gitleaks detect --config .gitleaks.toml --no-git || fail=1
else
	echo "scan-secrets: gitleaks not found, running pattern grep only" >&2
fi

patterns='AKIA[0-9A-Z]{16}|ghp_[A-Za-z0-9]{36}|xox[bpas]-[A-Za-z0-9-]+|sk-live-[A-Za-z0-9]+|AIza[0-9A-Za-z_-]{35}|-----BEGIN .*PRIVATE KEY-----'
if git grep -nE "$patterns" -- . ':!frontend' ':!plans' ':!documentation' ':!.gitleaks.toml' ':!scripts/scan-secrets.sh'; then
	echo "scan-secrets: secret-like pattern in source" >&2
	fail=1
fi

# 2. Artifact scan: strings over release archives and unpacked binaries.
if [ -d dist ]; then
	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT INT TERM
	for archive in dist/*; do
		[ -f "$archive" ] || continue
		case $archive in
		*.tar.gz) tar -xzf "$archive" -C "$tmp" ;;
		*.zip) unzip -q -o "$archive" -d "$tmp" ;;
		*) continue ;;
		esac
	done
	if grep -rE "$patterns" "$tmp" 2>/dev/null; then
		echo "scan-secrets: secret-like pattern in dist artifacts" >&2
		fail=1
	fi
else
	echo "scan-secrets: no dist/, skipping artifact scan (run scripts/package.sh first)" >&2
fi

exit $fail
