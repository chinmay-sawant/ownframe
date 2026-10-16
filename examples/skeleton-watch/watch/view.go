package watch

// View models the template and display data for the skeleton watch.
type View struct {
	Lume             bool
	LumeClass        string
	ClockText        string
	ChronoText       string
	ChronoRunning    bool
	PowerHours       float64
	PowerPercent     int
	HourAngle        float64
	MinuteAngle      float64
	SecondAngle      float64
	ChronoSecAngle   float64
	ChronoMinAngle   float64
	ChronoHourAngle  float64
	TourbillonAngle  float64
	BalanceAngle     float64
	ReserveAngle     float64
	CenterWheelAngle float64
	ThirdWheelAngle  float64
	FourthWheelAngle float64
	BarrelAngle      float64
	Status           string
}

func defaultView() View {
	return View{
		LumeClass:    "lume-off",
		ClockText:    "10:08:42",
		ChronoText:   "00:00.00",
		PowerHours:   48.0,
		PowerPercent: 66,
		ReserveAngle: 20.0,
		Status:       "RUNNING (28,800 vph)",
	}
}
