package store

import (
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/teams/calendar"
)

// saveCalendar replaces both calendar tables with the state in d. Events are
// written in slice order, numbered per week.
func saveCalendar(tx *sql.Tx, d calendar.Data) error {
	if _, err := tx.Exec(`DELETE FROM calendar_events`); err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM calendar_state`); err != nil {
		return err
	}

	position := map[int]int{}

	for _, e := range d.AllEvents() {
		p := position[e.Week]
		position[e.Week] = p + 1

		if _, err := tx.Exec(`INSERT INTO calendar_events
			(id, week, position, title, time, dur, color, location, day)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			e.ID, e.Week, p, e.Title, e.Time, e.Dur,
			e.Color, e.Location, e.Day); err != nil {
			return err
		}
	}

	_, err := tx.Exec(`INSERT INTO calendar_state (id, week, selected)
		VALUES (1, ?, ?)`, d.Week, d.Selected)

	return err
}

// loadCalendar reads every stored event and the state row.
func loadCalendar(db *sql.DB) (calendar.Data, error) {
	rows, err := db.Query(`SELECT id, week, title, time, dur, color, location, day
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

	if err := db.QueryRow(`SELECT week, selected FROM calendar_state
		WHERE id = 1`).Scan(&week, &selected); err != nil {
		return calendar.Data{}, err
	}

	return calendar.FromDB(week, events, selected), nil
}
