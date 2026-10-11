package player

// RadioData is the data behind the Radio screen.
type RadioData struct {
	Featured Card
	Stations []Card
}

// defaultRadio returns the offline radio screen.
func defaultRadio() RadioData {
	return RadioData{
		Featured: Card{Title: "Nova Waves Radio", Sub: "Nova Waves, Aster Field and more", Cover: "cover-late-night"},
		Stations: []Card{
			{Index: 0, Title: "Nova Waves Radio", Sub: "Nova Waves, Aster Field and more", Cover: "cover-late-night"},
			{Index: 1, Title: "Indie Coast Radio", Sub: "Indie Coast, The Far Coast and more", Cover: "cover-spring"},
			{Index: 2, Title: "Lo-Fi Nights Radio", Sub: "Lo-Fi Nights, Mono Arcade and more", Cover: "cover-chillhop"},
			{Index: 3, Title: "Synthwave FM", Sub: "Synthwave FM, Neon Horizon and more", Cover: "cover-funk"},
			{Index: 4, Title: "Acoustic Mornings Radio", Sub: "Acoustic Mornings, Paper Trails and more", Cover: "cover-hybrid"},
			{Index: 5, Title: "Deep Focus Radio", Sub: "Deep Focus, Blue Hour and more", Cover: "cover-focus"},
		},
	}
}
