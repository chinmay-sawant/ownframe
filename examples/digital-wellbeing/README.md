# Digital Wellbeing with Mochi

Mochi opens to a seven-day overview of foreground time across apps. You can
then focus the chart on one app. Android supplies the totals through
`UsageStatsManager`; the app asks you to enable Usage Access in system settings
before reading them.

The picker shows launchable apps with activity in the last seven days, sorted
from highest foreground time to lowest. It passes app names to the picker,
then requests and displays daily totals for only the selected package. The
totals stay on the phone. The app has no network permission. Mochi's
illustration was created for this example.

## Build and install

From WSL2, set `ADB` to your Windows `platform-tools/adb.exe` path:

```sh
ADB=/mnt/c/Users/your-windows-user/platform-tools/adb.exe sh examples/digital-wellbeing/android/build.sh install
```

The script binds the Go page, assembles a debug APK, and installs it with
`adb install -r`. Set `ANDROID_HOME` if the SDK is not at `~/Android/Sdk`.
It defaults to `android/arm64`, suitable for the Pixel 7.

Open Digital Wellbeing, tap **Open Usage Access**, and enable access for this
app in Android Settings. Return to Mochi to see the overview. Tap **Choose an
app** to focus the chart on one app. Usage data refreshes every 30 seconds while
the activity is open.

The generated APK is at
`android/app/build/outputs/apk/debug/app-debug.apk`. Generated AARs and Gradle
outputs are local build artifacts.
