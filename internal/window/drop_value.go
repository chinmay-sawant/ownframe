package window

import (
	"fmt"
	"io"
	"io/fs"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// maxDropFiles caps one drop, so a bulk drop cannot load memory and CPU.
// maxDropBytes caps one file's Read with a stream, not ReadFile.
const maxDropFiles = 32
const maxDropBytes = 16 << 20

// dropValue describes one entry. Read reads through fsys, so it works only
// while the frame that owns fsys is alive. Read streams with a byte limit
// and fails past maxDropBytes instead of loading a huge file.
func dropValue(fsys fs.FS, entry fs.DirEntry) (host.Drop, error) {
	info, err := entry.Info()
	if err != nil {
		return host.Drop{}, err
	}
	name := entry.Name()
	d := host.Drop{
		Name: name, Size: info.Size(), IsDir: info.IsDir(),
		Read: func() ([]byte, error) {
			f, err := fsys.Open(name)
			if err != nil {
				return nil, err
			}
			data, rerr := io.ReadAll(io.LimitReader(f, maxDropBytes+1))
			cerr := f.Close()
			if rerr != nil {
				return nil, rerr
			}
			if cerr != nil {
				return nil, cerr
			}
			if int64(len(data)) > maxDropBytes {
				return nil, fmt.Errorf("drop: %s exceeds %d bytes", name, maxDropBytes)
			}
			return data, nil
		},
	}
	if p, ok := entry.(interface{ AbsPath() string }); ok {
		d.Path = p.AbsPath()
	}
	return d, nil
}
