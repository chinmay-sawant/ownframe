# Phase 4: Native Accessibility & Assistive Technology

> **Parent:** [README.md](README.md)  
> **Area:** Accessibility (a11y), Screen Readers, Operating System Assistive Tech  
> **Status:** Open `[ ]`  

---

## Overview

OwnFrame generates an abstract accessibility tree via `page.Access()` ([internal/page/a11y](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/page/a11y)). Native operating system assistive technology coverage remains incomplete:
- **Linux:** [internal/atspi](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/atspi) maps the snapshot to AT-SPI and serves it over the accessibility bus. Orca and Accerciser can read the tree; AT-initiated actions are not wired.
- **Windows:** Windows UI Automation (UIA) provider interfaces (`IRawElementProviderSimple`, `IInvokeProvider`, `ITextProvider`) are not implemented, leaving NVDA, JAWS, and Windows Narrator blind to the canvas.
- **macOS:** Cocoa `NSAccessibilityElement` protocols are not hooked up, preventing VoiceOver navigation.
- **Android:** The custom `EbitenView` does not expose an `AccessibilityNodeProvider`, preventing TalkBack from focusing or announcing screen content.

For commercial software subject to legal accessibility standards (e.g. ADA Title III, Section 508, EN 301 549), full OS screen reader integration is mandatory.

---

## Strategy & Principles

1. **Leverage the Semantic Tree:** Map directly from OwnFrame’s existing HTML-derived accessibility tree (`Role`, `Name`, `Value`, `Bounds`, `States`, `Actions`) rather than creating a second parallel hierarchy.
2. **Incremental Tree Mutations:** Emit focused, granular change events (e.g. element added/removed, text changed, focus moved) when DOM/element patches occur, avoiding full accessibility tree invalidations.
3. **Screen-Space Geometry:** Ensure all accessible node bounds are transformed to absolute screen coordinates so magnifiers and assistive highlights track accurately.

---

## Checklist

### 4.1 Accessibility Protocol & Mutation Events

- [ ] `A11Y-01` Define the live accessibility event bus in `internal/host/a11y`:
  - Focus changed (`Focus(nodeID)`)
  - Value / text changed (`ValueChanged(nodeID, newValue)`)
  - Bounds changed (`BoundsChanged(nodeID, newRect)`)
  - Subtree structure changed (`ChildrenChanged(parentID, added, removed)`)
  *Acceptance Criteria:* `internal/page/a11y` dispatches incremental mutation events on form edits, navigation, and element patches.

- [ ] `A11Y-02` Map HTML elements and ARIA attributes to standard accessibility roles and states:
  - Form controls: `<button>` (`Button`), `<input type="text">` (`EditableText`), `<input type="checkbox">` (`CheckBox`), `<select>` (`ComboBox`).
  - Structural: `<h1>`-`<h6>` (`Heading`), `<nav>` (`Navigation`), `<table>` (`Table`), `<ul>`/`<ol>` (`List`).
  - States: `focused`, `disabled`, `checked`, `expanded`, `readonly`, `hidden`.
  *Acceptance Criteria:* `internal/page/a11y` tests verify complete role and state derivation from HTML templates.

### 4.2 Linux Desktop: Live AT-SPI D-Bus Client

- [ ] `A11Y-03` Implement the live AT-SPI D-Bus client in `internal/atspi` connecting to the session bus (`org.a11y.Bus` via `DBUS_SESSION_BUS_ADDRESS`). Register the application with the AT-SPI registry daemon (`org.a11y.atspi.Registry`).  
  *Acceptance Criteria:* Linux desktop `Run` connects to the AT-SPI registry without crashing; `at-spi2-core` tools (`accerciser`) inspect OwnFrame's window.

- [ ] `A11Y-04` Implement core AT-SPI D-Bus object interfaces for accessible nodes:
  - `org.a11y.atspi.Accessible`: `GetRole`, `GetName`, `GetDescription`, `GetParent`, `GetChildren`.
  - `org.a11y.atspi.Component`: `GetExtents`, `Contains`, `GetGrabFocus`.
  - `org.a11y.atspi.Action`: `DoAction`, `GetNActions`, `GetActionName`.
  - `org.a11y.atspi.Text`: `GetText`, `GetCaretOffset`, `GetSelection`.
  *Acceptance Criteria:* Orca screen reader on Linux successfully speaks button labels, reads text input fields, and navigates via Tab.

### 4.3 Windows Desktop: UI Automation (UIA)

- [ ] `A11Y-05` Implement Windows UI Automation COM provider for the window handle (`HWND`):
  - `IRawElementProviderFragmentRoot` on the main window.
  - `IRawElementProviderSimple` and `IRawElementProviderFragment` for each accessible node.
  - Coordinate transformations from OwnFrame canvas pixels to Windows desktop screen pixels.
  *Acceptance Criteria:* Windows Accessibility Insights (`AccessibilityInsights.exe`) inspects and validates the element tree hierarchy.

- [ ] `A11Y-06` Implement standard UIA control patterns:
  - `IInvokeProvider` for buttons and clickable links.
  - `IValueProvider` / `ITextProvider` for form inputs.
  - `IToggleProvider` for checkboxes.
  - Dispatch `UiaRaiseAutomationEvent` on focus and value changes.
  *Acceptance Criteria:* NVDA and Windows Narrator accurately announce focused elements and speak typed characters in `examples/login`.

### 4.4 macOS Desktop: NSAccessibility

- [ ] `A11Y-07` In `internal/window/a11y_darwin.go`: Implement the `NSAccessibilityElement` protocol on the NSView wrapper:
  - Return accessibility children matching OwnFrame nodes.
  - Implement `accessibilityRole`, `accessibilityLabel`, `accessibilityValue`, and `accessibilityFrame`.
  - Post notifications (`NSAccessibilityFocusedUIElementChangedNotification`, `NSAccessibilityValueChangedNotification`).
  *Acceptance Criteria:* macOS VoiceOver highlights controls with the black focus ring and speaks content accurately.

### 4.5 Android: AccessibilityNodeProvider

- [ ] `A11Y-08` In `examples/android-host` and `internal/window/a11y_android.go`: Subclass `EbitenView` to implement `getAccessibilityNodeProvider()` returning a custom `AccessibilityNodeProviderCompat`:
  - `createAccessibilityNodeInfo(virtualViewId)` returning bounding bounds, class name, text, and supported actions (`ACTION_CLICK`, `ACTION_FOCUS`).
  - `performAction(virtualViewId, action, arguments)` routing taps and focus back to `host.Screen`.
  - Call `sendAccessibilityEvent(TYPE_VIEW_FOCUSED / TYPE_VIEW_CLICKED)` on state changes.
  *Acceptance Criteria:* Android TalkBack on Pixel/emulator navigates controls via swipe gestures, announces labels, and activates buttons via double-tap.
