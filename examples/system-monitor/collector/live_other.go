//go:build !linux && !windows

package collector

import (
	"context"
	"runtime"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// liveSource reports that this platform has no collector. The UI labels the
// sections unavailable instead of showing zeroes.
func liveSource() domain.Source {
	return unsupported{reason: runtime.GOOS + " has no live collector in this build"}
}

// unsupported stands in for a platform with no live collector.
type unsupported struct{ reason string }

func (u unsupported) Name() string { return "unsupported" }

func (u unsupported) Capabilities() domain.Capabilities {
	note := u.reason
	if note == "" {
		note = "no live collector on this platform"
	}

	return domain.Capabilities{Notes: []string{note}}
}

func (unsupported) Sample(ctx context.Context) (domain.Sample, error) {
	return domain.Sample{}, domain.ErrUnsupported
}

func (unsupported) Processes(ctx context.Context) ([]domain.Process, error) {
	return nil, domain.ErrUnsupported
}

func (unsupported) Detail(ctx context.Context, id domain.ProcessIdentity) (domain.ProcessDetail, error) {
	return domain.ProcessDetail{}, domain.ErrUnsupported
}
