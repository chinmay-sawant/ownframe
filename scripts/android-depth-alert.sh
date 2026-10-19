#!/bin/sh
# Bind the OwnFrame page and build the Pixel 7 depth-alert debug APK.
set -eu

root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
android="$root/examples/depth-alert/android"
aar="$android/depth-alert/depth-alert.aar"
apk="$android/app/build/outputs/apk/debug/app-debug.apk"
target=${ANDROID_TARGET:-android/arm64}
adb=${ADB:-adb}

if ! command -v ebitenmobile >/dev/null 2>&1; then
	echo "android: ebitenmobile is not on PATH; install it with" >&2
	echo "  go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@v2.10.4" >&2
	exit 1
fi
if [ -z "${ANDROID_HOME:-}" ] && [ -z "${ANDROID_SDK_ROOT:-}" ]; then
	echo "android: set ANDROID_HOME to the Android SDK" >&2
	exit 1
fi
export ANDROID_HOME="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-}}"

# Keep GOWORK unset. The bind must start at the repository root.
cd "$root"
unset GOWORK
ebitenmobile bind -target "$target" \
	-ldflags "-s -w -extldflags '-Wl,-z,max-page-size=16384 -Wl,-z,common-page-size=16384'" \
	-javapkg com.chinmaysawant.depthalert -o "$aar" ./examples/depth-alert/mobile
rm -f "$android/depth-alert/depth-alert-sources.jar"
(cd "$android" && ./gradlew --no-daemon --max-workers=2 :app:assembleDebug)
echo "APK: $apk"

if [ "${1:-}" = "install" ]; then
	if ! command -v "$adb" >/dev/null 2>&1 && [ ! -x "$adb" ]; then
		echo "android: set ADB to the adb executable" >&2
		exit 1
	fi
	install_path=$apk
	case "$adb" in
	*.exe) install_path=$(wslpath -w "$apk") ;;
	esac
	"$adb" install -r "$install_path"
fi
