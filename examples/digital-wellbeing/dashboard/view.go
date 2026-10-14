package dashboard

// AppChoice is a launchable app returned by Android without usage totals.
type AppChoice struct {
	Package    string `json:"package"`
	Label      string `json:"label"`
	Letter     string `json:"letter"`
	Icon       string `json:"icon"`
	IconSource string `json:"-"`
}

// Day is one local calendar day's foreground time.
type Day struct {
	Label   string
	Minutes int
	Time    string
	Level   int
}

// View is the data shown by the template.
type View struct {
	Access       bool
	Selected     string
	SelectedName string
	Selecting    bool
	Apps         []AppChoice
	Days         []Day
	Today        string
	Week         string
	Average      string
	GoalMessage  string
	Screen       string
}
