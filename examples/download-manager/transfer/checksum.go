package transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// HashFile returns the hex sha256 of path.
func HashFile(path string) (string, error) {
	f, err := os.Open(path) //nolint:gosec // reason: checksum reads the download dest it just wrote, local file.
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	sum := sha256.New()
	if _, err := io.Copy(sum, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(sum.Sum(nil)), nil
}

// CheckChecksum verifies a "sha256:<hex>" checksum against path. An empty
// want always passes.
func CheckChecksum(path, want string) error {
	if want == "" {
		return nil
	}

	algo, hexWant, ok := strings.Cut(want, ":")
	if !ok || algo != "sha256" {
		return fmt.Errorf("%w: unsupported checksum %q", ErrChecksum, want)
	}

	got, err := HashFile(path)
	if err != nil {
		return err
	}

	if !strings.EqualFold(got, hexWant) {
		return fmt.Errorf("%w: sha256 %s != %s", ErrChecksum, got, hexWant)
	}

	return nil
}
