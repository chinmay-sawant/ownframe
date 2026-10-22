package store

import (
	"context"
	"database/sql"
	"errors"
)

func (s *Store) SeedDemo(ctx context.Context) (err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		rbErr := tx.Rollback()
		if rbErr == nil || errors.Is(rbErr, sql.ErrTxDone) {
			return
		}
		if err == nil {
			err = rbErr
		} else {
			err = errors.Join(err, rbErr)
		}
	}()
	var seeded int
	err = tx.QueryRowContext(ctx, `SELECT 1 FROM app_meta WHERE key='demo-seed-v1'`).Scan(&seeded)
	if err == nil {
		return tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	var pets int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM pets`).Scan(&pets); err != nil {
		return err
	}
	if pets == 0 {
		if err := seedDemoRows(ctx, tx); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO app_meta(key,value) VALUES('demo-seed-v1','done')`); err != nil {
		return err
	}
	return tx.Commit()
}
