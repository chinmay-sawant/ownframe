package app

import (
	"log"

	"github.com/chinmay-sawant/ownframe/examples/teams/store"
)

// openStore opens the state database, or returns nil when it cannot be
// opened, so the app still starts from the sample data.
func openStore(path string) *store.Store {
	st, err := store.Open(path)
	if err == nil {
		return st
	}

	log.Printf("teams: database unavailable, using memory: %s", cleanLog(err.Error()))

	st, err = store.Open(store.Memory)
	if err != nil {
		log.Printf("teams: memory database: %s", cleanLog(err.Error()))

		return nil
	}

	return st
}
