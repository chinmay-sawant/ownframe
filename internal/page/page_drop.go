package page

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// Drop is one file or directory dropped on the window.
type Drop = host.Drop

// maxPageDropFiles caps one drop. maxPageDropBytes caps one file's bytes,
// matching the window's stream limit for callers that skip the window.
const maxPageDropFiles = 32
const maxPageDropBytes = 16 << 20

// Drop calls the drop handler with the files dropped on the window and draws
// the page again. An error returns before the redraw.
func (p *Page) Drop(ctx context.Context, files []Drop) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if len(files) > maxPageDropFiles {
		files = files[:maxPageDropFiles]
	}
	for i := range files {
		files[i] = limitDrop(files[i])
	}
	if p.handlers.Drop != nil {
		if err := p.handlers.Drop(ctx, files); err != nil {
			return err
		}
	}
	return p.Redraw(ctx)
}

func limitDrop(d Drop) Drop {
	if d.IsDir || d.Read == nil {
		return d
	}
	next := d.Read
	name := d.Name
	d.Read = func() ([]byte, error) {
		data, err := next()
		if err != nil {
			return nil, err
		}
		if int64(len(data)) > maxPageDropBytes {
			return nil, fmt.Errorf("drop: %s exceeds %d bytes", name, maxPageDropBytes)
		}
		return data, nil
	}
	return d
}

var _ host.Dropper = (*Page)(nil)
