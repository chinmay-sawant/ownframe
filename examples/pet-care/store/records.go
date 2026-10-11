package store

import (
	"context"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/model"
)

func (s *Store) Records(ctx context.Context, petID string) ([]model.Record, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,pet_id,category,date,note,scheduled,complete FROM records WHERE ?='' OR pet_id=? ORDER BY date DESC`, petID, petID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Record
	for rows.Next() {
		var r model.Record
		if err := rows.Scan(&r.ID, &r.PetID, &r.Category, &r.Date, &r.Note, &r.Scheduled, &r.Complete); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) SaveRecord(ctx context.Context, r model.Record) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO records VALUES(?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET pet_id=excluded.pet_id,category=excluded.category,date=excluded.date,note=excluded.note,scheduled=excluded.scheduled,complete=excluded.complete`, r.ID, r.PetID, r.Category, r.Date, r.Note, r.Scheduled, r.Complete)
	return err
}

func (s *Store) DeleteRecord(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM records WHERE id=?`, id)
	return err
}
