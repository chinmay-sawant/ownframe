package window

// gestureState holds the shell's drag and key-chord tracking. The shell embeds
// this state; the game loop owns its fields.
type gestureState struct {
	gesture       pointerDrag
	touchCanceled bool
	fingers       touchGesture
	tabEaten      bool
	f11Eaten      bool
	clicks        clickWatch
	dragX, dragY  float64
}
