package insights

// streakHistory is the full heatmap, one string per weekday, Sunday first,
// oldest week first. Each rune is a Cell level: 0 empty, 1 no dictation,
// 2..5 teal shades. The last 19 columns are the newest weeks on screen.
var streakHistory = []string{
	"0002200012000110000400321111101000000001131115555555",
	"0003300120000010003340212030024400000002111115555555",
	"0000000012000000003450110122001100000001211115225455",
	"0002100013000020000310003121000000000012111115535555",
	"0002100012000000003440101102040100000031111155555555",
	"0003100310000120005550310124010000000031111155555555",
	"0001000122000100002440101330100000000023112155555555",
}

// streakMonth is one month band over streakHistory, measured in week
// columns. The bands sum to the 52 columns.
type streakMonth struct {
	Label string
	Weeks int
}

// streakMonths are the month labels in history order.
var streakMonths = []streakMonth{
	{Label: "Oct", Weeks: monthWeeks}, {Label: "Nov", Weeks: monthWeeks},
	{Label: "Dec", Weeks: monthWeeks}, {Label: "Jan", Weeks: weeksLong},
	{Label: "Feb", Weeks: monthWeeks}, {Label: "Mar", Weeks: monthWeeks},
	{Label: "Apr", Weeks: monthWeeks}, {Label: "May", Weeks: monthWeeks},
	{Label: "Jun", Weeks: monthWeeks}, {Label: "Jul", Weeks: weeksShort},
	{Label: "Aug", Weeks: monthWeeks}, {Label: "Sep", Weeks: weeksPeak},
	{Label: "Oct", Weeks: monthWeeksTail},
}
