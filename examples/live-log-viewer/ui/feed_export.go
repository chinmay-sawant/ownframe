package ui

import (
	"context"
	"os"
	"path/filepath"

	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/entry"
	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/store"
)

// Export writes the inclusive FromID..ToID range as text.
func (f *StoreFeed) Export(ctx context.Context, q Query, path string) (int, error) {
	cq, err := f.coreQuery(q)
	if err != nil {
		return 0, err
	}

	if q.FromID > 0 {
		cq.Cursor = entry.EntryID(q.FromID - 1)
	}

	if q.ToID > 0 {
		cq.MaxID = entry.EntryID(q.ToID)
	}

	ex, err := f.st.Export(ctx, store.ExportOptions{Query: cq, MaxEntries: q.Limit})
	if err != nil {
		return 0, err
	}

	if err := writeFile(path, ex.Data); err != nil {
		return 0, err
	}

	return ex.Entries, nil
}

// writeFile writes data through a temporary file and a rename, so a failure
// midway leaves no partial export behind.
func writeFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, uiExportPerm); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".live-log-ui-*")
	if err != nil {
		return err
	}

	name := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(name)

		return err
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(name)

		return err
	}

	if err := os.Rename(name, path); err != nil {
		_ = os.Remove(name)

		return err
	}

	return nil
}
