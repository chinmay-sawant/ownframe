package mobile

import "github.com/chinmay-sawant/ownframe/androidaccess"

// Accessibility returns the page's accessibility tree as JSON for the
// Android host's TalkBack provider. It is an empty tree on a desktop build.
func Accessibility() string {
	return androidaccess.Tree()
}

// AccessibilityPerform runs one action TalkBack asked for (1 focus, 2
// activate, 3 set value) on a node of the tree read with Accessibility. It
// returns "" on success and a short message when the request was refused.
func AccessibilityPerform(document int64, node string, action int, value string) string {
	return androidaccess.Perform(document, node, action, value)
}
