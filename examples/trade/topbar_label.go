package main

// TopbarScreenLabel names the active screen in the topbar tab. The label is
// derived from View.Screen so the chrome needs no extra template field.
func (v View) TopbarScreenLabel() string {
	switch v.Screen {
	case "watchlist":
		return "Watchlist"
	case "markets":
		return "Markets"
	case "portfolio":
		return "Portfolio"
	case "orders":
		return "Orders"
	case "screener":
		return "Screener"
	case "news":
		return "News"
	case "settings":
		return "Settings"
	default:
		return "Terminal"
	}
}
