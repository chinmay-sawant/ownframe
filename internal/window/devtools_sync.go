package window

import "github.com/chinmay-sawant/blinkless/layout"

// devSync follows the page's flag once per frame, so Config.DevTools and a
// SetDevTools call open or close the overlay without a key.
func (s *shell) devSync() error {
	insp := s.devInspector()
	if insp == nil {
		s.dev.on = false

		return nil
	}

	on := insp.DevTools()
	if on {
		// A page that opened the overlay once keeps F12 available after it
		// closes, including a close before the first frame.
		s.dev.available = true
	}
	if on == s.dev.on {
		return nil
	}

	s.dev.on = on
	if !on {
		return s.devClose()
	}

	return nil
}

// devOpsCache keeps the Ops tab rows for one page generation. The tab has one
// row per painted operation, so a still generation must not rebuild the list.
// The outline toggle changes a row without a new generation, so it is a key
// too. Generation zero never caches, so a fake screen in a test stays fresh.
type devOpsCache struct {
	rows    []devRow
	gen     uint64
	toggle  bool
	valid   bool
	display *layout.Display
}

// devOpsRows returns the Ops tab rows, rebuilt only when the page generation,
// the display, or the outline toggle moved.
func (s *shell) devOpsRows() []devRow {
	gen := s.devGeneration()
	c := &s.dev.opsCache

	if gen != 0 && c.valid && c.gen == gen && c.toggle == s.dev.ops && c.display == s.display {
		return c.rows
	}

	rows := s.devOpsBuildRows()
	if gen != 0 {
		*c = devOpsCache{rows: rows, gen: gen, toggle: s.dev.ops, valid: true, display: s.display}
	}

	return rows
}

// devGeneration is the page's build generation, zero before a screen exists.
func (s *shell) devGeneration() uint64 {
	if s.app == nil {
		return 0
	}

	return s.app.Generation()
}
