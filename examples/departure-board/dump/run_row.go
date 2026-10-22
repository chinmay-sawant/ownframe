package dump

import (
	"fmt"
	"strings"
)

// printHeader prints the table header once per Run. rMean and rMax stay the
// bounding box areas; rcts and aMean follow the damage list.
func printHeader() {
	fmt.Printf("%-8s %6s %6s %7s %7s %7s  %-34s %9s %9s %9s %5s %9s\n",
		"scenario", "frames", "ops", "p50ms", "p95ms", "maxms", "kinds",
		"rMean", "rMax", "oMean", "rcts", "aMean")
}

// printRow prints one scenario row. oMean is the mean replayed ops per frame.
// rcts and aMean average the frames that reported a damage list: the list
// length and the summed area.
func printRow(rep Report) {
	kinds := strings.Trim(fmt.Sprint(rep.Kinds), "map[]")
	ops, n := 0, max(len(rep.Frames), 1)
	for _, f := range rep.Frames {
		ops += f.ReplayedOps
	}
	fmt.Printf("%-8s %6d %6d %7.2f %7.2f %7.2f  %-34s %9d %9d %9.1f %5.1f %9d\n",
		rep.Scenario, len(rep.Frames), rep.Ops, float64(rep.TickP50NS)/nanosPerMilli,
		float64(rep.TickP95NS)/nanosPerMilli, float64(rep.TickMaxNS)/nanosPerMilli, kinds,
		rep.RectMeanArea, rep.RectMaxArea, float64(ops)/float64(n),
		rep.RectsMean, rep.AreaMean)
}
