package host

// SelectOption is one choice of a select control for the web preview.
type SelectOption struct {
	Value    string
	Label    string
	Selected bool
	Disabled bool
}

// WebControl is one live control for the browser document.
// Value is the stored value, including a password. Never the painted bullets.
type WebControl struct {
	ID             string
	Tag            string
	Type           string
	Name           string
	Value          string
	Checked        bool
	Disabled       bool
	Focused        bool
	SelectionStart int
	SelectionEnd   int
	Options        []SelectOption
}

// WebDOM is an optional screen capability exposing executed HTML, theme CSS,
// and live form controls to the web preview.
type WebDOM interface {
	ExecutedSource() string
	ThemeSource() string
	WebControls() []WebControl
	SelectControls() map[string][]SelectOption
}
