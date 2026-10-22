package player

// sampleStatus is the status line while the sample data is showing.
const sampleStatus = "Bundled collection (30 tracks)"

// defaultCredit is the now-bar credit before a bundled clip loads.
const defaultCredit = "Local MP3 collection · CC0 1.0" //nolint:gosec // reason: license label, not a credential.

// DefaultView returns the sample data the player draws offline.
func DefaultView() View {
	tracks := sampleTracks()
	now := tracks[0]

	view := View{
		Nav:        "home",
		Liked:      true,
		Volume:     defaultVolume,
		Progress:   defaultProgress,
		Elapsed:    "1:08",
		Remaining:  "2:24",
		Now:        now,
		Tracks:     tracks,
		Picks:      samplePicks(),
		Shelf:      sampleShelf(),
		Playlists:  samplePlaylists(),
		Status:     sampleStatus,
		ShelfTitle: "Recently played",
		Credit:     defaultCredit,
		Profile: Profile{
			Name:      "Chinmay",
			Playlists: defaultPlaylists,
			Followers: defaultFollowers,
			Following: defaultFollowing,
		},
		Search:     defaultSearch(),
		Library:    defaultLibrary(),
		LikedSongs: defaultLiked(),
		Browse:     defaultBrowse(),
		Radio:      defaultRadio(),
		Queue:      defaultQueue(),
	}
	paginateTracks(&view, 0)
	return view
}
