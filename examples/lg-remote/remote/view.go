package remote

// Key is one tappable control.
type Key struct {
	ID       string
	Label    string
	Class    string
	Name     string
	Icon     bool
	Disabled bool
}

// View is the data the template prints.
type View struct {
	FontSize         int
	PressedID        string
	Sensitivity      int
	SensitivityLabel string
	SensitivityFill  int
	Title            string
	Status           string
	Mode             string
	ThemeLabel       string
	Hint             string
	ShowBluetooth    bool
	Panel            string
	Prev             string
	KeyboardNote     string
	KeysNav          []Row
	Layout           string
	Accent           string
	Accents          []AccentOption
	Comfort          bool
	Dark             bool
	PowerOn          bool
	Face             []Key
	More             []Key
	RemoteRows       []Row
	NumberRows       []Row
	MediaRows        []Row
	PadKeys          []Key
}

// AccentOption is one swatch in Settings.
type AccentOption struct {
	ID, Label, Colour string
	On                bool
}

type Row struct {
	Class string
	Keys  []Key
}
