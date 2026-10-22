package player

import "time"

// Layout, search, and sample-data numbers for the player.
const (
	centerDiv        = 2
	maxSongs         = 6
	maxAlbums        = 12
	percentScale     = 100
	seekStep         = 10
	seekBase         = 10
	volumeStep       = 10
	volumeBase       = 10
	trackTitleMax    = 44
	trackArtistMax   = 40
	albumMax         = 26
	genreMax         = 24
	pickTitleMax     = 32
	shelfTitleMax    = 19
	playlistsMax     = 5
	playlistNameMax  = 24
	coverSplit       = 6
	defaultProgress  = 32
	defaultPlaylists = 5
	defaultFollowers = 48
	defaultFollowing = 31
	eqIdleHeight     = 0.3
	eqBasePeriod     = 120
	eqMinHeight      = 0.2
	millisPerSec     = 1000
	secsPerMin       = 60
	timeParts        = 2
	dateMinLen       = 4
	logDirPerm       = 0o750
	paginationBuf    = 64
	decimalBase      = 10
	sampleIdx2       = 2
	sampleIdx3       = 3
	sampleIdx4       = 4
	sampleIdx5       = 5
	likedCount       = 214
	slowThreshold    = 2 * time.Second
	watchdogLimit    = 120
	trackTitleClip   = 30
	eqHalf           = 0.5
	progressMax      = 100
)
