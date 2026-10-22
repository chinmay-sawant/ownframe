# Play Store

Android `BindMobile` is a 1.0 native host. A Play Store listing is not.
`scripts/android-release.sh` builds a signed arm64 APK for sideload. It does
not upload anything. Telegram and eyes are the exhibits, not store products.

## Remaining gaps

**TalkBack.** There is no TalkBack node. `Page.AccessibilityTree` is a Go
snapshot for host code. The window still calls the inner screen. That is not
an Android accessibility service.

**Process death.** Pause and resume keep in-process page state. If Android
kills the process, the page is gone. Apps that need restore must persist
their own state. OwnFrame has no savedInstanceState contract.

**IME.** Focusing a field opens `exp/textinput`. That is the phone keyboard
path in use, not a Play-supported IME contract. Preedit, rotation during
edit, and Back-to-dismiss need device checks on each app.

**GOWORK.** `ebitenmobile bind` copies packages into a temporary module.
The bind must run with `GOWORK` unset. `scripts/android-release.sh` unsets
it. Do not export `GOWORK` from the shell that runs the script.

**Signing.** The script reads `ANDROID_KEYSTORE`, `ANDROID_KEY_ALIAS`,
`ANDROID_STORE_PASS_FILE`, and `ANDROID_KEY_PASS_FILE`. Passwords live in
those files, never in the script or the repo. The default files are under
ignored `temp/android-release-signing/`. Play App Signing, a Play Console
listing, privacy policy, and store screenshots are still the publisher's
job.

Build and host setup is in [android-app-creation.md](android-app-creation.md).
The APK index is [examples/android-host/APPS.md](../examples/android-host/APPS.md).
