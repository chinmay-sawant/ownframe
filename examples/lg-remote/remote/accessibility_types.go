package remote

type accessibleNode struct {
	ID, Name, Value string
	X, Y, W, H      float64
	Enabled         bool
	Editable        bool
	Selected        bool
	Button          bool
	Slider          bool
	Progress        int
}

type accessState struct {
	gen  uint64
	x, y int
}

func (a *App) controlNames() map[string]Key {
	keys := map[string]Key{
		"theme":       {Name: "Switch to " + a.view.ThemeLabel + " theme"},
		"settings":    {Name: "Settings"},
		"accent-pink": {Name: "Pink accent"}, "accent-red": {Name: "Red accent"}, "accent-orange": {Name: "Orange accent"},
		"accent-green": {Name: "Green accent"}, "accent-blue": {Name: "Blue accent"}, "accent-purple": {Name: "Purple accent"},
		"settings-done":  {Name: "Close settings"},
		"set-dark":       {Name: "Dark theme"},
		"set-light":      {Name: "Light theme"},
		"layout-classic": {Name: "Classic layout"},
		"layout-comfort": {Name: "Comfort layout with larger keys"},
		"tab-keys":       {Name: "Keyboard"},
		"typing":         {Name: "Type to the TV"},
		"kb-back":        {Name: "Back on the TV"},
		"kb-delete":      {Name: "Delete a character on the TV"},
		"kb-enter":       {Name: "Enter or search on the TV"},
		"kb-clear":       {Name: "Clear the text"},
		"tab-media":      {Name: "Playback, guide, and apps"},
		"connect":        {Name: "Connect to TV"},
		"scan":           {Name: "Scan Wi-Fi for TVs"},
		"mode":           {Name: "Connection mode: " + a.view.Mode},
		"tab-remote":     {Name: "Remote"},
		"tab-pad":        {Name: "Pad"},
		"tab-nums":       {Name: "Numbers and more controls"},
		"pad":            {Name: "TV pointer pad. Tap to click. Drag to move.", Disabled: a.view.Mode != "Wi-Fi"},
		"host":           {Name: "TV IP address"},
		"sensitivity":    {Name: "Trackpad sensitivity"},
	}
	for _, row := range append(a.view.RemoteRows, a.view.NumberRows...) {
		for _, key := range row.Keys {
			keys[key.ID] = key
		}
	}
	for _, row := range a.view.MediaRows {
		for _, key := range row.Keys {
			keys[key.ID] = key
		}
	}
	for _, key := range a.view.PadKeys {
		keys[key.ID] = key
	}
	return keys
}
