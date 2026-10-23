package a11y

// Control is the form snapshot the tree reads.
// Editable is the page typing rule. File inputs are editable and still
// do not advertise SetValue.
type Control struct {
	ID       string
	Tag      string
	Type     string
	Value    string
	Checked  bool
	Disabled bool
	Editable bool
}
