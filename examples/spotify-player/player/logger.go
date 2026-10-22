package player

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var logDir = filepath.Join("temp", "logdump", "dump")

// openSpotifyLog creates a timestamped JSON-line dump for manual GUI analysis.
func openSpotifyLog() *os.File {
	if err := os.MkdirAll(logDir, logDirPerm); err != nil {
		return nil
	}

	name := fmt.Sprintf("spotify-%s.ndjson", time.Now().Format("20060102-150405"))
	file, err := os.Create(filepath.Join(logDir, name)) //nolint:gosec // reason: log file is a timestamped name in the fixed logDir.
	if err != nil {
		return nil
	}
	return file
}

func (a *App) closeLog() {
	if a.logger != nil {
		_ = a.logger.Close()
	}
}
