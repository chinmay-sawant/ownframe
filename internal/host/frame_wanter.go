package host

import "time"

// FrameWanter is a screen that can say when it next needs a frame. The
// desktop window reads NextFrameAt once per frame, before it draws, and
// schedules its idle wake for that time. A zero time means the screen does
// not ask: the window keeps its keep-alive fallback. A time at or before now
// asks for the next frame at once, which keeps a per-frame animation on the
// vsync loop. NextFrameAt may clear the request when it reads it, so a screen
// that wants more frames must ask again on every frame; Page does. A screen
// without this interface behaves exactly as before.
type FrameWanter interface {
	NextFrameAt() time.Time
}
