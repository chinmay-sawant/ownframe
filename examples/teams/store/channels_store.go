package store

import (
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/teams/channels"
)

// saveChannels replaces the channels tables with d.
func saveChannels(tx *sql.Tx, d channels.Data) error {
	for _, t := range []string{"teams", "channels", "posts", "replies", "reactions", "channel_files", "channel_state"} {
		if _, err := tx.Exec("DELETE FROM " + t); err != nil { //nolint:gosec // reason: table names are internal constants from a fixed list, never user input
			return err
		}
	}

	if err := saveTeams(tx, d.Teams); err != nil {
		return err
	}

	if err := savePosts(tx, d.AllPosts()); err != nil {
		return err
	}

	if err := saveChannelFiles(tx, d.Files); err != nil {
		return err
	}

	_, err := tx.Exec(`INSERT OR REPLACE INTO channel_state (id, team, channel, tab) VALUES (1, ?, ?, ?)`, d.ActiveTeam, d.ActiveChannel, d.Tab)

	return err
}

// saveTeams writes the teams and their channels in slice order.
func saveTeams(tx *sql.Tx, teams []channels.Team) error {
	for i, t := range teams {
		if _, err := tx.Exec(`INSERT INTO teams (id, position, name, initials, color, expanded) VALUES (?, ?, ?, ?, ?, ?)`, t.ID, i, t.Name, t.Initials, t.Color, t.Expanded); err != nil {
			return err
		}

		for j, c := range t.Channels {
			if _, err := tx.Exec(`INSERT INTO channels (id, team_id, position, name, unread) VALUES (?, ?, ?, ?, ?)`, c.ID, t.ID, j, c.Name, c.Unread); err != nil {
				return err
			}
		}
	}

	return nil
}
