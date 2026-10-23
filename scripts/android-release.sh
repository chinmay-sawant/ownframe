#!/bin/sh
# Build and sign one example's arm64 Android release APK.
set -eu

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
if [ "$#" -ne 1 ]; then
	echo "usage: scripts/android-release.sh <login|platform|dino|lg-remote|telegram|eyes|depth-alert>" >&2
	exit 2
fi

app=$1
case "$app" in
login|platform|dino|telegram)
	dir=$app
	aar=$app
	pkg="com.chinmaysawant.$app"
	;;
eyes)
	dir=$app
	aar=$app
	pkg=com.chinmaysawant.eyes
	;;
depth-alert)
	dir=$app
	aar=$app
	pkg=com.chinmaysawant.depthalert
	;;
lg-remote)
	dir=$app
	aar=lgremote
	pkg=com.chinmaysawant.lgremote
	;;
*)
	echo "android: unknown mobile example: $app" >&2
	exit 2
	;;
esac

sdk=${ANDROID_HOME:-${ANDROID_SDK_ROOT:-}}
if [ -z "$sdk" ]; then
	echo "android: set ANDROID_HOME to the Android SDK" >&2
	exit 1
fi
export ANDROID_HOME=$sdk

keydir="$root/temp/android-release-signing"
keystore=${ANDROID_KEYSTORE:-"$keydir/ownframe-release.jks"}
alias=${ANDROID_KEY_ALIAS:-ownframe-release}
storepass=${ANDROID_STORE_PASS_FILE:-"$keydir/store-pass.txt"}
keypass=${ANDROID_KEY_PASS_FILE:-"$keydir/key-pass.txt"}
for file in "$keystore" "$storepass" "$keypass"; do
	if [ ! -r "$file" ]; then
		echo "android: signing file not found or unreadable: $file" >&2
		exit 1
	fi
done

target=${ANDROID_TARGET:-android/arm64}
build_tools=${ANDROID_BUILD_TOOLS_VERSION:-35.0.0}
case "$target" in
android/arm64) abi=arm64 ;;
android/arm) abi=arm ;;
android) abi=all ;;
*) abi=$(printf '%s' "$target" | tr '/,' '--') ;;
esac

android="$root/examples/$dir/android"
cd "$root"
# gomobile copies packages into a temp module that cannot join go.work.
unset GOWORK
umask 077
source_commit=$(git rev-parse HEAD)
engine=$(go list -m -f '{{.Version}}' github.com/chinmay-sawant/blinkless)
source_dirty=$(git status --porcelain --untracked-files=no)
ebitenmobile bind -trimpath -target "$target" \
	-ldflags "-s -w -extldflags '-Wl,-z,max-page-size=16384 -Wl,-z,common-page-size=16384'" \
	-javapkg "$pkg" -o "$android/$aar/$aar.aar" "./examples/$dir/mobile"
(cd "$android" && ./gradlew --no-daemon --max-workers=2 :app:assembleRelease)

artifact="$android/artifacts/$app-$abi-release.apk"
mkdir -p -m 0700 "$(dirname "$artifact")"
"$sdk/build-tools/$build_tools/apksigner" sign \
	--ks "$keystore" --ks-key-alias "$alias" \
	--ks-pass "file:$storepass" --key-pass "file:$keypass" \
	--out "$artifact" "$android/app/build/outputs/apk/release/app-release-unsigned.apk"
"$sdk/build-tools/$build_tools/apksigner" verify --verbose "$artifact"
export source_commit engine source_dirty artifact
python3 - <<'META'
import hashlib
import json
import os
from pathlib import Path
import subprocess
apk = Path(os.environ['artifact'])
report = {
    'source_commit': os.environ['source_commit'],
    'source_dirty': bool(os.environ['source_dirty']),
    'engine': os.environ['engine'],
    'go': subprocess.check_output(['go', 'version'], text=True).strip(),
    'sha256': hashlib.sha256(apk.read_bytes()).hexdigest(),
}
apk.with_suffix('.json').write_text(json.dumps(report, indent=2) + '\n')
META
"$sdk/build-tools/$build_tools/apksigner" verify --print-certs "$artifact" > "$artifact.cert.txt"
echo "APK: $artifact"
