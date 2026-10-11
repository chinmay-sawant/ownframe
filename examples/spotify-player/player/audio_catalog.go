package player

import (
	"embed"
	"encoding/json"
	"sync"
)

//go:embed assets/audio/catalog.json assets/audio/tracks/*.mp3 assets/audio/covers/*.jpg
var audioFiles embed.FS

type audioTrack struct {
	Title           string `json:"title"`
	File            string `json:"file"`
	Category        string `json:"category"`
	Cover           string `json:"cover"`
	CoverFile       string `json:"cover_file"`
	Genre           string `json:"genre"`
	Description     string `json:"description"`
	DurationSeconds int    `json:"duration_seconds"`
	Bytes           int    `json:"bytes"`
	SHA256          string `json:"sha256"`
	Creator         string `json:"-"`
	License         string `json:"-"`
	LicenseURL      string `json:"-"`
	Source          string `json:"-"`
}

type audioManifest struct {
	Creator    string       `json:"creator"`
	License    string       `json:"license"`
	LicenseURL string       `json:"license_url"`
	Source     string       `json:"source_url"`
	Tracks     []audioTrack `json:"tracks"`
}

var bundled []audioTrack
var loadBundled sync.Once

func bundledTracks() []audioTrack {
	loadBundled.Do(func() {
		data, err := audioFiles.ReadFile("assets/audio/catalog.json")
		if err != nil {
			panic("player: read bundled audio catalog: " + err.Error())
		}
		var manifest audioManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			panic("player: parse bundled audio catalog: " + err.Error())
		}
		bundled = manifest.Tracks
		for i := range bundled {
			bundled[i].Creator = manifest.Creator
			bundled[i].License = manifest.License
			bundled[i].LicenseURL = manifest.LicenseURL
			bundled[i].Source = manifest.Source
		}
	})
	return append([]audioTrack(nil), bundled...)
}
