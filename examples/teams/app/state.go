package app

import (
	"log"

	"github.com/chinmay-sawant/ownframe/examples/teams/store"
)

// viewFromStore turns the database state into the printable view. The rail
// and the demo tiles are not stored; they never change.
func viewFromStore(d store.Data) View {
	return View{
		Section:  "chat",
		Dark:     d.Dark,
		Presence: d.Presence,
		Rail:     railItems(),
		Apps:     appTiles(),
		Activity: d.Activity,
		Chat:     d.Chat,
		Channels: d.Channels,
		Calendar: d.Calendar,
		Calls:    d.Calls,
		Files:    d.Files,
	}
}

// dataFromView packs the view for the database.
func dataFromView(v View) store.Data {
	return store.Data{
		Activity: v.Activity,
		Chat:     v.Chat,
		Channels: v.Channels,
		Calendar: v.Calendar,
		Calls:    v.Calls,
		Files:    v.Files,
		Presence: v.Presence,
		Dark:     v.Dark,
	}
}

// loadView returns the state to draw at startup. It runs the seed SQL on
// the first run, then reads the state back.
func (a *App) loadView(st *store.Store) View {
	empty, err := st.Empty()
	if err != nil {
		log.Printf("teams: database state: %s", cleanLog(err.Error()))

		return DefaultView()
	}

	if empty {
		if err := st.Seed(); err != nil {
			log.Printf("teams: seed: %s", cleanLog(err.Error()))

			return DefaultView()
		}
	}

	data, ok, err := st.Load()
	if err != nil || !ok {
		log.Printf("teams: load: %s", cleanLog(err.Error()))

		return DefaultView()
	}

	return viewFromStore(data)
}

// save writes the whole view back. A nil store skips the write.
func (a *App) save() {
	if a.store == nil {
		return
	}

	if err := a.store.Save(dataFromView(a.view)); err != nil {
		log.Printf("teams: save: %s", cleanLog(err.Error()))
	}
}

// Close closes the state database.
func (a *App) Close() error {
	if a.store == nil {
		return nil
	}

	return a.store.Close()
}
