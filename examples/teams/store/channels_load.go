package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/chinmay-sawant/ownframe/examples/teams/channels"
)

// loadChannels reads every table back and rebuilds the view state.
func loadChannels(db *sql.DB) (channels.Data, error) {
	teams, err := loadTeams(db)
	if err != nil {
		return channels.Data{}, err
	}

	posts, err := loadPosts(db)
	if err != nil {
		return channels.Data{}, err
	}

	index := postIndex(posts)

	if err := loadReplies(db, index); err != nil {
		return channels.Data{}, err
	}

	if err := loadReactions(db, index); err != nil {
		return channels.Data{}, err
	}

	files, err := loadChannelFiles(db)
	if err != nil {
		return channels.Data{}, err
	}

	var team, channel, tab string

	err = db.QueryRowContext(context.Background(), `SELECT team, channel, tab FROM channel_state WHERE id = 1`).Scan(&team, &channel, &tab)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return channels.Data{}, err
	}

	return channels.FromDB(teams, posts, files, team, channel, tab), nil
}

// loadTeams reads the teams, then attaches their channels in order.
