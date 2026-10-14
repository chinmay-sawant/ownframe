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
	Index    int
	Label    string
	Date     string
	Minutes  int
	Time     string
	Average  string
	Level    int
	Selected bool
	Apps     []AppUsage
}

// AppUsage is one app's foreground time within a day or week.
type AppUsage struct {
	Package    string `json:"package"`
	Label      string `json:"label"`
	Minutes    int    `json:"minutes"`
	Time       string `json:"-"`
	Percent    int    `json:"-"`
	Letter     string `json:"-"`
	IconSource string `json:"-"`
}

// View is the data shown by the template.
type View struct {
	Access         bool
	Selected       string
	SelectedName   string
	Selecting      bool
	Apps           []AppChoice
	Days           []Day
	TopApps        []AppUsage
	AllApps        []AppUsage
	SelectedDay    Day
	HasSelectedDay bool
	Today          string
	Week           string
	Average        string
	WeekRange      string
	WeekHeading    string
	WeekOffset     int
	HeroEyebrow    string
	HeroCaption    string
	InsightWeek    string
	InsightAverage string
	WeekdayTotal   string
	WeekendTotal   string
	TopApp         string
	TopAppTime     string
	CanPrevious    bool
	CanNext        bool
	GoalMessage    string
	Screen         string
}
