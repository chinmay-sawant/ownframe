package player

import (
	"context"
	"strings"
)

// QueueData is the data behind the Queue screen: the songs waiting next and
// how many songs the queue holds in total.
type QueueData struct {
	Count int
	Next  []Track
}

// defaultQueue returns the offline queue screen.
func defaultQueue() QueueData {
	tracks := sampleTracks()[:6]
	for i := range tracks {
		tracks[i].Active = false
	}
	return QueueData{
		Count: len(tracks),
		Next:  tracks,
	}
}

// clickQueue applies one action on the queue screen.
func (a *App) clickQueue(ctx context.Context, action string) error {
	if strings.HasPrefix(action, "queue-row-") {
		i := slot(action, "queue-row-")
		if i < 0 || i >= len(a.view.Queue.Next) {
			return nil
		}

		for j := range a.view.Queue.Next {
			a.view.Queue.Next[j].Active = j == i
		}

		a.view.Now = a.view.Queue.Next[i]
		a.resetProgress()
		a.playNow(ctx)

		return nil
	}

	if strings.HasPrefix(action, "queue-from-") {
		i := slot(action, "queue-from-")
		if i < 0 || i >= len(a.view.Shelf) {
			return nil
		}

		for j := range a.view.Shelf {
			a.view.Shelf[j].Active = j == i
		}

		a.setNow(a.view.Shelf[i])
		a.playNow(ctx)
	}

	return nil
}
