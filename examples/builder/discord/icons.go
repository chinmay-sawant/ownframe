package main

import (
	"github.com/chinmay-sawant/ownframe/ui"
)

// Voice and channel icons as inline SVGs. Registered with Page.SetImage in
// newApp so ui img tags resolve them by name; stroke uses currentColor so
// the muted-red pressed tint applies without extra assets.
var dcIcons = map[string]string{
	"dc-mic":            `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="6" y="1" width="4" height="8" rx="2"/><path d="M3.5 7a4.5 4.5 0 0 0 9 0M8 11.5V14M5.5 14h5"/></svg>`,
	"dc-mic-off":        `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="6" y="1" width="4" height="8" rx="2"/><path d="M3.5 7a4.5 4.5 0 0 0 9 0M8 11.5V14M5.5 14h5M2 2l12 12"/></svg>`,
	"dc-headphones":     `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9V7a5 5 0 0 1 10 0v2"/><rect x="2" y="8.5" width="3" height="5" rx="1.2"/><rect x="11" y="8.5" width="3" height="5" rx="1.2"/></svg>`,
	"dc-headphones-off": `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9V7a5 5 0 0 1 10 0v2"/><rect x="2" y="8.5" width="3" height="5" rx="1.2"/><rect x="11" y="8.5" width="3" height="5" rx="1.2"/><path d="M2 2l12 12"/></svg>`,
	"dc-phone-down":     `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M3 3.5c2.5 0 7.5 0 10 0-.3 2.5-1.2 4.8-2.7 6.5l-1.5-1.5c-.8.5-1.8.8-2.8.8s-2-.3-2.8-.8L1.7 10C.2 8.3-.7 6 1 3.5z"/><path d="M2 13.5l3-3"/></svg>`,
	"dc-speaker":        `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M2 6v4h3l4 3V3L5 6H2z"/><path d="M11 5.5a3.5 3.5 0 0 1 0 5"/></svg>`,
}

// registerIcons hands the voice icons to the page as named image sources.
func (a *app) registerIcons() {
	if a == nil || a.page == nil {
		return
	}
	for name, svg := range dcIcons {
		a.page.SetImage(name, []byte(svg))
	}
}

// dcIconImg is one 16px voice icon resolving through the page image lookup.
func dcIconImg(name string) *ui.Element {
	return ui.Tag("img").Attr("src", name).Attr("aria-hidden", "true").
		Width(ui.Px(16)).Height(ui.Px(16))
}
