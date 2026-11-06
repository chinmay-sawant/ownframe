package main

import (
	"github.com/chinmay-sawant/ownframe/ui"
)

// Voice and channel icon bodies as inline SVG paths. The page rasterizes
// each img standalone, so currentColor never inherits and always paints
// black; every icon is therefore registered in three explicit strokes and
// the render path picks one for the theme and pressed state.
const (
	dcIconMicBody           = `<rect x="6" y="1" width="4" height="8" rx="2"/><path d="M3.5 7a4.5 4.5 0 0 0 9 0M8 11.5V14M5.5 14h5"/>`
	dcIconMicOffBody        = `<rect x="6" y="1" width="4" height="8" rx="2"/><path d="M3.5 7a4.5 4.5 0 0 0 9 0M8 11.5V14M5.5 14h5M2 2l12 12"/>`
	dcIconHeadphonesBody    = `<path d="M3 9V7a5 5 0 0 1 10 0v2"/><rect x="2" y="8.5" width="3" height="5" rx="1.2"/><rect x="11" y="8.5" width="3" height="5" rx="1.2"/>`
	dcIconHeadphonesOffBody = `<path d="M3 9V7a5 5 0 0 1 10 0v2"/><rect x="2" y="8.5" width="3" height="5" rx="1.2"/><rect x="11" y="8.5" width="3" height="5" rx="1.2"/><path d="M2 2l12 12"/>`
	dcIconPhoneDownBody     = `<path d="M3 3.5c2.5 0 7.5 0 10 0-.3 2.5-1.2 4.8-2.7 6.5l-1.5-1.5c-.8.5-1.8.8-2.8.8s-2-.3-2.8-.8L1.7 10C.2 8.3-.7 6 1 3.5z"/><path d="M2 13.5l3-3"/>`
	dcIconSpeakerBody       = `<path d="M2 6v4h3l4 3V3L5 6H2z"/><path d="M11 5.5a3.5 3.5 0 0 1 0 5"/>`
)

// Icon strokes: light for dark backgrounds, dark for light backgrounds,
// red for pressed mute/deafen. All three read on either theme.
const (
	dcIconLightStroke = "#dbdee1"
	dcIconDarkStroke  = "#5c5e66"
	dcIconRedStroke   = "#f23f43"
)

// dcIconBases lists the icon bodies registered in every stroke.
var dcIconBases = map[string]string{
	"mic":            dcIconMicBody,
	"mic-off":        dcIconMicOffBody,
	"headphones":     dcIconHeadphonesBody,
	"headphones-off": dcIconHeadphonesOffBody,
	"phone-down":     dcIconPhoneDownBody,
	"speaker":        dcIconSpeakerBody,
}

// dcIconSVG wraps one body in a 16px stroke SVG document.
func dcIconSVG(stroke, body string) string {
	return `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="` +
		stroke + `" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">` + body + `</svg>`
}

// registerIcons hands every voice icon in every stroke to the page as named
// image sources, once at startup so theme switches need no re-registration.
func (a *app) registerIcons() {
	if a == nil || a.page == nil {
		return
	}
	for base, body := range dcIconBases {
		a.page.SetImage("dc-"+base+"-light", []byte(dcIconSVG(dcIconLightStroke, body)))
		a.page.SetImage("dc-"+base+"-dark", []byte(dcIconSVG(dcIconDarkStroke, body)))
		a.page.SetImage("dc-"+base+"-red", []byte(dcIconSVG(dcIconRedStroke, body)))
	}
}

// dcVoiceIcon is one 16px voice icon resolving through the page image
// lookup: red while pressed, otherwise light on dark backgrounds and dark
// on light ones.
func (a *app) dcVoiceIcon(base string, pressed bool) *ui.Element {
	stroke := "light"
	switch {
	case pressed:
		stroke = "red"
	case a != nil && a.appearance == dcLight:
		stroke = "dark"
	}
	return ui.Tag("img").Attr("src", "dc-"+base+"-"+stroke).Attr("aria-hidden", "true").
		Width(ui.Px(16)).Height(ui.Px(16))
}
