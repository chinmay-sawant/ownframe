package collector

import (
	"context"
	"sync"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// fakeSource is a controllable source for manager tests.
type fakeSource struct {
	mu      sync.Mutex
	samples uint64
	slow    time.Duration
	fail    bool
	gone    bool
}

func (f *fakeSource) Name() string { return "fake" }

// setFail toggles the sample failure mode.
func (f *fakeSource) setFail(v bool) {
	f.mu.Lock()
	f.fail = v
	f.mu.Unlock()
}

func (f *fakeSource) Capabilities() domain.Capabilities {
	return domain.Capabilities{Processes: true, ProcessDetail: true}
}

func (f *fakeSource) Sample(ctx context.Context) (domain.Sample, error) {
	f.mu.Lock()
	f.samples++
	slow, fail := f.slow, f.fail
	f.mu.Unlock()

	if slow > 0 {
		select {
		case <-time.After(slow):
		case <-ctx.Done():
			return domain.Sample{}, ctx.Err()
		}
	}
	if fail {
		return domain.Sample{}, context.DeadlineExceeded
	}

	return domain.Sample{
		CPUTotal: domain.CPUTimes{Busy: f.samples * 10, Total: f.samples * 100},
		Mem:      domain.Memory{Total: domain.Bytes(100), Available: domain.Bytes(40)},
		Disks:    []domain.Disk{{Device: "sda", ReadBytes: f.samples * 100, WriteBytes: f.samples * 50}},
	}, nil
}

func (f *fakeSource) Processes(ctx context.Context) ([]domain.Process, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return []domain.Process{
		{ID: domain.ProcessIdentity{PID: 10, Start: 1}, Name: "one", CPUTime: 100},
		{ID: domain.ProcessIdentity{PID: 11, Start: 2}, Name: "two", CPUTime: 50},
	}, nil
}

func (f *fakeSource) Detail(ctx context.Context, id domain.ProcessIdentity) (domain.ProcessDetail, error) {
	f.mu.Lock()
	gone := f.gone
	f.mu.Unlock()

	if gone || id.PID == 99 {
		return domain.ProcessDetail{}, domain.ErrGone
	}

	return domain.ProcessDetail{
		Process:   domain.Process{ID: id, Name: "one"},
		Command:   "one --flag",
		ReadBytes: domain.Bytes(1000),
	}, nil
}
