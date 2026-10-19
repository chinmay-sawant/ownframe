package depthalert

import "github.com/chinmay-sawant/ownframe/examples/depth-alert/alert"

const displaySamples = 5

type displayFilter struct {
	near, cover [displaySamples]int32
	next, size  int
	missing     int
	frames      int
	nearView    int32
	coverView   int32
	state       alert.State
	pending     alert.State
	pendingN    int
}

func (d *displayFilter) reset() { *d = displayFilter{} }

func (d *displayFilter) observe(mm, cover int32, tracking bool, state alert.State) bool {
	if !tracking || cover < alert.MinimumCoveragePercent || mm <= 0 {
		d.missing++
		if d.missing >= 3 {
			d.clearMeasurements()
			return true
		}
		return false
	}
	d.missing = 0
	changed := d.setState(state)
	d.near[d.next], d.cover[d.next] = mm, cover
	d.next = (d.next + 1) % displaySamples
	if d.size < displaySamples {
		d.size++
	}
	d.nearView = settle(d.nearView, median(d.near[:], d.size), 200)
	d.coverView = settle(d.coverView, median(d.cover[:], d.size), 8)
	d.frames++
	if changed || d.frames >= 3 {
		d.frames = 0
		return true
	}
	return false
}

func (d *displayFilter) unavailable() { d.reset() }

func (d *displayFilter) clearMeasurements() {
	d.near, d.cover = [displaySamples]int32{}, [displaySamples]int32{}
	d.next, d.size = 0, 0
	d.nearView, d.coverView, d.state = 0, 0, alert.Unknown
	d.frames, d.pending, d.pendingN = 0, alert.Unknown, 0
}
