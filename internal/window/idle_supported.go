//go:build !js && !android && !ios

package window

// idlePaceSupported reports that the desktop backend blocks in the event
// loop when idle, so a scheduled frame or an input event wakes the next
// frame. Browsers and phones must keep their frame callback, so pacing
// stays off there and runWindow passes the plain game to Ebiten.
const idlePaceSupported = true
