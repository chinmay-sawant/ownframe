#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
android="$root/examples/pet-care/android"
apk="$android/app/build/outputs/apk/debug/app-debug.apk"
artifact="$root/temp/pet-care/pet-care-debug.apk"
target=${ANDROID_TARGET:-android}

if ! command -v ebitenmobile >/dev/null 2>&1; then
	echo "pet-care: install ebitenmobile v2.10.4 first" >&2
	exit 1
fi
if [ -z "${ANDROID_HOME:-}" ] && [ -z "${ANDROID_SDK_ROOT:-}" ]; then
	echo "pet-care: set ANDROID_HOME to the Android SDK" >&2
	exit 1
fi
export ANDROID_HOME="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-}}"
cd "$root"
	ebitenmobile bind -target "$target" \
	-ldflags "-s -w -extldflags '-Wl,-z,max-page-size=16384 -Wl,-z,common-page-size=16384'" \
	-javapkg com.chinmaysawant.petcare \
	-o "$android/pet-care.aar" ./examples/pet-care/mobile
(cd "$android" && ./gradlew --no-daemon :app:assembleDebug)
mkdir -p "$(dirname "$artifact")"
mv "$apk" "$artifact"
echo "APK: $artifact"
if [ "${1:-}" = install ]; then
	if ! command -v adb >/dev/null 2>&1; then echo "pet-care: adb is not on PATH" >&2; exit 1; fi
	adb install -r "$artifact"
fi
