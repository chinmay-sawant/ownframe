package viewtransition

import (
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

type node struct {
	index int
	box   ownframe.Box
	boxes []ownframe.Box
}

func validRoot(d *ownframe.Display, i int) bool {
	return i == DocumentRoot || i >= 0 && i < len(d.Boxes)
}
func rect(b ownframe.Box) Rect { return Rect{b.X, b.Y, b.W, b.H} }
func mode(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return "normal"
	}
	return s
}
func classNames(s string) []string {
	var out []string
	for _, v := range strings.Fields(s) {
		if v != "none" {
			out = append(out, v)
		}
	}
	return out
}

func discover(d *ownframe.Display, root int) (map[string]node, []string, error) {
	items := make(map[string]node)
	var names []string
	for i, b := range d.Boxes {
		if !inside(d.Boxes, i, root) || blocked(d.Boxes, i, root) {
			continue
		}
		name := strings.TrimSpace(b.ViewTransitionName)
		if name == "" || name == "none" {
			continue
		}
		if _, ok := items[name]; ok {
			return nil, nil, ErrDuplicateName
		}
		items[name] = node{i, b, d.Boxes}
		names = append(names, name)
	}
	return items, names, nil
}

func inside(boxes []ownframe.Box, i, root int) bool {
	if root == DocumentRoot {
		return true
	}
	for steps := 0; i >= 0 && i < len(boxes) && steps < len(boxes); steps++ {
		if i == root {
			return true
		}
		i = boxes[i].ParentIndex
	}
	return false
}

func blocked(boxes []ownframe.Box, i, root int) bool {
	for steps := 0; i >= 0 && i < len(boxes) && steps < len(boxes); steps++ {
		if i != root && mode(boxes[i].ViewTransitionScope) == "all" {
			return true
		}
		if i == root {
			return false
		}
		i = boxes[i].ParentIndex
	}
	return false
}
