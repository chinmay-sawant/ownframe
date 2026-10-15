package page

import (
	"fmt"
	"os"
)

// sourceHTML returns the template source and the FileInfo of the file it
// came from, when there is one. Setting HTML and File together is an error.
func sourceHTML(cfg Config) (string, os.FileInfo, error) {
	if cfg.File == "" {
		return cfg.HTML, nil, nil
	}

	if cfg.HTML != "" {
		return "", nil, ErrBadSource
	}

	data, info, err := readSource(cfg.File)
	if err != nil {
		return "", nil, err
	}

	return string(data), info, nil
}

// sourceTheme returns the theme source and the FileInfo of its file.
// Setting Theme and ThemeFile together is an error.
func sourceTheme(cfg Config) (string, os.FileInfo, error) {
	if cfg.ThemeFile == "" {
		return cfg.Theme, nil, nil
	}

	if cfg.Theme != "" {
		return "", nil, ErrBadSource
	}

	data, info, err := readSource(cfg.ThemeFile)
	if err != nil {
		return "", nil, err
	}

	return string(data), info, nil
}

// readSource reads one watched file. A missing or unreadable file is
// ErrBadSource with the path and the reason.
func readSource(path string) ([]byte, os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s: %w", ErrBadSource, path, err)
	}

	if info.Size() > maxSourceSize {
		return nil, info, sourceSizeErr(path, info.Size())
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s: %w", ErrBadSource, path, err)
	}

	if int64(len(data)) > maxSourceSize {
		return nil, info, sourceSizeErr(path, int64(len(data)))
	}

	if fresh, err := os.Stat(path); err == nil {
		info = fresh
	}

	return data, info, nil
}

func sourceSizeErr(path string, size int64) error {
	return fmt.Errorf("%w: %w: %s: source too large (%d bytes, limit %d)", ErrBadSource, ErrSourceTooLarge, path, size, maxSourceSize)
}
