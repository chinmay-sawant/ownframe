package store

import (
	"context"
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/teams/calendar"
)

// saveCalendar replaces both calendar tables with the state in d. Events are
// written in slice order, numbered per week.
func saveCalendar(tx *sql.Tx, d calendar.Data) error {
	ctx := context.Background()
	if _, err := tx.ExecContext(ctx, `DELETE FROM calendar_events`); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM calendar_state`); err != nil {
		return err
	}

	position := map[int]int{}

	for _, e := range d.AllEvents() {
		p := position[e.Week]
		position[e.Week] = p + 1

		if _, err := tx.ExecContext(ctx, `INSERT INTO calendar_events
			(id, week, position, title, time, dur, color, location, day)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			e.ID, e.Week, p, e.Title, e.Time, e.Dur,
			e.Color, e.Location, e.Day); err != nil {
			return err
		}
	}

	_, err := tx.ExecContext(ctx, `INSERT INTO calendar_state (id, week, selected)
		VALUES (1, ?, ?)`, d.Week, d.Selected)

	return err
}
