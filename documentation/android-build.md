# Build and check the Android application

Run these commands from the repository root. Keep `go.work` in place and leave
`GOWORK` unset for the mobile bind, because gomobile builds a temporary module.

1. Check both Go modules.

   ```sh
   make build BUILD_P=4
   make test TEST_P=4
   ```

   On a machine without a display, prefix these commands with `xvfb-run -a`.

2. Build Telegram with the existing mobile pipeline.

   ```sh
   sh scripts/android.sh
   ```

   The script binds all Android ABIs by default and writes
   `examples/telegram/android/app/build/outputs/apk/debug/app-debug.apk`.
   For an arm64 device, use `ANDROID_TARGET=android/arm64 sh scripts/android.sh`.

3. Check that a device is connected, then install the debug application.

   ```sh
   adb devices
   sh scripts/android.sh install
   ```

4. Check the login form, a dashboard, a long settings page, and a modal on the
   device. Rotate during editing and scrolling. Open the keyboard, dismiss it
   with Back, and verify that the draft stays. Repeat navigation and background
   and resume cycles. Check status bars, navigation bars, and lateral cutouts
   at different device densities. Verify that a cancelled gesture never clicks.

5. Capture diagnostics when needed.

   ```sh
   adb logcat
   adb shell dumpsys meminfo com.chinmaysawant.telegram
   OWNFRAME_REPLAY_GPU_TEST=1 xvfb-run -a go test ./internal/replay -run TestGPUVisibleParity -count=1
   ```

   The GPU comparison runs on desktop and checks replay and culling. It does
   not validate Android keyboard behavior or a physical device's performance.
   Use Android profiling tools for device CPU, frame pacing, and GPU resources.
   Compare measurements on the same device and build before claiming a change.
