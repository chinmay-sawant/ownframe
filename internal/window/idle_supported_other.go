//go:build js || android || ios

package window

// idlePaceSupported is false where the platform cannot block its frame
// callback: browsers and mobile views keep their platform frame callbacks.
const idlePaceSupported = false
