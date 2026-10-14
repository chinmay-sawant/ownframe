package window

// gestureState holds the shell's drag and key-chord tracking. It is embedded
// so game_state.go stays inside the file size limit; no caller sets these
// fields in a shell literal, so promotion keeps every existing reference.
type gestureState struct {
	gesture       pointerDrag
	touchCanceled bool
	fingers       touchGesture
	tabEaten      bool
	f11Eaten      bool
	clicks        clickWatch
	dragX, dragY  float64
}
