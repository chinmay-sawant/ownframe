package page

// ElementEdit is one change to an element in the retained tree: an attribute
// to set, or the text to replace. A nil Text and an empty Attribute change
// nothing.
//
// ID names the element the edit applies to. An empty ID means the element
// the patch is for, and an edit on a descendant rides along with the patch.
// Text replaces the text content and drops the children.
//
// Text set this way does not go through the emoji pass. The pass runs on the
// source before it is parsed, and a text edit lands on the parsed tree, so an
// emoji in patched text paints as text rather than as the color image the
// same emoji gets from the template. The engine's tree API only writes a text
// node, so the image cannot be built inside the tree.
type ElementEdit struct {
	ID        string
	Attribute string
	Value     string
	Text      *string
}
