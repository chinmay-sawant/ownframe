#!/usr/bin/env bash
# Drive a running ownframe window with the mouse and keyboard, then read it back
# as a picture. Pairs with an app flag such as `-snapshot 127.0.0.1:8120`, which
# serves GET /frame.png of the window as it is now (see documentation/testing.md,
# "Driving the real window").
#
# Usage: scripts/gui-drive.sh COMMAND [ARGS]
#   setup                    fetch xdotool into a local folder without root
#   info                     print the window id, position, size and click origin
#   click X Y                click at page pixel X,Y (viewport pixels, not document)
#   move X Y                 move the pointer to page pixel X,Y (hover)
#   down | up                press or release the left button where the pointer is
#   drag X0 Y0 X1 Y1         press at X0,Y0, move in six steps, release at X1,Y1
#   key NAME                 send one key, for example Down, Return, BackSpace, Tab
#   type TEXT                type text
#   scroll N                 wheel N notches down (negative N scrolls up), about 48px each,
#                            with the pointer parked in the page margin first
#   snap NAME                save the window picture to $GUI_OUT/NAME.png
#
# Settings (environment):
#   XDOTOOL           path to xdotool (default: PATH, then the folder `setup` fills)
#   GUI_TOOLS         where `setup` unpacks xdotool (default: $TMPDIR/ownframe-xdotool)
#   GUI_OUT           where `snap` writes (default: $TMPDIR/ownframe-gui)
#   GUI_WINDOW_NAME   window title to search for (default: "blinkless all")
#   GUI_SNAP_ADDR     snapshot server address (default: 127.0.0.1:8120)
#   GUI_SCROLL_X/Y    where `scroll` parks the pointer (default 992,600: the right margin)
#   GUI_OX, GUI_OY    screen position of page pixel 0,0. Default: found from the
#                     window position with the WSLg offset measured on the host this
#                     was written on. Check it with `click` on a known button, and set
#                     both when a click lands in the wrong place.
#
# Needs an X display (DISPLAY). Under WSLg the window is an XWayland window.
set -eu

tmp=${TMPDIR:-/tmp}
tools=${GUI_TOOLS:-$tmp/ownframe-xdotool}
out=${GUI_OUT:-$tmp/ownframe-gui}
name=${GUI_WINDOW_NAME:-blinkless all}
addr=${GUI_SNAP_ADDR:-127.0.0.1:8120}
export DISPLAY=${DISPLAY:-:0}

xdo_bin() {
	if [ -n "${XDOTOOL:-}" ]; then
		printf '%s\n' "$XDOTOOL"
	elif command -v xdotool >/dev/null 2>&1; then
		command -v xdotool
	elif [ -x "$tools/usr/bin/xdotool" ]; then
		export LD_LIBRARY_PATH="$tools/usr/lib/x86_64-linux-gnu${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
		printf '%s\n' "$tools/usr/bin/xdotool"
	else
		echo "gui-drive: xdotool not found; run: $0 setup (or install the xdotool package)" >&2
		exit 1
	fi
}

xd() { timeout 10 "$(xdo_bin)" "$@"; }

window() {
	id=$(xd search --name "$name" | head -n 1)
	if [ -z "$id" ]; then
		echo "gui-drive: no window named \"$name\"; start the app first" >&2
		exit 1
	fi
	printf '%s\n' "$id"
}

# origin sets OX and OY, the screen position of page pixel 0,0.
origin() {
	if [ -n "${GUI_OX:-}" ] && [ -n "${GUI_OY:-}" ]; then
		OX=$GUI_OX OY=$GUI_OY
		return
	fi
	geo=$(xd getwindowgeometry "$(window)" | sed -n 's/.*Position: \([0-9-]*\),\([0-9-]*\).*/\1 \2/p')
	set -- $geo
	OX=$(($1 - 32)) OY=$(($2 + 4))
}

cmd=${1:-}
[ $# -gt 0 ] && shift

case $cmd in
setup)
	mkdir -p "$tools/debs"
	(cd "$tools/debs" && apt-get download xdotool libxdo3)
	for deb in "$tools"/debs/*.deb; do dpkg -x "$deb" "$tools"; done
	echo "gui-drive: xdotool unpacked in $tools"
	;;
info)
	w=$(window)
	origin
	xd getwindowgeometry "$w"
	echo "page pixel 0,0 is screen $OX,$OY"
	;;
click)
	origin
	xd mousemove $((OX + $1)) $((OY + $2))
	sleep 0.2
	xd click 1
	;;
move)
	origin
	xd mousemove $((OX + $1)) $((OY + $2))
	;;
down) xd mousedown 1 ;;
up) xd mouseup 1 ;;
drag)
	origin
	xd mousemove $((OX + $1)) $((OY + $2))
	sleep 0.2
	xd mousedown 1
	sleep 0.2
	for i in 1 2 3 4 5 6; do
		xd mousemove $((OX + $1 + ($3 - $1) * i / 6)) $((OY + $2 + ($4 - $2) * i / 6))
		sleep 0.15
	done
	xd mouseup 1
	;;
# Keys go straight to the window (--window). Plain XTEST keys need X focus,
# and a WSLg session with no window manager never gives the window any.
key) xd key --window "$(window)" "$1" ;;
type) xd type --window "$(window)" --delay 80 "$1" ;;
scroll)
	# The wheel scrolls whatever is under the pointer, and a box with its own
	# scrolling (the action-scroll sheet) eats the notches. Park the pointer in the
	# right margin first so the page itself scrolls.
	origin
	xd mousemove $((OX + ${GUI_SCROLL_X:-992})) $((OY + ${GUI_SCROLL_Y:-600}))
	n=$1
	button=5
	if [ "$n" -lt 0 ]; then
		n=$((-n))
		button=4
	fi
	for _ in $(seq 1 "$n"); do
		xd click $button
		sleep 0.05
	done
	;;
snap)
	mkdir -p "$out"
	sleep 0.8
	curl -sf --max-time 20 -o "$out/$1.png" "http://$addr/frame.png"
	echo "$out/$1.png"
	;;
*)
	sed -n '2,/^set -eu/p' "$0" | sed '$d' | sed 's/^# \{0,1\}//' >&2
	exit 2
	;;
esac
