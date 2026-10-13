package runtimebehavior

// DevtoolsCounter tracks patched vs repainted operations.
type DevtoolsCounter struct {
	Patched      int
	Repainted    int
	PatchedOps   int
	RepaintedOps int
}

// Add updates counters.
func (d *DevtoolsCounter) Add(patched, repainted int) {
	if d == nil {
		return
	}
	d.Patched += patched
	d.Repainted += repainted
	d.PatchedOps += patched
	d.RepaintedOps += repainted
}

// Record is an alias for Add.
func (d *DevtoolsCounter) Record(patched, repainted int) {
	d.Add(patched, repainted)
}

// Reset clears counters.
func (d *DevtoolsCounter) Reset() {
	if d == nil {
		return
	}
	d.Patched = 0
	d.Repainted = 0
	d.PatchedOps = 0
	d.RepaintedOps = 0
}

// Total returns patched plus repainted.
func (d *DevtoolsCounter) Total() int {
	if d == nil {
		return 0
	}
	return d.Patched + d.Repainted
}

// PatchedPercent returns the patched share as a percentage.
func (d *DevtoolsCounter) PatchedPercent() float64 {
	if d == nil {
		return 0
	}
	total := d.Total()
	if total == 0 {
		return 0
	}
	return float64(d.Patched) / float64(total) * 100
}
