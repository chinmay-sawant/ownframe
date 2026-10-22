package transfer

import (
	"context"
	"os"
	"path/filepath"
)

// Download streams the deterministic body to req.Partial and finalizes it.
func (f *Fake) Download(ctx context.Context, req Request, report Reporter) (Outcome, error) {
	if err := os.MkdirAll(filepath.Dir(req.Dest), transferDirPerm); err != nil {
		return Outcome{}, err
	}

	size := FakeSize(req.JobID)
	total := size

	if fakeUnknown(req.JobID) {
		total = Unknown
	}

	offset, resumed, err := partialOffset(req.Partial, size)
	if err != nil {
		return Outcome{}, err
	}

	if err := f.stream(ctx, req, offset, size, total, report); err != nil {
		return Outcome{}, err
	}

	if err := CheckChecksum(req.Partial, req.Checksum); err != nil {
		_ = os.Remove(req.Partial)

		return Outcome{}, err
	}

	if err := Finalize(req.Partial, req.Dest, FinalizeOptions{Sync: true}); err != nil {
		return Outcome{}, err
	}

	return Outcome{
		Path:    req.Dest,
		Bytes:   size,
		Total:   total,
		Resumed: resumed,
	}, nil
}
