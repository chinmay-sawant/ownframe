package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/chinmay-sawant/ownframe/examples/teams/chat"
)

// loadChat reads every chat and thread, then rebuilds the view state. The
// saved query is left out on purpose: a search box is per-visit state, and a
// stale query would hide every chat on the next run.
func loadChat(db *sql.DB) (chat.Data, error) {
	chats, err := loadChats(db)
	if err != nil {
		return chat.Data{}, err
	}

	threads, err := loadThreads(db)
	if err != nil {
		return chat.Data{}, err
	}

	var active, filter string

	err = db.QueryRowContext(context.Background(), `SELECT active, filter FROM chat_state WHERE id = 1`).
		Scan(&active, &filter)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return chat.Data{}, err
	}

	return chat.FromDB(chats, threads, filter, "", active), nil
}
