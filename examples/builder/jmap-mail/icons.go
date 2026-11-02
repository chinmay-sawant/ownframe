package main

import "github.com/chinmay-sawant/ownframe/ui"

// Icons are application-owned nodes; an image registered with Page.SetImage
// can replace the glyph without changing Button or any theme.
func settingsIcon() *ui.Element { return ui.Text("⚙").Attr("aria-hidden", "true") }
