package main

// scrollwinCSS paints the Discord-clone windowing chrome leftovers: the
// channel mention badge hide state. servers.go textRow always renders the
// badge span and adds dc-chanbadge--hidden when the channel has zero
// mentions, so this sheet hides that state without touching layout.
//
// Const name: scrollwinCSS. Wiring: append it after appCSS in themeSheet
// (app.go), which this file must not touch:
// chadcn.Dark + chadcn.Stylesheet + appCSS + scrollwinCSS.
//
// Class audit against the live markup (messages.go, scroll.go, servers.go,
// chadcn.Scrollbar):
//   - dc-chanbadge--hidden: PRESENT (servers.go textRow) — styled below.
//   - dc-scrollpad: ABSENT — no such class or element exists; the scroll
//     region is #dc-scroll-viewport > #dc-scroll-content plus the overlay
//     from chadcn.Scrollbar (track #dc-scroll-bar, thumb
//     #dc-scroll-thumb.of-scrollbar-thumb). No rule emitted; do not invent
//     the markup.
//   - dc-scrollthumb: ABSENT as a class — the thumb carries ID
//     dc-scroll-thumb with class of-scrollbar-thumb, painted inline by
//     chadcn.Scrollbar. No rule emitted; do not invent the markup.
const scrollwinCSS = `
.dc-chanbadge--hidden { display: none; }
`

// scrollwinCSSAnchor keeps scrollwinCSS tied to the markup that needs it
// until themeSheet (app.go) appends the sheet; without a reference the
// unused linter flags the const.
var _ = scrollwinCSS
