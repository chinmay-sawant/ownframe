package player

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/jpeg"
	"testing"
)

func TestCatalogTracksAndArtworkAreUniqueAndBundled(t *testing.T) {
	titles, files, hashes, covers := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, track := range bundledTracks() {
		if titles[track.Title] || files[track.File] {
			t.Fatalf("duplicate title or file: %q %q", track.Title, track.File)
		}
		titles[track.Title], files[track.File] = true, true
		data, err := audioFiles.ReadFile(track.File)
		if err != nil {
			t.Fatalf("read %s: %v", track.File, err)
		}
		hash := fmt.Sprintf("%x", sha256.Sum256(data))
		if hashes[hash] || hash != track.SHA256 {
			t.Fatalf("duplicate or wrong audio hash for %q", track.Title)
		}
		hashes[hash] = true
		covers[track.CoverFile] = true
	}
	if len(titles) != 30 {
		t.Fatalf("unique tracks = %d, want 30", len(titles))
	}
	if len(covers) != 30 {
		t.Fatalf("unique track covers = %d, want 30", len(covers))
	}
	entries, err := audioFiles.ReadDir("assets/audio/covers")
	if err != nil || len(entries) != 40 {
		t.Fatalf("bundled cover files = %d, want 40 (err %v)", len(entries), err)
	}
	coverHashes := map[string]bool{}
	for _, entry := range entries {
		cover := "assets/audio/covers/" + entry.Name()
		data, err := audioFiles.ReadFile(cover)
		if err != nil {
			t.Fatalf("read artwork %s: %v", cover, err)
		}
		cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil || format != "jpeg" || cfg.Width != 512 || cfg.Height != 512 {
			t.Fatalf("artwork %s: format=%q size=%dx%d err=%v", cover, format, cfg.Width, cfg.Height, err)
		}
		hash := fmt.Sprintf("%x", sha256.Sum256(data))
		if coverHashes[hash] {
			t.Fatalf("duplicate bundled cover image: %s", cover)
		}
		coverHashes[hash] = true
	}
}
