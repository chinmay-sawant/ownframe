package collector

import (
	"context"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// Processes reads the fixture process table. It advances the simulation too,
// so CPU times keep moving between summary samples.
func (d *Dummy) Processes(ctx context.Context) ([]domain.Process, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.advance(d.opts.Stamp())

	return d.procRows(), nil
}

// Detail reads one fixture process in full. A process that exited since the
// table was read returns ErrGone.
func (d *Dummy) Detail(ctx context.Context, id domain.ProcessIdentity) (domain.ProcessDetail, error) {
	if err := ctx.Err(); err != nil {
		return domain.ProcessDetail{}, err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	p := d.findProc(id)
	if p == nil {
		return domain.ProcessDetail{}, domain.ErrGone
	}

	d.evolveProc(p)

	return d.detailOf(p), nil
}

// detailOf builds the full detail for one process.
func (d *Dummy) detailOf(p *dmProc) domain.ProcessDetail {
	return domain.ProcessDetail{
		Process:    d.rowOf(p),
		Command:    p.name + " --user " + p.user,
		Exe:        "/usr/bin/" + p.name,
		Cwd:        "/home/" + p.user,
		Parent:     domain.ProcessIdentity{PID: 1, Start: 1},
		ParentName: "init",
		Virtual:    domain.Bytes(float64(p.mem * virtMult)),
		ReadBytes:  domain.Bytes(float64(p.read)),
		WriteBytes: domain.Bytes(float64(p.write)),
	}
}
