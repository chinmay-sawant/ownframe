package store

import (
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/teams/chat"
)

// loadChats reads the chat rows in position order.
func loadChats(db *sql.DB) ([]chat.Chat, error) {
	rows, err := db.Query(`SELECT id, name, initials, color, preview, time, unread,
		pinned, muted, is_group FROM chats ORDER BY position`)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var chats []chat.Chat

	for rows.Next() {
		var c chat.Chat

		if err := rows.Scan(&c.ID, &c.Name, &c.Initials, &c.Color, &c.Preview,
			&c.Time, &c.Unread, &c.Pinned, &c.Muted, &c.Group); err != nil {
			return nil, err
		}

		chats = append(chats, c)
	}

	return chats, rows.Err()
}

// loadThreads reads every message and groups it by chat id.
func loadThreads(db *sql.DB) (map[string][]chat.Message, error) {
	rows, err := db.Query(`SELECT id, chat_id, author, initials, color, time, text,
		own FROM chat_messages ORDER BY chat_id, position`)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	threads := map[string][]chat.Message{}

	for rows.Next() {
		var (
			id  string
			msg chat.Message
		)

		if err := rows.Scan(&msg.ID, &id, &msg.Author, &msg.Initials, &msg.Color,
			&msg.Time, &msg.Text, &msg.Own); err != nil {
			return nil, err
		}

		threads[id] = append(threads[id], msg)
	}

	return threads, rows.Err()
}
