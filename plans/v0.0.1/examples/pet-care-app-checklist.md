# Pet care app checklist

> **Parent:** Ownframe examples, v0.0.1.
> **Status:** First implementation delivered. Pixel 7 startup and sample-data flow verified; Windows acceptance and remaining Android interaction checks remain open.
> **Branch:** `examples/pet-care-doc`.
> **Recorded:** 2026-10-11.

## Goal

Build one small pet-care application that works in the native Windows window
and as an Android app. The Go application, screens, data rules, and navigation
are shared. Platform adapters handle startup and native lifecycle details.
The app stores its records locally on each device. Cross-device sync is out
of scope for this first version.

The app opens on a demo sign-in screen with the demo credentials filled in.
The user must still submit valid credentials before any app screen appears.
This gate demonstrates the sign-in flow. It is not account security and must
not imply that the sample credentials protect private data.

The first version supports multiple pets and user-entered care records. It can
show due and overdue items from dates the user entered. It does not provide
medical advice, medication dosing, or background operating-system alerts.

## Checklist rules

- `[ ]` means open.
- `[x]` requires implementation and evidence recorded in this file.
- `[~]` means a documented partial result or a device check that cannot yet
  be completed. Name the remaining evidence beside the row.
- Keep every row open until its specific proof exists. A build does not prove
  touch behavior, and an emulator does not prove physical-device behavior.
- Keep screenshots, APKs, traces, and device logs under ignored `temp/pet-care/`.
- This is the new canonical checklist. Do not change other plans, indexes, or
  ledgers as part of the planning phase.

## Implementation state

The shared app lives under `examples/pet-care`. It has the sign-in gate, pet
profiles, care records, schedule labels, settings, SQLite storage, and JSON
copy/paste backup. The Android Activity uses `examples/android-host` for
insets, Back dispatch, and lifecycle. Two transparent illustrations are
embedded from `examples/pet-care/assets/`.

The interaction smoke flow signed in, created a pet, and rendered its profile.
The Android app now starts from the Activity with its private files directory,
then creates the shared app and binds the Ebiten view. This avoids the Android
startup panic caused by `os.UserConfigDir()` having neither `HOME` nor
`XDG_CONFIG_HOME`. A first-run seed inserts three example pets and seven care
records once for an empty database. The Pixel 7 installed and launched the APK,
passed the sign-in gate, and rendered the sample pets and care states. Device
screenshots and a pulled SQLite inspection are under ignored `temp/pet-care/`.

`make build` and `make test` passed after the Go source changes. The final
Android debug APK built and was installed at `temp/pet-care/pet-care-debug.apk`.
The desktop shell uses the available window width, the login panel and Add Pet
form are centered, and shared Material-style fields and action spacing were
checked in desktop and Android renders. The shared stylesheet now uses rounded
filled controls and removes square focus and hover outlines across routes.
Settings cards also add consistent space around buttons and the backup input.
The Pixel 7 light-theme login screenshot verifies the rounded fields and
button; a dedicated dark Settings screenshot is still open. The native Windows
window has not been launched by the agent. No commit was created.

## Current source baseline

- `examples/telegram/telegram` holds shared app behavior; its `mobile` package
  binds that same Go app, and `examples/telegram/android` supplies Android
  activity integration. Use this as the starting app pattern.
- `examples/android-host` is the shared Android module for insets, Back,
  configuration changes, and lifecycle. Keep app routes and data rules in Go.
- Android page pinch can be disabled with `Config.DisableTouchZoom` or
  `Page.SetTouchZoomAllowed(false)`. Pinch policy is independent of scrolling.
- `Config.Theme` and `Page.SetTheme` apply a stylesheet after the template's
  styles. `SetTheme` needs a redraw; keep the two palettes on the same geometry.
- Android has no OS file picker through an HTML file input. Backup import and
  export need an app-owned Android document-picker bridge; do not treat a typed
  filename as a working backup flow.

## Product contract

### Screens

The first version has these working screens:

1. **Sign in.** Demo email and password are prefilled. Submit is required.
   Invalid values show an inline error and keep the user on this screen.
2. **Dashboard.** Show each pet, the next due care item, recent activity,
   and direct actions to add a pet or record care.
3. **Pets.** List, search, add, edit, and remove pet profiles.
4. **Pet profile.** Show the selected pet's details, upcoming care, and
   recent records, with links to edit the profile and add a record.
5. **Care schedule.** List dated care items as upcoming, due today, overdue,
   or complete. Dates use the device's local calendar day.
6. **Care history.** Add, edit, filter, and remove dated records such as
   feeding, grooming, walks, vet visits, and user-entered medication notes.
7. **Settings.** Switch light and dark themes, export a backup, restore a
   backup with confirmation, and sign out.

All screens use the same Go state and action handlers on Windows and Android.
The Android Activity must not duplicate screen navigation or data rules.

### Data and safety

- A pet has a stable ID, name, species, optional breed and birth date, notes,
  and an optional image.
- A care record has a stable ID, pet ID, category, local date/time, optional
  note, and completion state where the category uses a schedule.
- Reject blank names, invalid dates, unknown pet IDs, malformed imports, and
  duplicate IDs. Confirm destructive actions and identify the affected pet.
- Editing or deleting a pet must not silently orphan its records. Offer a
  clear cancel or delete-with-records choice.
- Medication entries store user-provided facts only. Do not calculate doses,
  recommend treatment, or present due dates as medical advice.
- Persist desktop data in SQLite under
  `os.UserConfigDir()/ownframe/pet-care/`. Android stores it under the app's
  private files directory at `ownframe/pet-care/pet-care.db`.
  If storage cannot open, explain that changes are temporary before allowing
  the user to continue.
- Keep a versioned JSON backup format. Reject an invalid backup without
  replacing the current database. Confirm before replacing existing data.

### Shared and platform behavior

| Area | Windows | Android |
|---|---|---|
| Application | `ownframe.Run` opens the shared app. | `BindMobile` opens the same shared app through the existing mobile package pattern. |
| Native host | Resizable window; preserve the current screen and data through resize and minimize/restore. | Reuse `examples/android-host` for system bars, Back, configuration, and lifecycle behavior. |
| Layout | Dashboard can use the available width; forms and pet details remain readable in a narrow window. | Single-column layout, reachable controls, safe-area handling, and keyboard-aware forms. |
| Zoom | Keep ordinary desktop keyboard zoom behavior. | Disable page pinch zoom for both pinch-out and pinch-in on every screen. Keep vertical scrolling enabled. |
| Touch | Mouse and keyboard remain usable. | Important actions have final hit areas of at least 48 dp; scrolling does not activate a row or button. |
| Rotation/resize | Reflow without losing the selected pet, typed draft, or scroll position. | Follow the shared host's rotation handling; preserve page state and focused draft through portrait/landscape changes and IME resize. |
| Storage | Local to the Windows user profile. | Local to the Android app. No network account or sync is implied. |

Android page zoom is disabled through the existing page policy, for example:

```go
page.SetTouchZoomAllowed(false)
```

Verify the policy after every route change and after Activity pause/resume. A
two-finger gesture must not change page scale, text size, hit boxes, or scroll
position. One-finger vertical scrolling must still work. The app has no custom
pinch interaction in this release.

## Visual direction

Use a quiet, minimal product style: warm off-white page background, white or
near-white cards, dark green ink, a soft green accent, muted supporting text,
and clear status labels that do not depend on color alone. Keep geometry in
the template. The theme changes paint tokens so switching themes does not
move controls.

The template should use custom properties with fallbacks. `Config.Theme` or
`Page.SetTheme` supplies the selected palette:

```css
:root {
  --page: #f4f6f2;
  --card: #ffffff;
  --ink: #202a23;
  --muted: #68756b;
  --line: #dce4dc;
  --accent: #477453;
  --accent-soft: #e8f0e7;
  --warning: #9a6127;
  --danger: #a8433e;
}

body {
  color: var(--ink, #202a23);
  background: var(--page, #f4f6f2);
  font-family: Arial, sans-serif;
}

.card {
  color: var(--ink, #202a23);
  background: var(--card, #ffffff);
  border: 1px solid var(--line, #dce4dc);
  border-radius: 16px;
  padding: 16px;
}

.primary-action {
  color: #ffffff;
  background: var(--accent, #477453);
  border-radius: 12px;
  min-height: 48px;
}
```

The dark palette changes the same variables, not layout values:

```css
:root {
  --page: #171c19;
  --card: #222a24;
  --ink: #edf2ed;
  --muted: #b3beb5;
  --line: #39443b;
  --accent: #8fbd96;
  --accent-soft: #2d3b30;
  --warning: #e6b66c;
  --danger: #ef9288;
}
```

Theme selection should call the page API and redraw after a successful change:

```go
if err := page.SetTheme(themeCSS); err != nil {
    return err
}
return page.Redraw(ctx)
```

Check every proposed CSS property and selector against the current Blinkless
compatibility docs before relying on it. Keep status meaning visible in text
or an icon as well as color. Avoid transitions and CSS animation.

Generate separate project assets with the imagegen skill during Phase 1. Start
with a transparent dog mascot/profile illustration and a separate empty-state
illustration. If a wagging idle animation survives the quality and performance
checks, use separate consistent frames and retain a static fallback. Keep
functional icons in one coherent vector icon set rather than rasterizing each
button. Store final assets under `examples/pet-care/assets/` and test their
crop, contrast, and legibility at phone and desktop sizes.

## Phase 0: confirm the baseline and settle the app contract

- [x] P0.01 Compared Telegram's mobile binding and Android Activity with the
  current shared host. The app keeps screen rules in Go and delegates Android
  insets, Back, and lifecycle to `examples/android-host`.
- [x] P0.02 Read the app, mobile, forms, theme, and Android host patterns.
  Implementation points are `examples/pet-care/app`, `examples/pet-care/mobile`,
  and `examples/pet-care/android`.
- [x] P0.03 `examples/pet-care/README.md` and the sign-in copy identify the
  bundled demo credentials and state that the gate does not protect data.
- [x] P0.04 The seven screens are sign in, dashboard, pets, pet profile,
  schedule, history, and settings. The UI makes no dosage, medical advice, or
  background alert claims.
- [x] P0.05 Both platforms use the same SQLite schema and version 1 JSON
  backup model. Data stays on each device; backup currently transfers through
  copy and paste.

Phase exit: the app boundaries, data lifecycle, and host ownership are explicit
before implementation starts.

## Phase 1: build the responsive visual shell and assets

- [x] P1.01 Added a shared app shell, navigation, responsive single-column
  phone layout, form rows, cards, status labels, and action buttons in
  `examples/pet-care/app/`. Used the Blinkless compatibility matrix for the
  CSS declarations and selectors; the width media query still needs visual
  confirmation at the listed sizes.
- [~] P1.02 Light and dark custom property sheets are wired through
  `Page.SetTheme`. Visual geometry and form-state retention still need a
  rendered comparison.
- [x] P1.03 Added separate transparent dog profile and empty-care PNGs under
  `examples/pet-care/assets/`, embedded by the app with no text in the art.
  Both PNG alpha channels contain transparent and opaque pixels.
- [x] P1.04 No animation was selected. The profile and empty-state images are
  static, so no animation loop or fallback frame is needed.
- [~] P1.05 Image crop, contrast, and wrapping have not been checked at the
  listed viewport sizes. Record screenshots or measurements after visual QA.
- [~] P1.06 CSS uses the available desktop width and centers the login and Add
  Pet form. Desktop screenshots and a Pixel 7 login render were inspected.
  Rounded input and button edges were verified on the Pixel 7 after removing
  square outline strokes. Native Windows maximized-window acceptance remains
  with the user; dark Settings visual confirmation is open.

Phase exit: the same shell is readable and stable at phone and desktop sizes;
all assets are in the repository and no essential text is baked into images.

## Phase 2: implement the data model, storage, and sign-in gate

- [~] P2.01 Added typed pet, care record, and version 1 backup models with
  random stable IDs and local `YYYY-MM-DD` dates. Care records do not yet store
  time of day.
- [~] P2.02 SQLite creation, `user_version=1`, desktop default config path,
  Android private files path, `-db` override, close, temporary in-memory
  warning, and an idempotent first-run sample seed are implemented. Upgrade
  migrations and failure/restart cases still need deterministic coverage.
- [~] P2.03 Pet and care record create/read/update/delete, filtering, and
  schedule completion are implemented. Repository boundary tests are open.
- [~] P2.04 JSON export and restore validate version, IDs, references, dates,
  and size before transactional replacement. Android uses the system document
  picker; Windows uses copy/paste. Malformed-input tests remain open.
- [x] P2.05 Fresh sessions show prefilled demo values and require the sign-in
  action. The headless interaction flow exercised successful submission.
- [~] P2.06 Invalid credentials render an inline error in the Go handler.
  Invalid-value, focus-retention, and Android IME behavior have not been run.
- [~] P2.07 Sign out returns to the login screen while database records remain
  in SQLite. Restart and second-instance persistence checks remain open.
- [ ] P2.08 Deterministic storage and login boundary tests have not been added.
  Required cases remain second instance, open failure, restart, malformed data,
  and repeated migration.

Phase exit: records persist correctly, backup failures cannot corrupt current
data, and no application screen is reachable before successful sign-in.

## Phase 3: make every screen useful

- [~] P3.01 Dashboard reads saved pets and shows upcoming care, recent
  activity, and an empty state. The Pixel 7 render showed three seeded pets and
  care entries. Per-pet next-item summaries and no-upcoming scenarios remain
  open.
- [~] P3.02 Pet search, add, and edit are implemented with blank and future
  birth-date validation. Long-name limits and flow tests remain open.
- [~] P3.03 Pet details use the bundled default dog illustration. User image
  selection, cancel behavior, and invalid-image fallback are not implemented.
- [~] P3.04 Care entry and filtering by care type or pet are implemented.
  Draft/focus retention across validation errors needs platform verification.
- [~] P3.05 Scheduled records show upcoming, due today, overdue, and complete
  labels from local calendar dates. Seeded examples cover each state, and the
  Pixel 7 schedule render showed the upcoming, due-today, and complete rows.
  Fixed-clock boundary tests and the no-care state remain open.
- [~] P3.06 Pet deletion asks for confirmation and states that care records
  will be deleted; SQLite cascades the dependent rows. Cancel and transaction
  behavior need repository tests.
- [~] P3.07 Theme, Android document-picker export/restore, Windows JSON
  copy/paste, validated restore confirmation, and sign out are wired. Picker
  cancellation and URI behavior still need device verification.
- [~] P3.08 Visible navigation and action buttons have Go handlers, including
  confirmations for pet and care deletion. Every route and feedback state has
  not been exercised.

Phase exit: every screen has real state and actions, and the primary workflows
work without placeholder-only controls.

## Phase 4: connect the native Windows and Android apps

- [~] P4.01 The desktop entry point calls `ownframe.Run`. Native Windows
  startup, resize, minimize/restore, close, and relaunch are for the user to
  check on Windows.
- [~] P4.02 Added the Android mobile package, Gradle project, shared-host
  Activity, app-owned document picker bridge, and app-private storage startup.
  The Pixel 7 installed and launched the debug APK and opened the sign-in and
  seeded app screens. Picker cancellation, URI access, Activity recreation,
  and backup round trips need device verification.
- [~] P4.03 `DisableTouchZoom` is enabled in the shared page config. Pinch
  gestures and scale invariance have not been checked on a device.
- [~] P4.04 One-finger scroll, taps, and long-press behavior have not been
  exercised on Android.
- [~] P4.05 Shared host handles IME-first Back dispatch and Go handles app
  Back. Form/detail ordering needs device checks; the login gate has not been
  verified through Android Back.
- [~] P4.06 Android host uses its shared inset padding policy. Cutout, bars,
  IME, font scale, and smallest viewport remain unverified on a device.
- [~] P4.07 Rotation, draft retention, focus, and scroll-state behavior were
  not tested on Android.
- [~] P4.08 CSS buttons and fields use a 48 px minimum height. Final dp targets
  after device scaling have not been measured.
- [ ] P4.09 Windows mouse, keyboard focus, Enter, Escape, and text editing are
  awaiting the user's native Windows run.

Phase exit: the same workflows operate on Windows and Android, with the
platform-specific input and lifecycle policy tested on each host.

## Phase 5: complete acceptance and handoff

- [ ] P5.01 App-level tests for login gating, routes, CRUD, schedule boundaries,
  backup restore, confirmations, and restart retention have not been added.
- [x] P5.02 `make build` and `make test` both passed on the final Go source tree
  on 2026-10-11.
- [x] P5.03 Android debug APK built at
  `temp/pet-care/pet-care-debug.apk`, installed on Pixel 7 serial
  `2C161FDH200GBQ`, and launched through the sign-in screen. The startup log
  contains no Go panic or Android fatal exception. See
  `temp/pet-care/pixel-rounded-final.png` for the latest light-theme render.
- [~] P5.04 Native Windows run is awaiting the user. Record OS version,
  display scale, viewport, input, local storage, and relaunch results here.
  Verify that the full-width shell uses the right side of a maximized window.
- [~] P5.05 The headless flow used the prefilled demo credentials and required
  a click to enter the app. Pixel 7 also passed the prefilled sign-in and
  displayed the seeded dashboard. Invalid credentials and Windows remain
  unverified.
- [~] P5.06 The generated PNGs are transparent and embedded. Phone/desktop
  crop, absent/corrupt image handling, and fallback rendering remain open.
- [x] P5.07 `examples/pet-care/README.md` and `android/README.md` document the
  screens, run/build/install steps, storage path, demo-gate limits, versioned
  backup format, and device checks that remain unverified.
- [ ] P5.08 Add the example to `examples/readme.md` only after the app works
  and the user authorizes that index update.

Phase exit: the example works through its own UI on Windows and Android, and
the checklist records which platform and device evidence exists.

## Deferred

- Cross-device sync, accounts, cloud storage, remote services, and real
  authentication.
- Background notifications and OS alarms.
- Medical advice, diagnosis, drug lookup, or dose calculation.
- Multiple user profiles, sharing, and permissions.
- App-store publication or Windows installer distribution.
