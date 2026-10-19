#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
blinkless=$(CDPATH= cd -- "$root/../blinkless" && pwd)
tmp=$(mktemp -d "${TMPDIR:-/tmp}/ownframe-css-preview.XXXXXX")
trap 'rm -rf "$tmp"' EXIT HUP INT TERM

(
	cd "$tmp"
	go work init "$root" "$root/examples" "$blinkless"
)
cd "$root"
mode=${1:-all}
case "$mode" in
	focused)
		GOWORK="$tmp/go.work" go vet -p 1 ./internal/csspreview ./cmd/css-preview
		GOWORK="$tmp/go.work" go test -p 1 ./internal/csspreview ./cmd/css-preview
		;;
	test)
		GOWORK="$tmp/go.work" go vet -p 1 ./internal/csspreview ./cmd/css-preview
		GOWORK="$tmp/go.work" make test
		;;
	all|animation|transition)
		GOWORK="$tmp/go.work" go run ./cmd/css-preview "$mode"
		;;
	*)
		printf 'usage: %s [all|animation|transition|focused|test]\n' "$0" >&2
		exit 2
		;;
esac
