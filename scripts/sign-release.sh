#!/bin/sh
# Sign packaged executables; real credentials and native signing tools required.
# Use --platform macos or --platform windows when signing on separate hosts.
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
dist=${OWNFRAME_DIST_DIR:-"$root/dist"}
python3 "$root/scripts/sign_release.py" --directory "$dist" "$@"
if [ "${COSIGN:-0}" = 1 ]; then
    (cd "$dist" && cosign sign-blob --yes --output-signature SHA256SUMS.sig --output-certificate SHA256SUMS.pem SHA256SUMS)
fi
