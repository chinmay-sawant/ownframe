package page

import (
	"html/template"
	"os"
)

// applySource swaps the template or the theme. A parse error keeps the last
// good one and stores the bytes as pending.
func (p *Page) applySource(kind watchKind, data []byte, info os.FileInfo) (bool, error) {
	wp := p.watch.at(kind)

	var err error
	if kind == watchHTML {
		err = p.setTemplate(data)
	} else {
		err = p.SetTheme(string(data))
	}

	if err != nil {
		p.watch.pending = &pendingEdit{kind: kind, data: data, info: info}

		return false, p.sourceErr(wp.path, err)
	}

	p.watch.pending = nil
	wp.accept(data, info)

	return true, nil
}

// setTemplate parses and installs the source, and points the current history
// entry at it. The history length and index do not move.
func (p *Page) setTemplate(data []byte) error {
	tpl, err := template.New("page").Parse(string(data))
	if err != nil {
		return err
	}

	p.tpl = tpl
	p.nav.Retarget(string(data))
	p.invalidateAccessibilityDocument()

	return nil
}
