package wire

import (
	"errors"
	"fmt"
	"log"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/store"
)

// pickDir returns the storage directory or "" when none is available.
func (b *Backend) pickDir(override string) string {
	if override != "" {
		return override
	}

	dir, err := store.DefaultDir()
	if err != nil {
		return ""
	}

	return dir
}

// openDir opens and locks one on-disk store. ErrLocked stops a second
// instance; any other failure falls back to memory.
func (b *Backend) openDir(dir string) error {
	st, err := store.Open(dir)
	if err != nil {
		log.Printf("download-manager: open failed: %s", cleanLog(err.Error()))

		return nil
	}

	release, err := st.Lock()
	if err != nil {
		_ = st.Close()

		if errors.Is(err, store.ErrLocked) {
			return fmt.Errorf("download-manager: another instance is using %s", cleanLog(dir))
		}

		log.Printf("download-manager: lock failed: %s", cleanLog(err.Error()))

		return nil
	}

	b.store, b.dir, b.release = st, dir, release

	return nil
}
