//go:build linux

package collector

import (
	"errors"
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// TestProcSourceSmoke reads the real /proc on this host: counters, the own
// process identity, and a detail read. It skips when /proc is absent.
func TestProcSourceSmoke(t *testing.T) {
	if !pathExists("/proc/stat") {
		t.Skip("no /proc on this host")
	}

	s := newProcSource()

	smp, err := s.Sample(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if smp.CPUTotal.Total == 0 || len(smp.CPUCores) == 0 {
		t.Fatalf("cpu = %+v cores=%d", smp.CPUTotal, len(smp.CPUCores))
	}
	if smp.Uptime <= 0 {
		t.Fatalf("uptime = %v", smp.Uptime)
	}

	rows, err := s.Processes(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	self := os.Getpid()
	var id domain.ProcessIdentity

	for _, row := range rows {
		if int(row.ID.PID) == self {
			id = row.ID
		}
	}
	if id.Start == 0 {
		t.Fatal("own process missing or without a start value")
	}

	d, err := s.Detail(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}
	if d.Command == "" && d.Exe == "" {
		t.Fatal("detail has neither command nor executable path")
	}
	if d.Memory.Valid == false && d.Virtual.Valid == false {
		t.Fatal("detail has no memory values")
	}
}

// TestProcSourceGone checks that a stale start value reports ErrGone instead
// of another run's data.
func TestProcSourceGone(t *testing.T) {
	if !pathExists("/proc/self/stat") {
		t.Skip("no /proc on this host")
	}

	s := newProcSource()
	rows, err := s.Processes(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	self := os.Getpid()
	for _, row := range rows {
		if int(row.ID.PID) != self {
			continue
		}

		stale := row.ID
		stale.Start++

		if _, err := s.Detail(t.Context(), stale); !errors.Is(err, domain.ErrGone) {
			t.Fatalf("err = %v, want ErrGone", err)
		}

		return
	}

	t.Fatal("own process missing from the table")
}

// TestProcSourceCapabilities is in live_caps_test.go.
