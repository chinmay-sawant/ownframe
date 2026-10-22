package transfer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// openPartial opens or creates the partial file at the plan's offset.
func openPartial(path string, p plan) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), transferDirPerm); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPartialUnwritable, err)
	}

	flags := os.O_CREATE | os.O_WRONLY
	if p.truncate {
		flags |= os.O_TRUNC
	}

	file, err := os.OpenFile(path, flags, transferFilePerm) //nolint:gosec // reason: partial path is the download temp file, local file.
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPartialUnwritable, err)
	}

	if _, err := file.Seek(p.start, io.SeekStart); err != nil {
		_ = file.Close()

		return nil, err
	}

	return file, nil
}

// complete checks the checksum, renames the partial into place, and
// reports the outcome. Every handle must already be closed. A checksum
// mismatch removes the corrupt partial so a retry starts clean.
func (h *HTTP) complete(req Request, p plan, out Outcome) (Outcome, error) {
	if err := CheckChecksum(req.Partial, req.Checksum); err != nil {
		_ = os.Remove(req.Partial)

		return Outcome{}, err
	}

	if err := Finalize(req.Partial, req.Dest, FinalizeOptions{Sync: true}); err != nil {
		return Outcome{}, err
	}

	size, err := statSize(req.Dest)
	if err != nil {
		return Outcome{}, err
	}

	out.Path = req.Dest
	out.Bytes = size
	out.Total = p.total
	out.Validators = p.validators

	return out, nil
}
