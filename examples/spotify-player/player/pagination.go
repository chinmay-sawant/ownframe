package player

import "fmt"

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
	view.TrackPage, view.TrackPages = page, pages
	view.PageTracks = view.Tracks[start:end]
	view.HasPrev, view.HasNext = page > 0, page+1 < pages
	view.TrackRange = fmt.Sprintf("Showing %d–%d of %d · Page %d of %d",
		start+1, end, count, page+1, pages)
	if count == 0 {
		view.TrackRange = "Showing 0 tracks"
	}
}
