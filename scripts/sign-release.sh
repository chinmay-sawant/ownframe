#!/bin/sh
# Sign and notarize the dist/ archives from scripts/package.sh.
#
# This runs in CI only, behind the protected release-signing
# environment that holds the keys. Never run it on a dev laptop and
# never commit keys or tokens: every secret arrives as an environment
# variable. macOS needs APPLE_DEVELOPER_ID plus NOTARY_KEY_ID,
# NOTARY_ISSUER_ID, and NOTARY_KEY for notarytool; Windows needs
# AZURE_TENANT_ID, AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, and
# AZURE_VAULT_URL for AzureSignTool (HSM-backed, no local key file).
# Missing platform tooling skips that platform with a message.
# See documentation/packaging.md.
set -eu

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
dist="$root/dist"

[ -d "$dist" ] || {
	echo "sign-release: no dist/ directory; run scripts/package.sh first" >&2
	exit 1
}

sign_macos() {
	command -v codesign >/dev/null 2>&1 || {
		echo "sign-release: codesign not found, skipping macOS"
		return 0
	}
	[ -n "${APPLE_DEVELOPER_ID:-}" ] || {
		echo "sign-release: APPLE_DEVELOPER_ID unset, skipping macOS"
		return 0
	}
	for app in "$dist"/*.app; do
		[ -e "$app" ] || continue
		codesign --deep --force --options runtime \
			--sign "$APPLE_DEVELOPER_ID" "$app"
		codesign --verify --deep --strict "$app"
	done
}

notarize_macos() {
	command -v xcrun >/dev/null 2>&1 || return 0
	[ -n "${NOTARY_KEY_ID:-}" ] || {
		echo "sign-release: notary credentials unset, skipping notarization"
		return 0
	}
	for zip in "$dist"/*macos*.zip; do
		[ -e "$zip" ] || continue
		xcrun notarytool submit "$zip" --key "$NOTARY_KEY" \
			--key-id "$NOTARY_KEY_ID" --issuer "$NOTARY_ISSUER_ID" \
			--wait
		xcrun stapler staple "$zip"
	done
}

sign_windows() {
	command -v AzureSignTool >/dev/null 2>&1 || {
		echo "sign-release: AzureSignTool not found, skipping Windows"
		return 0
	}
	for req in AZURE_TENANT_ID AZURE_CLIENT_ID AZURE_CLIENT_SECRET AZURE_VAULT_URL AZURE_CERT_NAME; do
		eval "test -n \"\${$req:-}\" || { echo \"sign-release: $req unset, skipping Windows\"; return 0; }"
	done
	for exe in "$dist"/*.exe; do
		[ -e "$exe" ] || continue
		AzureSignTool sign -kvu "$AZURE_VAULT_URL" -kvi "$AZURE_CLIENT_ID" \
			-kvt "$AZURE_TENANT_ID" -kvs "$AZURE_CLIENT_SECRET" \
			-kvc "$AZURE_CERT_NAME" -tr http://timestamp.digicert.com \
			-td sha256 "$exe"
	done
}

sign_macos
notarize_macos
sign_windows

echo "sign-release: done"
