package board

import (
	"context"
	"time"
)

// Cell is one flap tile.
type Cell struct{ ID, Ch, Kind string }

// Row is one departure line. Cells splits the four strings into tiles.
type Row struct {
	Time     string
	Dest     string
	Platform string
	Status   string
	Cells    []Cell
}

// View is the data the template renders.
type View struct {
	Title  string
	Clock  string
	Livery string
	Night  bool
	Paused bool
	Status string
	Rows   []Row
}

// formatClock renders the header clock as HH:MM:SS.
func formatClock(t time.Time) string { return t.Format("15:04:05") }

// newView builds the first frame at feed.
func newView(feed int) View {
	return View{
		Title:  "Departures",
		Clock:  formatClock(time.Now()),
		Status: "live",
		Rows:   rowsAt(feed),
	}
}

// rowsAt returns the eight rows at feed; one or two status tiles step on.
func rowsAt(feed int) []Row {
	rows := make([]Row, rowCount)
	for i := range rows {
		d := seed[i]
		rows[i] = Row{
			Time:     d.time,
			Dest:     d.dest,
			Platform: d.plat,
			Status:   statusAt(i, feed),
		}
		rows[i].Cells = buildCells(i, rows[i])
	}

	return rows
}

// Advance moves to the next feed, then SetData and Redraw. It is the resync
// path tests and the dump use; a live tick retargets instead.
func (a *App) Advance(ctx context.Context) error {
	now := time.Now()
	a.feed++
	a.lastRotate = now
	a.view.Rows = rowsAt(a.feed)
	a.view.Clock = formatClock(now)
	a.page.SetData(&a.view)

	return a.page.Redraw(ctx)
}

// retarget moves the board to rows without a Redraw: the view takes the new
// rows, SetData stores them for the next real Redraw, and the flipper starts
// the cascade on the tiles whose glyph changed.
func (a *App) retarget(rows []Row) {
	a.view.Rows = rows
	a.page.SetData(&a.view)
	a.flip.retarget(rows)
}
