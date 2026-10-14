package player

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var logDir = filepath.Join("temp", "logdump", "dump")

// openSpotifyLog creates a timestamped JSON-line dump for manual GUI analysis.
func openSpotifyLog() *os.File {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil
	}

	name := fmt.Sprintf("spotify-%s.ndjson", time.Now().Format("20060102-150405"))
	file, err := os.Create(filepath.Join(logDir, name))
	if err != nil {
		return nil
	}
	return file
}

// logEvent writes one JSON line with the current page patch report and view.
func (a *App) logEvent(event, action string) {
	if a.logger == nil || a.page == nil {
		return
	}

	patch := a.page.PatchReport()
	entry := map[string]any{
		"time":   time.Now().Format(time.RFC3339Nano),
		"event":  event,
		"action": action,
		"page": map[string]any{
			"changed": patch.Changed,
			"added":   patch.Added,
			"removed": patch.Removed,
			"regions": patch.Regions,
			"full":    patch.Full,
		},
		"view": map[string]any{
			"playing":  a.view.Playing,
			"liked":    a.view.Liked,
			"shuffle":  a.view.Shuffle,
			"repeat":   a.view.Repeat,
			"progress": a.view.Progress,
			"volume":   a.view.Volume,
			"nav":      a.view.Nav,
		},
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	_, _ = a.logger.WriteString(string(data) + "\n")
}

func (a *App) closeLog() {
	if a.logger != nil {
		_ = a.logger.Close()
	}
}
