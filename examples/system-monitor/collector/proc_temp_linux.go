//go:build linux

package collector

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// milliScale converts millidegree sensor readings to degrees.
const milliScale = 1000

// readTemps reads hwmon temperature inputs in millidegrees. A sensor that
// disappears between the glob and the read is skipped, and an unreadable
// hwmon leaves every temperature unavailable.
func readTemps(ctx context.Context) []domain.Temp {
	files, err := filepath.Glob("/sys/class/hwmon/hwmon*/temp*_input")
	if err != nil || len(files) == 0 {
		return nil
	}

	out := make([]domain.Temp, 0, len(files))

	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return out
		}

		raw := readFile(file)
		if raw == nil {
			continue
		}

		milli, err := strconv.ParseInt(strings.TrimSpace(string(raw)), 10, 64)
		if err != nil {
			continue
		}

		out = append(out, domain.Temp{
			Name:    tempName(file),
			Celsius: domain.Celsius(float64(milli) / milliScale),
		})
	}

	return out
}

// tempName builds "hwmon/label" from the hwmon name and the sensor label.
func tempName(file string) string {
	dir := filepath.Dir(file)

	name := strings.TrimSpace(string(readFile(dir + "/name")))
	if name == "" {
		name = filepath.Base(dir)
	}

	label := strings.TrimSpace(string(readFile(strings.TrimSuffix(file, "_input") + "_label")))
	if label != "" {
		name += "/" + label
	}

	return name
}
