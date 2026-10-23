#!/bin/sh
# Build one example into a release archive for this system.
#
# Usage:
#   scripts/package.sh [-n] <example>
#
#   -n  dry run: print the archive names and their entries, build nothing
#
# The archives land in dist/ and SHA256SUMS sits next to them. Linux gets a
# tar.gz with the binary, a .desktop entry, and the README. macOS gets a zip
# with an .app bundle. Windows gets a zip with the .exe. Every system also
# gets the wasm zip. No network, no new modules.
set -eu

# Packaging uses the already verified module and toolchain caches.
export GOPROXY=off
# Resolve an already cached toolchain before forbidding further selection.
package_goroot=$(go env GOROOT)
if command -v cygpath >/dev/null 2>&1; then
	package_goroot=$(cygpath -u "$package_goroot")
fi
PATH="$package_goroot/bin:$PATH"
export PATH
export GOTOOLCHAIN=local

usage() {
	echo "usage: scripts/package.sh [-n] <example>" >&2
	exit 2
}

dry=0
while [ $# -gt 0 ]; do
	case $1 in
	-n) dry=1; shift ;;
	-*) usage ;;
	*) break ;;
	esac
done

[ $# -eq 1 ] || usage
example=$1

case $example in
	'' | */* | .*) usage ;;
esac

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$root"

[ -f "examples/$example/main.go" ] || {
	echo "package.sh: no example examples/$example/main.go" >&2
	exit 1
}

target_os=$(go env GOOS)
target_arch=$(go env GOARCH)
name="ownframe-$example"
dist=${OWNFRAME_DIST_DIR:-"$root/dist"}
SOURCE_DATE_EPOCH=${SOURCE_DATE_EPOCH:-$(git log -1 --format=%ct)}
export SOURCE_DATE_EPOCH

show() { # $1 archive, $2.. entries
	printf 'archive: %s\n' "$1"
	shift
	for entry in "$@"; do
		printf '  %s\n' "$entry"
	done
}

show_wasm() {
	show "dist/$name-wasm.zip" \
		ownframe.wasm wasm_exec.js "index.html (browser/index.html)"
}

case $target_os in
linux)
	show "dist/$name-linux-$target_arch.tar.gz" \
		"$name/$example" "$name/$example.desktop" "$name/README.md"
	;;
darwin)
	show "dist/$name-macos-$target_arch.zip" \
		"$name.app/Contents/Info.plist" "$name.app/Contents/MacOS/$example"
	;;
windows)
	show "dist/$name-windows-$target_arch.zip" "$name.exe"
	;;
*)
	echo "package.sh: no archive layout for GOOS=$target_os" >&2
	exit 1
	;;
esac

show_wasm
printf 'checksums: dist/SHA256SUMS\n'

if [ "$dry" -eq 1 ]; then
	exit 0
fi

need() {
	command -v "$1" >/dev/null 2>&1 || {
		echo "package.sh: $1 is required to build the $2 archive" >&2
		exit 1
	}
}

build() { # $1 output, rest source packages
	out=$1
	shift
	# GEN-16: position-independent binaries plus C stack and link
	# hardening where the toolchain supports them. Windows gets ASLR
	# from the OS loader, so pie stays off there; check with
	# `checksec --file=` on the result. See documentation/packaging.md.
	pie=""
	link_hardening=""
	if [ "$target_os" = linux ]; then
		link_hardening="-Wl,-z,relro,-z,now"
	fi
	case $target_os in
	linux | darwin) pie="-buildmode=pie" ;;
	esac
	env GOOS="$target_os" GOARCH="$target_arch" \
		CGO_CFLAGS="-O2 -D_FORTIFY_SOURCE=2 -fstack-protector-strong" \
		CGO_LDFLAGS="$link_hardening" \
		go build -C "$root" -trimpath $pie -ldflags "-s -w" -o "$out" "$@" "./examples/$example"
}

desktop_entry() {
	cat > "$1" <<EOF
[Desktop Entry]
Type=Application
Name=ownframe $example
Exec=$example
Terminal=false
Categories=Utility;
EOF
}

info_plist() {
	cat > "$1" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key><string>ownframe $example</string>
	<key>CFBundleIdentifier</key><string>dev.ownframe.$example</string>
	<key>CFBundleExecutable</key><string>$example</string>
	<key>CFBundlePackageType</key><string>APPL</string>
	<key>CFBundleShortVersionString</key><string>0.0.2</string>
</dict>
</plist>
EOF
}

mkdir -p "$dist"
staging=$(mktemp -d)
trap 'rm -rf "$staging"' EXIT

case $target_os in
linux)
	need python3 "Linux"
	stage="$staging/$name"
	mkdir -p "$stage"
	build "$stage/$example"
	desktop_entry "$stage/$example.desktop"
	cp "$root/README.md" "$stage/README.md"
	python3 "$root/scripts/release_archive.py" pack "$staging" "$dist/$name-linux-$target_arch.tar.gz" "$name"
	;;
darwin)
	need python3 "macOS"
	app="$staging/$name.app"
	mkdir -p "$app/Contents/MacOS"
	build "$app/Contents/MacOS/$example"
	info_plist "$app/Contents/Info.plist"
	python3 "$root/scripts/release_archive.py" pack "$staging" "$dist/$name-macos-$target_arch.zip" "$name.app"
	;;
windows)
	need python3 "Windows"
	build "$staging/$name.exe"
	python3 "$root/scripts/release_archive.py" pack "$staging" "$dist/$name-windows-$target_arch.zip" "$name.exe"
	;;
esac

need python3 "wasm"
wasm_stage="$staging/wasm"
mkdir -p "$wasm_stage"
env GOOS=js GOARCH=wasm \
	go build -C "$root" -trimpath -ldflags "-s -w" -o "$wasm_stage/ownframe.wasm" "./examples/$example"

goroot=$(go env GOROOT)
if [ -f "$goroot/lib/wasm/wasm_exec.js" ]; then
	cp "$goroot/lib/wasm/wasm_exec.js" "$wasm_stage/wasm_exec.js"
else
	cp "$goroot/misc/wasm/wasm_exec.js" "$wasm_stage/wasm_exec.js"
fi
cp "$root/browser/index.html" "$wasm_stage/index.html"
python3 "$root/scripts/release_archive.py" pack "$wasm_stage" "$dist/$name-wasm.zip" ownframe.wasm wasm_exec.js index.html

python3 "$root/scripts/release_archive.py" checksums "$dist"

# GEN-12: keyless cosign signature over the checksums. Opt in with
# COSIGN=1; the default stays unsigned so local builds need no network.
if [ "${COSIGN:-0}" = 1 ]; then
	need cosign "signing"
	(cd "$dist" && cosign sign-blob --yes \
		--output-signature SHA256SUMS.sig \
		--output-certificate SHA256SUMS.pem SHA256SUMS)
fi

echo "wrote $dist"
