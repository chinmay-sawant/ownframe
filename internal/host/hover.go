package host

// Hoverer is a screen that reports the element it paints as hovered. The
// window reads it to tell its own cursor hover from one the page set, so a
// page that moves its hover itself keeps it until the pointer moves.
type Hoverer interface {
	HoverID() string
}
