package store

import (
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/teams/channels"
)

func loadTeams(db *sql.DB) ([]channels.Team, error) {
	var teams []channels.Team

	err := readInto(db, `SELECT id, name, initials, color, expanded FROM teams ORDER BY position`, &teams, func(r *sql.Rows, t *channels.Team) error {
		return r.Scan(&t.ID, &t.Name, &t.Initials, &t.Color, &t.Expanded)
	})
	if err != nil {
		return nil, err
	}

	type row struct {
		team string
		ch   channels.Channel
	}

	var rows []row

	err = readInto(db, `SELECT c.id, c.team_id, c.name, c.unread FROM channels c JOIN teams t ON t.id = c.team_id ORDER BY t.position, c.position`, &rows, func(r *sql.Rows, v *row) error {
		return r.Scan(&v.ch.ID, &v.team, &v.ch.Name, &v.ch.Unread)
	})
	if err != nil {
		return nil, err
	}

	byID := map[string]int{}
	for i, t := range teams {
		byID[t.ID] = i
	}

	for _, v := range rows {
		if i, ok := byID[v.team]; ok {
			teams[i].Channels = append(teams[i].Channels, v.ch)
		}
	}

	return teams, nil
}
