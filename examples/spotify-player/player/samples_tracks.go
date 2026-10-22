package player

import (
	"strconv"
)

// sampleTracks mirrors the bundled audio catalogue.
func sampleTracks() []Track {
	clips := bundledTracks()
	tracks := make([]Track, 0, len(clips))
	for i, clip := range clips {
		tracks = append(tracks, Track{
			Index: i, Num: strconv.Itoa(i + 1),
			Title: clip.Title, Artist: clip.Creator,
			Description: clip.Description,
			Album:       clip.Category, Year: "2026", Genre: clip.Genre,
			Length: formatSeconds(clip.DurationSeconds),
			Cover:  clip.Cover, Active: i == 0,
		})
	}

	return tracks
}

// samplePicks is the six-tile Good evening grid.
func samplePicks() []Card {
	return []Card{
		{Index: 0, Title: "Daily Mix 1", Sub: "Nova Waves & more", Cover: "cover-hybrid"},
		{Index: 1, Title: "Discover Weekly", Sub: "The Far Coast & more", Cover: "cover-spring"},
		{Index: sampleIdx2, Title: "Release Radar", Sub: "Aster Field & more", Cover: "cover-funk"},
		{Index: sampleIdx3, Title: "Deep Focus", Sub: "Mono Arcade & more", Cover: "cover-focus"},
		{Index: sampleIdx4, Title: "Lo-Fi Beats", Sub: "Night Cartography & more", Cover: "cover-chillhop"},
		{Index: sampleIdx5, Title: "On Repeat", Sub: "Echo Valley & more", Cover: "cover-soul"},
	}
}
