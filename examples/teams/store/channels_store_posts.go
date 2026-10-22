package store

import (
	"context"
	"database/sql"

	"github.com/chinmay-sawant/ownframe/examples/teams/channels"
)

// savePosts writes every channel's posts with their replies and reactions.
func savePosts(tx *sql.Tx, posts map[string][]channels.Post) error {
	for id, list := range posts {
		for i, p := range list {
			if _, err := tx.ExecContext(context.Background(), `INSERT INTO posts (id, channel_id, position, author, initials, color, time, subject, text, pinned, likes, liked) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, p.ID, id, i, p.Author, p.Initials, p.Color, p.Time, p.Subject, p.Text, p.Pinned, p.Likes, p.Liked); err != nil {
				return err
			}

			for j, r := range p.Replies {
				if _, err := tx.ExecContext(context.Background(), `INSERT INTO replies (id, post_id, position, author, initials, color, time, text, own) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, r.ID, p.ID, j, r.Author, r.Initials, r.Color, r.Time, r.Text, r.Own); err != nil {
					return err
				}
			}

			for j, r := range p.Reactions {
				if _, err := tx.ExecContext(context.Background(), `INSERT INTO reactions (post_id, position, emoji, count, mine) VALUES (?, ?, ?, ?, ?)`, p.ID, j, r.Emoji, r.Count, r.Mine); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// saveChannelFiles writes the files of the active channel in slice order.
func saveChannelFiles(tx *sql.Tx, files []channels.FileItem) error {
	for i, f := range files {
		if _, err := tx.ExecContext(context.Background(), `INSERT INTO channel_files (id, position, name, badge, badge_text, kind, modified, modified_by, size, location, team, shared, starred) VALUES (?, ?, ?, ?, '', ?, ?, ?, ?, '', 0, 0, ?)`, f.ID, i, f.Name, f.Badge, f.Kind, f.Modified, f.ModifiedBy, f.Size, f.Starred); err != nil {
			return err
		}
	}

	return nil
}
