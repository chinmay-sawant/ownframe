package remote

func (a *App) setData() {
	a.setSensitivity(float64(a.view.Sensitivity))
	keys := map[string]Key{}
	for _, key := range append(faceKeys(), moreKeys()...) {
		key.Name = key.Label
		switch key.ID {
		case "volup":
			key.Name = "Volume up"
		case "voldn":
			key.Name = "Volume down"
		case "chup":
			key.Name = "Channel up"
		case "chdn":
			key.Name = "Channel down"
		}
		switch key.ID {
		case "up", "down", "left", "right", "play", "pause", "stop":
			key.Icon = true
			key.Class += " circ"
		case "ok", "rew", "ff":
			key.Class += " circ"
		case "hotstar", "prime", "youtube":
			key.Icon = true
			key.Class += " logo"
		case "power":
			if a.view.PowerOn {
				key.Class += " pwr-on"
			} else {
				key.Class += " pwr-off"
			}
		}
		_, bt := btCommand(key.ID)
		key.Disabled = a.view.Mode == "Bluetooth" && !bt && key.ID != "wake"
		if key.Disabled {
			key.Class += " unavailable"
		}
		key.Class += a.pressedClass(key.ID)
		keys[key.ID] = key
	}
	for _, key := range []Key{
		{ID: "wheelup", Label: "Wheel up"},
		{ID: "clicker", Label: "Click"},
		{ID: "wheeldown", Label: "Wheel down"},
	} {
		key.Name = key.Label
		key.Disabled = a.view.Mode == "Bluetooth"
		if key.Disabled {
			key.Class += " unavailable"
		}
		key.Class += a.pressedClass(key.ID)
		keys[key.ID] = key
	}
	a.view.Comfort = a.view.Layout == layoutComfort
	a.view.KeysNav = rows(keysNavRows, keys)
	for i := range a.view.KeysNav {
		a.view.KeysNav[i].Class = "midrow"
	}
	a.view.Dark = a.dark
	a.view.Accents = a.view.Accents[:0]
	for _, pick := range accents {
		colour := pick.Dark
		if !a.dark {
			colour = pick.Light
		}
		a.view.Accents = append(a.view.Accents, AccentOption{ID: pick.ID, Label: pick.Label, Colour: colour, On: pick.ID == a.view.Accent})
	}
	if a.view.Comfort {
		a.view.RemoteRows = rows(comfortRemoteRows, keys)
		a.view.NumberRows = rows(comfortNumberRows, keys)
		a.view.MediaRows = rows(comfortMediaRows, keys)
	} else {
		a.view.RemoteRows = rows(remoteRows, keys)
		a.view.NumberRows = rows(numberRows, keys)
		a.view.MediaRows = nil
	}
	a.view.PadKeys = rows("wheelup clicker wheeldown", keys)[0].Keys
	a.page.SetData(&a.view)
}
