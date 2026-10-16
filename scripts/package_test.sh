#!/bin/sh
# Check the dry-run layout of scripts/package.sh against the layouts in
# documentation/packaging.md. No build and no network, so it stays fast.
set -eu

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$root"

fail=0

want() { # $1 output, $2 literal line, $3 label
	case $1 in
	*"$2"*) ;;
	*)
		echo "package layout: missing $3: $2" >&2
		fail=1
		;;
	esac
}

linux=$(GOOS=linux sh scripts/package.sh -n login)
want "$linux" "archive: dist/ownframe-login-linux-" "Linux archive"
want "$linux" "  ownframe-login/login" "Linux binary"
want "$linux" "  ownframe-login/login.desktop" "Linux desktop entry"
want "$linux" "  ownframe-login/README.md" "Linux README"
want "$linux" "  ownframe.wasm" "wasm binary"
want "$linux" "  wasm_exec.js" "wasm loader"
want "$linux" "  index.html (browser/index.html)" "wasm page"
want "$linux" "checksums: dist/SHA256SUMS" "checksums"

darwin=$(GOOS=darwin sh scripts/package.sh -n login)
want "$darwin" "archive: dist/ownframe-login-macos-" "macOS archive"
want "$darwin" "  ownframe-login.app/Contents/Info.plist" "macOS plist"
want "$darwin" "  ownframe-login.app/Contents/MacOS/login" "macOS binary"

windows=$(GOOS=windows sh scripts/package.sh -n login)
want "$windows" "archive: dist/ownframe-login-windows-" "Windows archive"
want "$windows" "  ownframe-login.exe" "Windows executable"

if [ "$fail" -ne 0 ]; then
	exit 1
fi

echo "package layout ok"
