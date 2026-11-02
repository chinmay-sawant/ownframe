package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// SkeletonOptions describes a loading placeholder. Width and Height are CSS
// pixels; zero Width fills the parent and zero Height uses 12px.
type SkeletonOptions struct {
	Width   float64
	Height  float64
	Rounded bool
}

// Skeleton is a pulsing placeholder for content that is still loading.
func Skeleton(options SkeletonOptions) *ui.Element {
	name := "of-skeleton"
	if options.Rounded {
		name = classes(name, "of-skeleton--rounded")
	}
	element := ui.Div().Class(name)
	if options.Width > 0 {
		element.Width(ui.Px(options.Width))
	}
	if options.Height > 0 {
		element.Height(ui.Px(options.Height))
	}
	return element
}
