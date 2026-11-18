package remote

import "strings"

const (
	primeSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">` +
		`<circle cx="32" cy="32" r="30" fill="#00A8E1"/>` +
		`<path fill="none" stroke="#fff" stroke-width="4" d="M16 36c6 8 26 8 32-2"/>` +
		`<path fill="#fff" d="M44 28l8 8-10 2z"/>` +
		`</svg>`
	youTubeSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">` +
		`<rect width="64" height="64" rx="16" fill="#FF0000"/>` +
		`<polygon points="26,18 50,32 26,46" fill="#fff"/>` +
		`</svg>`
	arrowNSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">` +
		`<circle cx="32" cy="32" r="30" fill="#2a2a30"/>` +
		`<polygon points="32,14 50,42 14,42" fill="#f4f4f5"/>` +
		`</svg>`
	arrowSSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">` +
		`<circle cx="32" cy="32" r="30" fill="#2a2a30"/>` +
		`<polygon points="14,22 50,22 32,50" fill="#f4f4f5"/>` +
		`</svg>`
	arrowWSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">` +
		`<circle cx="32" cy="32" r="30" fill="#2a2a30"/>` +
		`<polygon points="42,14 42,50 14,32" fill="#f4f4f5"/>` +
		`</svg>`
	arrowESVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">` +
		`<circle cx="32" cy="32" r="30" fill="#2a2a30"/>` +
		`<polygon points="22,14 50,32 22,50" fill="#f4f4f5"/>` +
		`</svg>`
)

const gearSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">` +
	`<g fill="#f4f4f5">` +
	`<rect x="28" y="6" width="8" height="14" rx="2"/>` +
	`<rect x="28" y="44" width="8" height="14" rx="2"/>` +
	`<rect x="6" y="28" width="14" height="8" rx="2"/>` +
	`<rect x="44" y="28" width="14" height="8" rx="2"/>` +
	`<rect x="28" y="6" width="8" height="14" rx="2" transform="rotate(45 32 32)"/>` +
	`<rect x="28" y="44" width="8" height="14" rx="2" transform="rotate(45 32 32)"/>` +
	`<rect x="6" y="28" width="14" height="8" rx="2" transform="rotate(45 32 32)"/>` +
	`<rect x="44" y="28" width="14" height="8" rx="2" transform="rotate(45 32 32)"/>` +
	`<circle cx="32" cy="32" r="19"/></g>` +
	`<circle cx="32" cy="32" r="8" fill="#2a2a30"/></svg>`

// gearDarkSVG is the gear drawn for the light theme.
func gearDarkSVG() string {
	return strings.NewReplacer("#f4f4f5", "#1a1a1a", "#2a2a30", "#e7e2d8").Replace(gearSVG)
}

func (a *App) installIcons() {
	if a.page == nil {
		return
	}

	a.page.SetImage("logo-prime", []byte(primeSVG))
	a.page.SetImage("logo-hotstar", hotstarIcon)
	a.page.SetImage("logo-youtube", []byte(youTubeSVG))
	a.page.SetImage("icon-settings", []byte(gearSVG))
	a.page.SetImage("icon-settings-dark", []byte(gearDarkSVG()))
	a.page.SetImage("arrow-n", []byte(arrowNSVG))
	a.page.SetImage("arrow-s", []byte(arrowSSVG))
	a.page.SetImage("arrow-w", []byte(arrowWSVG))
	a.page.SetImage("arrow-e", []byte(arrowESVG))
}
