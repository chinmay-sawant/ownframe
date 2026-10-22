package collector

import (
	"math/rand"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// sampleTemps builds the three fixture sensors; two stay deliberately
// unreadable.
func sampleTemps(rng *rand.Rand) []domain.Temp {
	return []domain.Temp{
		{Name: "coretemp/package", Celsius: domain.Celsius(tempBase + rng.Float64()*tempSpan)},
		{Name: "gpu/nvidia", Celsius: domain.Value{}},
		{Name: "acpi/fan1", Celsius: domain.Value{}},
	}
}
