package transfer

import (
	"errors"
	"os"
	"time"
)

// FinalizeOptions controls the partial-to-final rename.
type FinalizeOptions struct {
	// Retries is the rename retry count. Zero means five, the Windows
	// sharing-violation budget.
	Retries int
	// Backoff is the pause between retries. Zero means 30 ms.
	Backoff time.Duration
	// Sync flushes the finished file before the rename.
	Sync bool
}

// retries and backoff fill zero values.
func (o FinalizeOptions) retries() int {
	if o.Retries <= 0 {
		return finalizeRetries
	}

	return o.Retries
}

func (o FinalizeOptions) backoff() time.Duration {
	if o.Backoff <= 0 {
		return finalizeBackoff
	}

	return o.Backoff
}

// Finalize moves a closed partial file to dest. It never overwrites an
// existing dest and retries a failed rename, which on Windows covers a
// sharing violation or an antivirus lock. The caller must close every
// handle on partial first. After too many failures partial is left in place.
func Finalize(partial, dest string, opts FinalizeOptions) error {
	if _, err := os.Lstat(dest); err == nil {
		return ErrDestinationExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if opts.Sync {
		if err := syncFile(partial); err != nil {
			return err
		}
	}

	var last error

	for i := 0; i <= opts.retries(); i++ {
		if i > 0 {
			time.Sleep(opts.backoff())
		}

		if last = os.Rename(partial, dest); last == nil {
			return nil
		}
	}

	return last
}

// syncFile flushes one file's bytes to the device.
func syncFile(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0) //nolint:gosec // reason: partial path is the download temp file, local file.
	if err != nil {
		return err
	}

	err = f.Sync()
	closeErr := f.Close()

	if err != nil {
		return err
	}

	return closeErr
}
