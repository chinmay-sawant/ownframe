package player

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
)

func TestBundledAudioDecodes(t *testing.T) {
	tracks := bundledTracks()
	if len(tracks) != 30 {
		t.Fatalf("bundled tracks = %d, want 30", len(tracks))
	}
	for i, track := range tracks {
		t.Run(track.Title, func(t *testing.T) {
			clip, err := (bundleResolver{}).Resolve(context.Background(), "", i)
			if err != nil {
				t.Fatal(err)
			}
			if clip.Title != track.Title || clip.Creator == "" || clip.License != "CC0 1.0 Universal" {
				t.Fatalf("clip metadata = %+v", clip)
			}
			if len(clip.Data) != track.Bytes || fmt.Sprintf("%x", sha256.Sum256(clip.Data)) != track.SHA256 {
				t.Fatalf("bundled bytes do not match catalog for %q", track.Title)
			}
			stream, err := mp3.DecodeWithSampleRate(48000, bytes.NewReader(clip.Data))
			if err != nil {
				t.Fatalf("decode MP3: %v", err)
			}
			buf := make([]byte, 4096)
			if n, err := stream.Read(buf); err != nil || n == 0 {
				t.Fatalf("read decoded audio: bytes=%d err=%v", n, err)
			}
			seconds := int(stream.Length() / (48000 * 4))
			if seconds < track.DurationSeconds-1 || seconds > track.DurationSeconds+1 {
				t.Fatalf("duration = %d seconds, catalog says %d", seconds, track.DurationSeconds)
			}
		})
	}
}
