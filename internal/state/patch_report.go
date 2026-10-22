package state

import (
	"sort"
	"time"
)

// PatchWithSkipped applies p, returning applied and skipped paths.
// Skipped holds Changed entries missing from Values or failing setPath.
// len(skipped) is the skipped count. One lock covers the origin write,
// all origin reads, and the batch; notify happens after unlock in order.
func (s *Store) PatchWithSkipped(p Patch) ([]Change, []string) {
	if s == nil {
		return nil, nil
	}
	if p.IsWhole() {
		if p.Values != nil {
			if v, ok := p.Values[""]; ok {
				s.SetData(v)
			}
		}
		return nil, nil
	}
	keys := make([]string, len(p.Changed))
	copy(keys, p.Changed)
	sort.Strings(keys)
	s.mu.Lock()
	s.origin = p.Origin
	out := make([]Change, 0, len(keys))
	var skip []string
	for _, path := range keys {
		v, ok := p.Values[path]
		if !ok {
			skip = append(skip, path)
			continue
		}
		c := Change{Path: path, Old: valueAtPath(s.data, path),
			New: v, Origin: s.origin, At: time.Now(),
			Properties: []string{path}, Changed: []string{path},
			Props: []string{path}}
		nd, ok := setPath(s.data, path, v)
		if !ok {
			skip = append(skip, path)
			continue
		}
		s.data = nd
		s.changes = append(s.changes, c)
		s.undo = append(s.undo, c)
		out = append(out, c)
	}
	if len(out) > 0 {
		s.redo = nil
	}
	ls := s.listenersLocked()
	s.mu.Unlock()
	for _, c := range out {
		for _, l := range ls {
			l(c)
		}
	}
	return out, skip
}
