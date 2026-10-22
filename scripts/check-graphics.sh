#!/usr/bin/env bash
# Each RunGame test gets its own process. Xvfb is the default session.
set -euo pipefail
cd -- "$(dirname -- "$0")/.."
session=${GRAPHICS_SESSION:-xvfb}
runner=()
case "$session" in
    xvfb)
        for tool in xvfb-run Xvfb xauth; do
            command -v "$tool" >/dev/null || { echo "graphics-check: install $tool (Ubuntu: xvfb xauth libgl1-mesa-dri)" >&2; exit 1; }
        done
        export LIBGL_ALWAYS_SOFTWARE=1
        runner=(xvfb-run -a -s '-screen 0 1920x1080x24')
        ;;
    desktop)
        [[ -n ${DISPLAY:-} ]] || { echo 'graphics-check: desktop mode requires DISPLAY' >&2; exit 1; }
        ;;
    *) echo "graphics-check: unknown GRAPHICS_SESSION: $session (use xvfb or desktop)" >&2; exit 1 ;;
esac
echo "graphics-check: session=$session; DISPLAY=${DISPLAY:-unset}; LIBGL_ALWAYS_SOFTWARE=${LIBGL_ALWAYS_SOFTWARE:-unset}"
export EBITENGINE_GRAPHICS_LIBRARY=opengl
export OWNFRAME_FRAME_GPU_TEST=1
export OWNFRAME_REPLAY_GPU_TEST=1
export OWNFRAME_SCROLLBAR_GPU_TEST=1

run_test() {
    # Do not set HOME or touch the user's Xauthority. xvfb-run owns a temporary cookie.
    "${runner[@]}" \
        go test -p 1 -count=1 -json -timeout=90s -run "^$2$" "$1" |
        python3 -c '
import json, sys
passed = False
for line in sys.stdin:
    event = json.loads(line)
    if "Output" in event:
        print(event["Output"], end="")
    if event.get("Test") == sys.argv[1] and event["Action"] == "pass":
        passed = True
if not passed:
    sys.exit("graphics-check: required test did not pass: " + sys.argv[1])
' "$2"
}

# A missing, renamed, or skipped test is a failure, even when go test exits zero.
run_test ./internal/window TestPartialFrameLeavesTheScreen
run_test ./internal/window TestWindowRepaintPixels
run_test ./internal/window TestWindowIdleDeadline
run_test ./internal/window TestDrawPaintsVerticalThumb
run_test ./internal/replay TestGPUVisibleParity
