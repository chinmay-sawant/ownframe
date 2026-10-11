package store

import (
	"context"
	"database/sql"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/model"
)

func (s *Store) Replace(ctx context.Context, b model.Backup) error {
	if err := ValidateBackup(b); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM records`); err != nil {
		return rollback(tx, err)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM pets`); err != nil {
		return rollback(tx, err)
	}
	for _, p := range b.Pets {
		if _, err = tx.ExecContext(ctx, `INSERT INTO pets VALUES(?,?,?,?,?,?)`, p.ID, p.Name, p.Species, p.Breed, p.BirthDate, p.Notes); err != nil {
			return rollback(tx, err)
		}
	}
	for _, r := range b.Records {
		if _, err = tx.ExecContext(ctx, `INSERT INTO records VALUES(?,?,?,?,?,?,?)`, r.ID, r.PetID, r.Category, r.Date, r.Note, r.Scheduled, r.Complete); err != nil {
			return rollback(tx, err)
		}
	}
	return tx.Commit()
}

func rollback(tx *sql.Tx, err error) error { _ = tx.Rollback(); return err }

func (s *Store) Backup(ctx context.Context) (model.Backup, error) {
	pets, err := s.Pets(ctx)
	if err != nil {
		return model.Backup{}, err
	}
	records, err := s.Records(ctx, "")
	if err != nil {
		return model.Backup{}, err
	}
	return model.Backup{Version: 1, Pets: pets, Records: records}, nil
}
