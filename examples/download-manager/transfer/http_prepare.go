package transfer

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

// prepared is the request plus the resume offset decision.
type prepared struct {
	request *http.Request
	offset  int64 // bytes already on disk when the request went out
	resumed bool
}

// prepare stats the partial file, decides whether a resume is provable, and
// builds the GET with Range and If-Range.
func prepare(ctx context.Context, req Request, ua string) (prepared, error) {
	offset, err := statSize(req.Partial)
	if err != nil {
		return prepared{}, fmt.Errorf("%w: %w", ErrPartialUnwritable, err)
	}

	resumed := offset > 0

	// Without a saved validator there is no way to prove the remote entity
	// is the one the partial came from, so restart instead of appending.
	if resumed && req.Validators.Empty() {
		if err := os.Truncate(req.Partial, 0); err != nil {
			return prepared{}, err
		}

		offset, resumed = 0, false
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, req.URL, nil)
	if err != nil {
		return prepared{}, err
	}

	// Asking for identity keeps Content-Length meaningful and stops the
	// transport from silently gunzipping a body we want at byte offsets.
	r.Header.Set("Accept-Encoding", "identity")

	if ua != "" {
		r.Header.Set("User-Agent", ua)
	}

	if offset > 0 {
		r.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
		r.Header.Set("If-Range", validatorHeader(req.Validators))
	}

	return prepared{request: r, offset: offset, resumed: resumed}, nil
}

// statSize returns a file size or zero when the file does not exist.
func statSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}

		return 0, err
	}

	return info.Size(), nil
}

// validatorHeader renders the strongest saved validator for If-Range.
func validatorHeader(v Validators) string {
	if v.ETag != "" {
		return v.ETag
	}

	return v.LastModified
}
