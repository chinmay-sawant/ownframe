# Pet care app checklist

> **Parent:** Ownframe examples, v0.0.1.
> **Status:** Planning. No application code or image assets have been created.
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

## Planning state

The branch `examples/pet-care-doc` was created from `master`. The worktree
contains this new checklist only. This pass did not create app code, generate
images, or run builds or tests.

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
- Persist data in SQLite under `os.UserConfigDir()/ownframe/pet-care/`.
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

- [ ] P0.01 Compare the current Android and app patterns with the source
  baseline above before coding. Proof: note any source changes that affect
  ownership or platform behavior.
- [ ] P0.02 Read the current app, mobile, forms, theme, and Android runtime
  patterns before coding. Proof: implementation notes identify the exact
  shared packages and Android host integration points.
- [ ] P0.03 Confirm the demo sign-in is a UI gate, not security. Proof: the
  README and login copy state the demo credentials are bundled and no account
  authentication or private-data protection is provided.
- [ ] P0.04 Confirm the seven-screen scope and the record-only medical policy.
  Proof: screen map and data model agree; no background alert or dosage
  feature is claimed.
- [ ] P0.05 Confirm local-only storage and the first-release backup contract.
  Proof: Windows and Android use the same versioned data format, with no
  hidden sync dependency.

Phase exit: the app boundaries, data lifecycle, and host ownership are explicit
before implementation starts.

## Phase 1: build the responsive visual shell and assets

- [ ] P1.01 Create the shared app shell, navigation, responsive page regions,
  and reusable cards, form rows, status labels, and action buttons.
- [ ] P1.02 Implement the light and dark token sheets shown above. Verify the
  theme switch changes colors without changing geometry or losing form state.
- [ ] P1.03 Create the dog mascot and empty-state images as separate assets
  with transparent backgrounds where appropriate. Keep a stable illustration
  style, no embedded text, and no watermark.
- [ ] P1.04 If animation is selected, create consistent separate dog frames and
  implement a low-frequency, bounded idle animation with a static fallback.
  It must stop when the app is paused and must not relayout every frame.
- [ ] P1.05 Check contrast, text wrapping, and image crop at 360x640, 390x844,
  a narrow split-screen viewport, and a 1280x800 desktop window.
- [ ] P1.06 Check that the mobile layout reflows into one column and that the
  desktop layout uses wider space without stretching form controls.

Phase exit: the same shell is readable and stable at phone and desktop sizes;
all assets are in the repository and no essential text is baked into images.

## Phase 2: implement the data model, storage, and sign-in gate

- [ ] P2.01 Define typed pet, care record, schedule, and backup models with
  stable IDs and explicit local date/time handling.
- [ ] P2.02 Implement SQLite schema creation and migration, default storage
  path, explicit data-directory override, close behavior, and temporary-mode
  error state.
- [ ] P2.03 Implement repository operations for pet and care record create,
  read, update, delete, filtering, and schedule completion.
- [ ] P2.04 Implement versioned JSON export and restore. Validate the full
  backup before changing stored data; handle empty, malformed, old-version,
  duplicate-ID, missing-pet, and oversized files without partial replacement.
- [ ] P2.05 Start every fresh app session at the sign-in screen with the demo
  email and password prefilled. Require an explicit submit before navigation.
- [ ] P2.06 Reject empty, altered, or invalid credentials with an inline error;
  keep focus and typed values correct through redraw and Android IME updates.
- [ ] P2.07 Sign out to the login screen and clear transient screen state while
  preserving saved pet data.
- [ ] P2.08 Cover storage and login boundaries with deterministic tests,
  including a second app instance, database-open failure, restart persistence,
  malformed data, and repeated migration.

Phase exit: records persist correctly, backup failures cannot corrupt current
data, and no application screen is reachable before successful sign-in.

## Phase 3: make every screen useful

- [ ] P3.01 Build the dashboard from stored pets and schedule records. Empty
  state, one pet, many pets, no upcoming items, and overdue items all have
  useful copy and reachable actions.
- [ ] P3.02 Implement pet list search and add/edit flows with validation for
  blank names, long names, optional fields, and invalid or future birth dates.
- [ ] P3.03 Implement pet details and image selection. Canceling image choice
  keeps the previous image; missing or invalid image bytes show the fallback.
- [ ] P3.04 Implement care record entry and history filters. Preserve the
  selected pet, date, category, and text draft across redraws and validation
  errors.
- [ ] P3.05 Implement schedule states for upcoming, due today, overdue,
  completed, and no scheduled care. Test midnight/date-boundary behavior using
  a fixed clock and the device's local calendar day.
- [ ] P3.06 Confirm deleting a pet explains which records will be deleted.
  Cancel leaves all rows intact; confirm removes the pet and its dependent
  records in one transaction.
- [ ] P3.07 Add settings for theme, backup, restore, and sign out. Confirm
  restore and destructive actions; cancel leaves the current data untouched.
- [ ] P3.08 Verify every visible button, tab, card action, and empty-state
  action has a working handler and clear feedback.

Phase exit: every screen has real state and actions, and the primary workflows
work without placeholder-only controls.

## Phase 4: connect the native Windows and Android apps

- [ ] P4.01 Wire the shared app package to `ownframe.Run` for Windows and
  verify startup, window close, minimize/restore, resize, and a second launch.
- [ ] P4.02 Add the Android mobile package and Gradle project using the current
  `examples/telegram` structure; include `examples/android-host` instead of
  copying its lifecycle, Back, or inset behavior. Add an app-owned Android
  document-picker adapter for backup import/export and handle cancellation,
  URI access, and Activity recreation.
- [ ] P4.03 Set the page pinch policy to disabled for the Android app. Test
  pinch-out and pinch-in on the login, dashboard, pet list, profile, schedule,
  history, and settings screens. Page scale, text scale, hit boxes, and
  navigation must remain unchanged.
- [ ] P4.04 Verify one-finger scrolling remains enabled and does not trigger a
  row action; taps and long presses on buttons and list rows do not begin a
  scroll accidentally.
- [ ] P4.05 Verify Android Back order: dismiss the IME first, then close an
  open form/detail or return to the pet list, then return to the dashboard;
  it must not bypass the login gate.
- [ ] P4.06 Verify safe areas, status/navigation bars, cutouts, keyboard
  insets, font scaling, and smallest supported viewport. Do not apply Android
  density conversion a second time in Go.
- [ ] P4.07 Rotate portrait to both landscape orientations while signed in,
  editing a field, showing the IME, viewing a pet, and scrolling history.
  Preserve data, route, draft, selected pet, and valid focus; do not leave a
  stuck pointer, stale keyboard, or stale scroll velocity.
- [ ] P4.08 Confirm all essential touch targets remain at least 48 dp after
  responsive layout and device scaling. Increase spacing or reflow instead of
  shrinking the whole page to fit.
- [ ] P4.09 Verify desktop mouse, keyboard focus order, Enter submission,
  Escape behavior, and text editing without changing Android-specific rules.

Phase exit: the same workflows operate on Windows and Android, with the
platform-specific input and lifecycle policy tested on each host.

## Phase 5: complete acceptance and handoff

- [ ] P5.01 Run app-level tests for login gating, all screen routes, CRUD,
  schedule boundaries, backup restore, destructive confirmations, and state
  retention after redraw/restart.
- [ ] P5.02 Run `make test` from the repository root and `make build`; record
  the results against the final source tree.
- [ ] P5.03 Build the Android debug APK, install it on a named physical device,
  and exercise all seven screens, pinch-in/out, rotation, IME, Back, pause and
  resume, and process restart.
- [ ] P5.04 Run the Windows app on a native Windows machine. Record OS version,
  display scale, window sizes, and results for resize, minimize/restore,
  keyboard, mouse, local storage, and relaunch.
- [ ] P5.05 Verify the first-run demo credentials and required submit on both
  platforms; verify invalid credentials cannot open app routes.
- [ ] P5.06 Check all images at phone density and desktop size, including
  absent image, transparent edges, corrupt data, and fallback rendering.
- [ ] P5.07 Update the new example's README with setup, Windows run, Android
  build/install, storage, demo-gate limits, backup format, supported screens,
  and explicitly unverified device boundaries.
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
