package depthalert

import (
	"errors"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/frame"
)

type summaryOps struct {
	generation                   uint64
	signal, scan, near, coverage *ownframe.DisplayOp
}

func (s *summaryOps) bind(p *ownframe.Page) error {
	d := p.Display()
	if d == nil {
		return errors.New("depth alert: summary has no display list")
	}
	for _, b := range p.Boxes() {
		op := frame.Text(d, b)
		switch b.ID {
		case "signal":
			s.signal = op
		case "scan-state":
			s.scan = op
		case "distance":
			s.near = op
		case "coverage":
			s.coverage = op
		}
	}
	if s.signal == nil || s.scan == nil || s.near == nil || s.coverage == nil {
		return errors.New("depth alert: summary text operations are missing")
	}
	s.generation = p.Generation()
	return nil
}

func (s *summaryOps) update(p *ownframe.Page, v view) {
	s.set(p, "signal", s.signal, v.State)
	scan := "Scanning is paused"
	if v.Scanning {
		scan = "Scanning is active"
	}
	s.set(p, "scan-state", s.scan, scan)
	s.set(p, "distance", s.near, v.Distance)
	s.set(p, "coverage", s.coverage, v.Coverage)
}

func (s *summaryOps) set(p *ownframe.Page, id string, op *ownframe.DisplayOp, value string) {
	if op.Text != value {
		op.Text = value
		p.Invalidate(id)
	}
}
