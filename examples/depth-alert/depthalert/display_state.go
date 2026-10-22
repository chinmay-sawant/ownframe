package depthalert

import "github.com/chinmay-sawant/ownframe/examples/depth-alert/alert"

func (d *displayFilter) setState(next alert.State) bool {
	if next == d.state {
		d.pending, d.pendingN = next, 0
		return false
	}
	if severity(next) > severity(d.state) {
		d.state, d.pending, d.pendingN = next, next, 0
		return true
	}
	if next != d.pending {
		d.pending, d.pendingN = next, 1
	} else {
		d.pendingN++
	}
	if d.pendingN < settleFrames {
		return false
	}
	d.state, d.pendingN = next, 0
	return true
}

func severity(state alert.State) int {
	switch state {
	case alert.StopNow:
		return severityStopNow
	case alert.Stop:
		return severityStop
	case alert.Unknown:
		return severityUnknown
	default:
		return 0
	}
}
