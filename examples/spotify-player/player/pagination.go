package player

import (
	"strconv"
)

const trackPageSize = 5

func paginateTracks(view *View, page int) {
	count := len(view.Tracks)
	pages := (count + trackPageSize - 1) / trackPageSize
	if pages == 0 {
		pages = 1
	}
	if page < 0 {
		page = 0
	}
	if page >= pages {
		page = pages - 1
	}
	start := page * trackPageSize
	end := start + trackPageSize
	if end > count {
		end = count
	}

	if cap(view.PageTracks) < trackPageSize {
		view.PageTracks = make([]Track, 0, trackPageSize)
	}
	view.PageTracks = view.PageTracks[:0]
	view.PageTracks = append(view.PageTracks, view.Tracks[start:end]...)

	view.TrackPage, view.TrackPages = page, pages
	view.HasPrev, view.HasNext = page > 0, page+1 < pages

	if count == 0 {
		view.TrackRange = "Showing 0 tracks"
		return
	}

	buf := make([]byte, 0, 64)
	buf = append(buf, "Showing "...)
	buf = strconv.AppendInt(buf, int64(start+1), 10)
	buf = append(buf, "–"...)
	buf = strconv.AppendInt(buf, int64(end), 10)
	buf = append(buf, " of "...)
	buf = strconv.AppendInt(buf, int64(count), 10)
	buf = append(buf, " · Page "...)
	buf = strconv.AppendInt(buf, int64(page+1), 10)
	buf = append(buf, " of "...)
	buf = strconv.AppendInt(buf, int64(pages), 10)
	view.TrackRange = string(buf)
}
