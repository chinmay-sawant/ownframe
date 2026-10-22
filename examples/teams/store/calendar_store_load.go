package store

import (
	"context"
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/teams/calendar"
)

// loadCalendar reads every stored event and the state row.
func loadCalendar(db *sql.DB) (calendar.Data, error) {
	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `SELECT id, week, title, time, dur, color, location, day
		FROM calendar_events ORDER BY week, position`)
	if err != nil {
		return calendar.Data{}, err
	}

	defer func() { _ = rows.Close() }()

	var events []calendar.Event

	for rows.Next() {
		var e calendar.Event
		if err := rows.Scan(&e.ID, &e.Week, &e.Title, &e.Time, &e.Dur,
			&e.Color, &e.Location, &e.Day); err != nil {
			return calendar.Data{}, err
		}

		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return calendar.Data{}, err
	}

	var week int
	var selected string

	if err := db.QueryRowContext(ctx, `SELECT week, selected FROM calendar_state
		WHERE id = 1`).Scan(&week, &selected); err != nil {
		return calendar.Data{}, err
	}

	return calendar.FromDB(week, events, selected), nil
}
