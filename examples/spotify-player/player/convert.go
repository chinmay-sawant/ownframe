package player

import (
	"strconv"
	"strings"
)

// tracksFrom converts songs into tracklist rows.
func tracksFrom(songs []searchResult) []Track {
	tracks := make([]Track, 0, len(songs))

	for i, s := range songs {
		album := s.CollectionName
		if album == "" {
			album = s.Genre
		}

		tracks = append(tracks, Track{
			Index: i, Num: strconv.Itoa(i + 1),
			Title: clip(s.TrackName, trackTitleMax), Artist: clip(s.ArtistName, trackArtistMax),
			Album: clip(album, albumMax), Year: yearOf(s.ReleaseDate),
			Genre:  clip(s.Genre, genreMax),
			Length: formatMillis(s.TrackMillis),
			Cover:  "track-" + strconv.Itoa(i),
			Active: i == 0,
		})
	}

	return tracks
}

// cardsFrom converts albums into six greeting tiles and six shelf cards.
func cardsFrom(albums []searchResult) (picks, shelf []Card) {
	for i, al := range albums {
		if i < maxSongs {
			picks = append(picks, Card{Index: i, Title: clip(al.CollectionName, pickTitleMax), Sub: clip(al.ArtistName, pickTitleMax), Cover: "pick-" + strconv.Itoa(i)})
			continue
		}

		j := i - coverSplit
		shelf = append(shelf, Card{Index: j, Title: clip(al.CollectionName, shelfTitleMax), Sub: clip(al.ArtistName, shelfTitleMax), Cover: "card-" + strconv.Itoa(j)})
	}

	return picks, shelf
}

// playlistsFrom fills the sidebar library from the shelf cards.
func playlistsFrom(shelf []Card) []Playlist {
	out := make([]Playlist, 0, playlistsMax)

	for _, card := range shelf {
		if len(out) == playlistsMax {
			break
		}

		out = append(out, Playlist{
			Index: len(out),
			Name:  clip(card.Title, playlistNameMax),
			Meta:  clip("Playlist · "+card.Sub, playlistNameMax),
			Cover: card.Cover,
		})
	}

	return out
}

// upscale asks iTunes for a larger copy of the same artwork.
func upscale(raw string) string {
	return strings.Replace(raw, "100x100bb", "300x300bb", 1)
}
