package store

import "context"

// Save replaces every table with the state in d.
func (s *Store) Save(d Data) error {
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	if err := saveActivity(tx, d.Activity); err != nil {
		return fail(tx, err)
	}

	if err := saveChat(tx, d.Chat); err != nil {
		return fail(tx, err)
	}

	if err := saveChannels(tx, d.Channels); err != nil {
		return fail(tx, err)
	}

	if err := saveCalendar(tx, d.Calendar); err != nil {
		return fail(tx, err)
	}

	if err := saveCalls(tx, d.Calls); err != nil {
		return fail(tx, err)
	}

	if err := saveFiles(tx, d.Files); err != nil {
		return fail(tx, err)
	}

	if err := saveProfile(tx, d.Presence, d.Dark); err != nil {
		return fail(tx, err)
	}

	return tx.Commit()
}
