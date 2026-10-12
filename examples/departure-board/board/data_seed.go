package board

// rowCount is the number of departures on screen.
const rowCount = 8

// dep is one seeded departure. The names avoid umlauts so tiles stay ASCII.
type dep struct {
	time, dest, plat string
	status           string
}

// seed is the live feed: sixteen departures ordered by time, first eight on
// the board. Times are HH:MM and platforms 1..12.
var seed = []dep{
	{"05:10", "Frankfurt Hbf", "3", "Boarding"},
	{"05:22", "Berlin Hbf", "7", "On time"},
	{"05:35", "Hamburg Altona", "2", "Delayed 5"},
	{"05:48", "Koeln Hbf", "9", "On time"},
	{"06:02", "Stuttgart Hbf", "4", "Delayed 10"},
	{"06:15", "Muenchen Ost", "11", "On time"},
	{"06:28", "Leipzig Hbf", "1", "On time"},
	{"06:41", "Dresden Neustadt", "6", "On time"},
	{"06:55", "Bremen Hbf", "5", "Departed"},
	{"07:08", "Nuernberg Hbf", "12", "On time"},
	{"07:22", "Hannover Hbf", "3", "Boarding"},
	{"07:36", "Dortmund Hbf", "8", "Delayed 5"},
	{"07:50", "Essen Hbf", "2", "On time"},
	{"08:04", "Freiburg Hbf", "10", "Delayed 10"},
	{"08:18", "Kiel Hbf", "7", "On time"},
	{"08:33", "Bonn Hbf", "1", "Departed"},
}

// statusCycle is the order one row's status walks, a step at a time. No two
// neighbours differ by fewer than six characters.
var statusCycle = [...]string{"On time", "Boarding", "Delayed 5", "Departed", "Delayed 10"}

// statusAt returns row i's status at feed. One or two rows step forward each
// feed and the rest hold, so every rotation changes a few tiles only.
func statusAt(i, feed int) string {
	step := (feed+i)/len(statusCycle) - i/len(statusCycle)
	k := (cycleIndex(seed[i].status) + step) % len(statusCycle)

	return statusCycle[k]
}

// cycleIndex maps a seed status to its place in the status cycle.
func cycleIndex(status string) int {
	for i, name := range statusCycle {
		if name == status {
			return i
		}
	}

	return 0
}
