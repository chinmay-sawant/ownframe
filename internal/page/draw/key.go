package draw

import (
	"github.com/chinmay-sawant/blinkless/css"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// Viewport is the frame size in CSS pixels.
type Viewport struct {
	Width  int
	Height int
}

// Key is every input that changes the layout of a frame. Images is a
// counter because a resolver function cannot be compared.
//
// SourceHash and SourceLen cache the source identity: Unchanged compares
// those u64s first and only memcmps Source on a hash hit. SourceRev is
// the page's synced-source revision; when both sides carry a nonzero
// revision, equal revisions prove equal sources with zero memcmp.
//
// Source pins the full page bytes only while the key carries no revision
// (the KeyFor path). KeyForHashed drops the string once the revision is
// nonzero, so a committed key keeps identity (hash, length, revision)
// without pinning a second copy of the page; the zero-revision fallback
// keeps the string because it has no revision to compare.
type Key struct {
	Source               string
	SourceHash           uint64
	SourceLen            int
	SourceRev            uint64
	Width, Height        int
	Focus, Hover, Active string
	Theme                *css.Sheet
	Images               uint64
}

// KeyFor drops pointer state the sheets cannot match. Leaving it out of the
// key lets a hover or a press land without forcing a layout.
func KeyFor(source string, view Viewport, state render.State, pseudos render.StatePseudos, images uint64) Key {
	if !pseudos.Hover {
		state.Hover = ""
	}
	if !pseudos.Active {
		state.Active = ""
	}
	if !pseudos.Focus {
		state.Focus = ""
	}

	return Key{
		Source:     source,
		SourceHash: HashString(source),
		SourceLen:  len(source),
		Width:      view.Width,
		Height:     view.Height,
		Focus:      state.Focus,
		Hover:      state.Hover,
		Active:     state.Active,
		Theme:      state.Theme,
		Images:     images,
	}
}

// KeyForHashed is KeyFor when the caller already cached the source hash,
// length, and revision (the page loop keeps them beside p.source), so a
// key costs no source scan. source must be the string the hash, length,
// and revision were taken from. It builds the Key directly instead of
// delegating to KeyFor, which would hash the source first and throw the
// result away. A nonzero revision drops the source string: Unchanged never
// reads Source on that path, so the key pins no page bytes. A zero
// revision keeps the string for the memcmp fallback.
func KeyForHashed(
	source string,
	hash uint64,
	length int,
	rev uint64,
	view Viewport,
	state render.State,
	pseudos render.StatePseudos,
	images uint64,
) Key {
	if !pseudos.Hover {
		state.Hover = ""
	}
	if !pseudos.Active {
		state.Active = ""
	}
	if !pseudos.Focus {
		state.Focus = ""
	}

	key := Key{
		SourceHash: hash,
		SourceLen:  length,
		SourceRev:  rev,
		Width:      view.Width,
		Height:     view.Height,
		Focus:      state.Focus,
		Hover:      state.Hover,
		Active:     state.Active,
		Theme:      state.Theme,
		Images:     images,
	}
	// A nonzero revision identifies the source without its bytes, so the
	// key drops the string and pins no page copy. A zero revision has no
	// such proof and keeps the string for the memcmp fallback.
	if rev == 0 {
		key.Source = source
	}

	return key
}

// Unchanged reports that this frame was already committed. Cheap fields go
// first; the source compares its revision, then its hash and length, and
// only memcmps the bytes on a hash hit, so an unchanged frame with a
// known revision pays no full-source memcmp.
func Unchanged(hasLast bool, last, next Key) bool {
	if !hasLast {
		return false
	}

	if last.Width != next.Width || last.Height != next.Height {
		return false
	}

	if last.Focus != next.Focus || last.Hover != next.Hover || last.Active != next.Active {
		return false
	}

	if last.Theme != next.Theme || last.Images != next.Images {
		return false
	}

	if last.SourceRev != 0 || next.SourceRev != 0 {
		if last.SourceRev != next.SourceRev {
			return false
		}

		if last.SourceHash != next.SourceHash || last.SourceLen != next.SourceLen {
			return false
		}

		// Equal revisions prove equal sources, so this path never reads
		// Source and committed keys may leave it empty.
		return true
	}

	if last.SourceHash != next.SourceHash || last.SourceLen != next.SourceLen {
		return false
	}

	return last.Source == next.Source
}
