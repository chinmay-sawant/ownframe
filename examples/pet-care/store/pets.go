package store

import (
	"context"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/model"
)

func (s *Store) Pets(ctx context.Context) (_ []model.Pet, err error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,name,species,breed,birth_date,notes FROM pets ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	var pets []model.Pet
	for rows.Next() {
		var p model.Pet
		if err := rows.Scan(&p.ID, &p.Name, &p.Species, &p.Breed, &p.BirthDate, &p.Notes); err != nil {
			return nil, err
		}
		pets = append(pets, p)
	}
	return pets, rows.Err()
}

func (s *Store) SavePet(ctx context.Context, p model.Pet) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO pets VALUES(?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,species=excluded.species,breed=excluded.breed,birth_date=excluded.birth_date,notes=excluded.notes`, p.ID, p.Name, p.Species, p.Breed, p.BirthDate, p.Notes)
	return err
}

func (s *Store) DeletePet(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM pets WHERE id=?`, id)
	return err
}
