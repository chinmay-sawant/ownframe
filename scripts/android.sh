#!/bin/sh
# Build the Telegram demo as an Android debug APK.
#
# Usage:
#   scripts/android.sh           bind the Go page, build the APK, print it
#   scripts/android.sh install   build, then adb install -r the APK
#
# One-time setup is the Android SDK (platform-tools, a platform, build-tools,
# an NDK) plus ebitenmobile. examples/telegram/android/README.md has the
# commands for a fresh machine.
set -eu

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
android="$root/examples/telegram/android"
aar="$android/telegram/telegram.aar"
apk="$android/app/build/outputs/apk/debug/app-debug.apk"

# A phone runs one ABI. The default binds every Android ABI (arm64-v8a,
# armeabi-v7a, x86, x86_64) so the APK runs anywhere, including an emulator.
# Narrow it for a much smaller APK, for example:
#   ANDROID_TARGET=android/arm64 sh scripts/android.sh install
target=${ANDROID_TARGET:-android}

if ! command -v ebitenmobile >/dev/null 2>&1; then
	echo "android: ebitenmobile is not on PATH; install it with" >&2
	echo "  go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@v2.10.4" >&2
	exit 1
fi

if [ -z "${ANDROID_HOME:-}" ] && [ -z "${ANDROID_SDK_ROOT:-}" ]; then
	if [ -d "$HOME/Android/Sdk" ]; then
		ANDROID_HOME="$HOME/Android/Sdk"
	else
		echo "android: set ANDROID_HOME to the Android SDK (see the README)" >&2
		exit 1
	fi
fi

export ANDROID_HOME="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-}}"

# gomobile resolves the packages under the current directory, where the
# committed go.work points ownframe at this checkout. Do not export GOWORK:
# the generated temp module it builds in cannot join the workspace.
cd "$root"

# The extldflags align the shared library for 16 KB page devices (NDK r27 and
# lower need them; NDK r28 defaults to this) and -s -w drops the Go symbols.
echo "==> ebitenmobile bind -target $target"
ebitenmobile bind -target "$target" \
	-ldflags "-s -w -extldflags '-Wl,-z,max-page-size=16384 -Wl,-z,common-page-size=16384'" \
	-javapkg com.chinmaysawant.telegram \
	-o "$aar" ./examples/telegram/mobile

echo "==> ./gradlew :app:assembleDebug"
(cd "$android" && ./gradlew --no-daemon :app:assembleDebug)

echo "APK: $apk"

if [ "${1:-}" = "install" ]; then
	adb=${ADB:-adb}
	if ! command -v "$adb" >/dev/null 2>&1; then
		echo "android: adb is not on PATH; install platform-tools" >&2
		exit 1
	fi

	install_path=$apk
	case "$adb" in
	*.exe) install_path=$(wslpath -w "$apk") ;;
	esac
	"$adb" -P "${ADB_PORT:-5037}" install -r "$install_path"
fi
