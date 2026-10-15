#!/bin/sh
# Build the cute robot eyes demo as an Android debug APK.
#
# Usage:
#   scripts/android-eyes.sh           bind the Go page, build the APK, print it
#   scripts/android-eyes.sh install   build, then adb install -r the APK
#
# One-time setup is the Android SDK (platform-tools, a platform, build-tools,
# an NDK) plus ebitenmobile. examples/telegram/android/README.md has the
# commands for a fresh machine, alongside the WSL2 and Windows-adb notes.
#
# From WSL2 the phone is visible to the Windows adb, not the Linux one:
#   ADB=/mnt/c/Users/<you>/platform-tools/adb.exe sh scripts/android-eyes.sh install
# When ADB ends in .exe the script hands the APK path over with wslpath -w.
set -eu

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
android="$root/examples/eyes/android"
aar="$android/eyes/eyes.aar"
apk="$android/app/build/outputs/apk/debug/app-debug.apk"

# A phone runs one ABI. The default is arm64, which Pixel-class phones run.
# Bind every Android ABI with ANDROID_TARGET=android for an emulator-ready
# APK, or android/arm for a 32-bit phone.
target=${ANDROID_TARGET:-android/arm64}

adb=${ADB:-adb}

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
	-javapkg com.chinmaysawant.eyes \
	-o "$aar" ./examples/eyes/mobile

echo "==> ./gradlew :app:assembleDebug"
(cd "$android" && ./gradlew --no-daemon :app:assembleDebug)

echo "APK: $apk"

if [ "${1:-}" = "install" ]; then
	if ! command -v "$adb" >/dev/null 2>&1 && [ ! -x "$adb" ]; then
		echo "android: adb is not on PATH; set ADB to the adb executable" >&2
		exit 1
	fi

	install_path=$apk
	case "$adb" in
	*.exe) install_path=$(wslpath -w "$apk") ;;
	esac

	"$adb" install -r "$install_path"
fi
