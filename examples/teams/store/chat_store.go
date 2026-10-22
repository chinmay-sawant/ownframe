package store

import (
	"context"
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/teams/chat"
)

// saveChat replaces the chat tables with d.
func saveChat(tx *sql.Tx, d chat.Data) error {
	if _, err := tx.ExecContext(context.Background(), `DELETE FROM chats`); err != nil {
		return err
	}

	if _, err := tx.ExecContext(context.Background(), `DELETE FROM chat_messages`); err != nil {
		return err
	}

	for i, c := range d.AllChats() {
		_, err := tx.ExecContext(context.Background(), `INSERT INTO chats (id, position, name, initials, color,
			preview, time, unread, pinned, muted, is_group)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			c.ID, i, c.Name, c.Initials, c.Color, c.Preview, c.Time, c.Unread,
			c.Pinned, c.Muted, c.Group)
		if err != nil {
			return err
		}
	}

	for id, msgs := range d.AllThreads() {
		for i, m := range msgs {
			_, err := tx.ExecContext(context.Background(), `INSERT INTO chat_messages (id, chat_id, position,
				author, initials, color, time, text, own)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				m.ID, id, i, m.Author, m.Initials, m.Color, m.Time, m.Text, m.Own)
			if err != nil {
				return err
			}
		}
	}

	_, err := tx.ExecContext(context.Background(), `INSERT OR REPLACE INTO chat_state (id, active, filter, query)
		VALUES (1, ?, ?, '')`, d.Active, d.Filter)

	return err
}
