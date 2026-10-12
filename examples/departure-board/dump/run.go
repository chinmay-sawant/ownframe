// Package dump measures departure board scenarios headless.
package dump

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"github.com/chinmay-sawant/ownframe/examples/departure-board/board"
)

// RunFlag parses the -scenario list and calls Run.
func RunFlag(list, out, cpuPath string) error {
	fields := strings.FieldsFunc(list, func(r rune) bool { return r == ',' || r == ' ' })
	return Run(fields, out, cpuPath)
}

// Run measures scenarios (all when empty), writes <out>/<name>.json, profiles
// to cpuPath when set, and prints a row per scenario.
func Run(scenarios []string, out, cpuPath string) error {
	if len(scenarios) == 0 {
		scenarios = Scenarios()
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	if cpuPath != "" {
		f, err := os.Create(cpuPath)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			return err
		}
		defer pprof.StopCPUProfile()
	}
	fmt.Printf("%-8s %6s %6s %7s %7s %7s  %-34s %9s %9s %9s\n",
		"scenario", "frames", "ops", "p50ms", "p95ms", "maxms", "kinds",
		"rMean", "rMax", "oMean")
	for _, name := range scenarios {
		app, err := board.New()
		if err != nil {
			return err
		}
		rep, err := Measure(context.Background(), app, name)
		if err != nil {
			return err
		}
		body, err := json.MarshalIndent(rep, "", "  ")
		if err != nil {
			return err
		}
		path := filepath.Join(out, name+".json")
		if err := os.WriteFile(path, append(body, '\n'), 0644); err != nil {
			return err
		}
		kinds := strings.Trim(fmt.Sprint(rep.Kinds), "map[]")
		ops, n := 0, max(len(rep.Frames), 1)
		for _, f := range rep.Frames {
			ops += f.ReplayedOps
		}
		fmt.Printf("%-8s %6d %6d %7.2f %7.2f %7.2f  %-34s %9d %9d %9.1f\n",
			rep.Scenario, len(rep.Frames), rep.Ops, float64(rep.TickP50NS)/1e6,
			float64(rep.TickP95NS)/1e6, float64(rep.TickMaxNS)/1e6, kinds,
			rep.RectMeanArea, rep.RectMaxArea, float64(ops)/float64(n))
	}
	return nil
}
