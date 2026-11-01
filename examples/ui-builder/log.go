package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	builderLogDirPerm  = 0o750
	builderLogFilePerm = 0o600
)

type builderLogOptions struct {
	enabled bool
	path    string
}

// openBuilderLog opens the ui-builder diagnostic log for append. It returns a
// logger that writes to both stderr and the file, the open file for the
// caller to close, and the file path for the startup line. When options.path
// is empty the path is os.UserCacheDir()/ownframe/ui-builder/ui-builder.log,
// falling back to os.TempDir(). A failure to create the file still returns a
// stderr-only logger with a nil file, so logging never breaks the demo. When
// disabled it returns a nil logger without creating a file.
func openBuilderLog(options builderLogOptions) (*log.Logger, *os.File, string) {
	if !options.enabled {
		return nil, nil, ""
	}
	path := options.path
	if path == "" {
		dir, err := os.UserCacheDir()
		if err != nil || dir == "" {
			path = filepath.Join(os.TempDir(), "ownframe-ui-builder", "ui-builder.log")
		} else {
			path = filepath.Join(dir, "ownframe", "ui-builder", "ui-builder.log")
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), builderLogDirPerm); err != nil {
		return log.New(os.Stderr, "ui-builder ", log.LstdFlags|log.Lmicroseconds), nil, path
	}
	//nolint:gosec // reason: log path is a fixed cache-dir name or an explicit -log flag.
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, builderLogFilePerm)
	if err != nil {
		return log.New(os.Stderr, "ui-builder ", log.LstdFlags|log.Lmicroseconds), nil, path
	}
	writer := io.MultiWriter(os.Stderr, file)
	return log.New(writer, "ui-builder ", log.LstdFlags|log.Lmicroseconds), file, path
}

func builderLogf(logger *log.Logger, format string, args ...any) {
	if logger == nil {
		return
	}
	logger.Printf(format, args...)
}

// logStartup records one line with the time, mode, and log path so a later
// crash can be matched to its run.
func logStartup(logger *log.Logger, path, mode, addr string) {
	builderLogf(logger, "startup time=%s mode=%s addr=%s log=%s",
		time.Now().Format(time.RFC3339), mode, addr, path)
}
