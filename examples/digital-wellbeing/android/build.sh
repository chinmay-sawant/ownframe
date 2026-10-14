#!/bin/sh
# Bind and build the Digital Wellbeing Android app; optionally install it.
set -eu

example=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
root=$(CDPATH= cd -- "$example/../.." && pwd)
android="$example/android"
aar="$android/digital-wellbeing/digital-wellbeing.aar"
apk="$android/app/build/outputs/apk/debug/app-debug.apk"
target=${ANDROID_TARGET:-android/arm64}
adb=${ADB:-adb}

if ! command -v ebitenmobile >/dev/null 2>&1; then
	echo "android: ebitenmobile is not on PATH" >&2
	exit 1
fi
if [ -z "${ANDROID_HOME:-}" ] && [ -z "${ANDROID_SDK_ROOT:-}" ]; then
	if [ -d "$HOME/Android/Sdk" ]; then
		ANDROID_HOME="$HOME/Android/Sdk"
	else
		echo "android: set ANDROID_HOME to the installed Android SDK" >&2
		exit 1
	fi
fi
export ANDROID_HOME="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-}}"

# Keep the workspace unset: gomobile builds its generated temporary module.
cd "$root"
ebitenmobile bind -target "$target" \
	-ldflags "-s -w -extldflags '-Wl,-z,max-page-size=16384 -Wl,-z,common-page-size=16384'" \
	-javapkg com.chinmaysawant.digitalwellbeing \
	-o "$aar" ./examples/digital-wellbeing/mobile
(cd "$android" && ./gradlew --no-daemon :app:assembleDebug)
echo "APK: $apk"

if [ "${1:-}" = install ]; then
	if ! command -v "$adb" >/dev/null 2>&1 && [ ! -x "$adb" ]; then
		echo "android: set ADB to the installed adb executable" >&2
		exit 1
	fi
	install_path=$apk
	case "$adb" in
	*.exe) install_path=$(wslpath -w "$apk") ;;
	esac
	"$adb" install -r "$install_path"
fi
