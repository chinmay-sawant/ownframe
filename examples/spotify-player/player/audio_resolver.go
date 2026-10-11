package player

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe/examples/music"
)

type bundleResolver struct{}

func (bundleResolver) Resolve(ctx context.Context, _ string, pick int) (music.Clip, error) {
	if err := ctx.Err(); err != nil {
		return music.Clip{}, err
	}
	tracks := bundledTracks()
	if len(tracks) == 0 {
		return music.Clip{}, fmt.Errorf("player: bundled audio catalog is empty")
	}
	track := tracks[(pick%len(tracks)+len(tracks))%len(tracks)]
	data, err := audioFiles.ReadFile(track.File)
	if err != nil {
		return music.Clip{}, fmt.Errorf("player: read bundled track %q: %w", track.Title, err)
	}
	return music.Clip{
		Title: track.Title, Creator: track.Creator, License: track.License,
		URL: track.Source, Data: data,
	}, nil
}
