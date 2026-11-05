package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// notifItem is one row in the notifications popover: a title, a detail
// line, and a relative age.
type notifItem struct {
	id    string
	title string
	body  string
	age   string
}

// defaultNotifs returns the three seeded notifications shown until the
// user dismisses them.
func defaultNotifs() []notifItem {
	return []notifItem{
		{id: "n1", title: "Weekly Creator Brief #52 sent", body: "52.4% open · 21.2% click", age: "2h ago"},
		{id: "n2", title: "Sponsor payment received", body: "$1,812 · Creator Economy Report", age: "1d ago"},
		{id: "n3", title: "3 campaigns scheduled", body: "Referral Push · Win-back · Upsell", age: "2d ago"},
	}
}

// searchMatch pairs a display label with the canonical tab ID its result
// row navigates to.
type searchMatch struct {
	label string
	id    string
}

// topbarTabOrder lists canonical tabs in sidebar order so title matching
// stays deterministic. It returns a fresh slice per call so callers never
// share backing state.
func topbarTabOrder() []string {
	return []string{
		"overview", "audience", "campaigns", "content", "engagement",
		"revenue", "subscribers", "automations", "segments",
		"integrations", "analytics", "reports", "learn", "help", "settings",
	}
}

// searchMatches returns the sections matching the current query, matched
// case-insensitively against navSections labels and pageTitle titles.
// Results dedupe on the canonical tab ID.
func (a *app) searchMatches() []searchMatch {
	q := strings.ToLower(strings.TrimSpace(a.query))
	if q == "" {
		return nil
	}
	var out []searchMatch
	seen := make(map[string]bool)
	add := func(label string) {
		id := navID(label)
		if id == "" || !validTabs()[id] || seen[id] {
			return
		}
		seen[id] = true
		out = append(out, searchMatch{label: label, id: id})
	}
	for _, section := range navSections() {
		for _, item := range section.items {
			if strings.Contains(strings.ToLower(item.label), q) {
				add(item.label)
			}
		}
	}
	current := a.tab
	for _, tab := range topbarTabOrder() {
		a.tab = tab
		title, _ := a.pageTitle()
		if strings.Contains(strings.ToLower(title), q) {
			add(title)
		}
	}
	a.tab = current
	return out
}

// searchClear renders the × affordance that clears the query, or nil while
// the query is empty.
func (a *app) searchClear() *ui.Element {
	if strings.TrimSpace(a.query) == "" {
		return nil
	}
	return ui.Span().Class("mp-search-clear").
		Attr("data-action", "search:clear").
		Attr("title", "Clear search").
		Text("×")
}

// searchDrop renders the results dropdown under the search box, or nil
// while the query is empty.
func (a *app) searchDrop() *ui.Element {
	q := strings.TrimSpace(a.query)
	if q == "" {
		return nil
	}
	matches := a.searchMatches()
	rows := make([]*ui.Element, 0, len(matches)+1)
	for _, m := range matches {
		rows = append(rows, ui.Div().Class("mp-search-row").
			Attr("data-action", "nav:"+m.id).
			Attr("title", m.label).
			Text(m.label))
	}
	if len(matches) == 0 {
		rows = append(rows, ui.Div().Class("mp-search-empty").
			Text(fmt.Sprintf("No matches for %q", q)))
	}
	return ui.Div().Class("mp-search-drop").Children(rows...)
}

// notifTrigger builds the bell button with the notification card nested in a
// zero-height anchor beneath it. The card is absolutely positioned against
// the wrapper with pixel offsets (never percentages: this engine resolves
// percentage insets against the viewport), so opening it never shifts the
// topbar row and the card lands just below the bell.
func (a *app) notifTrigger() *ui.Element {
	bellIcon := ui.Div().Class("mp-bell-icon").Children(
		ui.Div().Class("mp-bell-dome").Attr("aria-hidden", "true"),
		ui.Div().Class("mp-bell-skirt").Attr("aria-hidden", "true"),
		ui.Div().Class("mp-bell-clap").Attr("aria-hidden", "true"),
	)
	bellBtn := chadcn.Button("", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
		Icon:    bellIcon,
	}).Attr("aria-label", "Notifications")
	trigger := ui.Span().Class("mp-bell-wrap").Children(
		bellBtn,
	).OnClick(func(ctx context.Context) error {
		a.notifOpen = !a.notifOpen
		a.userOpen = false
		return a.togglePanel(ctx, "mp-notif-card", a.notifOpen)
	})
	if len(a.notifs) > 0 {
		trigger.Children(ui.Span().Class("mp-bell-dot").Attr("aria-hidden", "true"))
	}
	anchor := ui.Div().Class("mp-panel-anchor")
	anchor.Children(ui.Div().ID("mp-notif-card").Class(notifCardClass(a.notifOpen)).Children(a.notifRows()...))
	return trigger.Children(anchor)
}

// notifRows renders the notification rows, or the empty state once every
// notification is dismissed.
func (a *app) notifRows() []*ui.Element {
	if len(a.notifs) == 0 {
		return []*ui.Element{
			ui.Div().Class("mp-notif-empty").Text("No new notifications"),
			ui.Div().Class("mp-notif-caught").Text("You're all caught up."),
		}
	}
	rows := make([]*ui.Element, 0, len(a.notifs))
	for _, n := range a.notifs {
		rows = append(rows, ui.Div().Class("mp-notif-row").Children(
			ui.Div().Class("mp-notif-main").Children(
				ui.Div().Class("mp-notif-title").Text(n.title),
				ui.Div().Class("mp-notif-body").Text(n.body),
				ui.Div().Class("mp-notif-age").Text(n.age),
			),
			ui.Span().Class("mp-notif-x").
				Attr("data-action", "notif:dismiss:"+n.id).
				Attr("title", "Dismiss").
				Text("×"),
		))
	}
	return rows
}

// togglePanel flips a header card between shown and hidden without a full
// refresh, so neighboring controls never repaint or flicker. A refused
// patch falls back to a refresh.
func (a *app) togglePanel(ctx context.Context, id string, open bool) error {
	err := a.page.PatchElement(ctx, id, ownframe.ElementEdit{Attribute: "class", Value: notifCardClass(open)})
	if err != nil {
		if errors.Is(err, ownframe.ErrPatchRefused) {
			return a.refresh(ctx)
		}
		return err
	}
	return nil
}

// notifCardClass reports the panel class for the open state. The panel is
// always rendered; hiding is a display class so toggling never re-renders
// the topbar and its neighbors never flicker.
func notifCardClass(open bool) string {
	if open {
		return "mp-float-panel"
	}
	return "mp-float-panel mp-hidden"
}

// userTrigger is the clickable avatar in the topbar, carrying the profile
// card in a zero-height anchor beneath it like the bell.
func (a *app) userTrigger() *ui.Element {
	trigger := ui.Span().Class("mp-avatar-btn").
		Attr("role", "button").
		Attr("aria-label", "Account").
		Attr("title", "Account").
		OnClick(func(ctx context.Context) error {
			a.userOpen = !a.userOpen
			a.notifOpen = false
			return a.togglePanel(ctx, "mp-user-card", a.userOpen)
		}).
		Children(chadcn.Avatar(chadcn.AvatarOptions{Initials: "JD"}))
	anchor := ui.Div().Class("mp-panel-anchor")
	anchor.Children(
		ui.Div().ID("mp-user-card").Class(notifCardClass(a.userOpen)).Children(
			ui.Div().Class("mp-user-head").Children(
				chadcn.Avatar(chadcn.AvatarOptions{Initials: "JD"}),
				ui.Div().Class("mp-user-text").Children(
					ui.Div().Class("mp-user-name").Text("Jordan Doe"),
					ui.Div().Class("mp-user-muted").Text("Admin \u00b7 jordan@mailpulse.co"),
				),
			),
			ui.Div().Class("mp-user-row").Attr("data-action", "user:profile").Text("My profile"),
			ui.Div().Class("mp-user-row").Attr("data-action", "nav:settings").Text("System settings"),
			ui.Div().Class("mp-user-row").Attr("data-action", "user:signout").Text("Sign out"),
		),
	)
	return trigger.Children(anchor)
}

// dismissNotif removes one notification, keeping the panel open.
func (a *app) dismissNotif(id string) {
	kept := make([]notifItem, 0, len(a.notifs))
	for _, n := range a.notifs {
		if n.id != id {
			kept = append(kept, n)
		}
	}
	a.notifs = kept
}

// handleSubmit navigates to the best search match on Enter, or toasts when
// nothing matches.
func (a *app) handleSubmit(ctx context.Context) error {
	matches := a.searchMatches()
	if len(matches) == 0 {
		a.toast = "No matches"
		a.toastOpen = true
		return a.refresh(ctx)
	}
	if tab := tabFromAction(matches[0].label); tab != "" {
		a.tab = tab
		a.selectedKind = ""
		a.selectedID = ""
		a.notifOpen = false
		a.userOpen = false
		if a.page != nil {
			a.page.ScrollTo(0, 0)
		}
	}
	return a.refresh(ctx)
}

// handleKey closes the header panels on Escape.
func (a *app) handleKey(ctx context.Context, key string) error {
	if key != "escape" {
		return nil
	}
	if !a.notifOpen && !a.userOpen {
		return nil
	}
	a.notifOpen = false
	a.userOpen = false
	return a.refresh(ctx)
}

// toastTTL is how long a toast stays visible before tickToast closes it.
const toastTTL = 5 * time.Second

// ensureToastTick arms the toast expiry after every refresh. It runs at
// the end of refresh so every toast path is covered; a closed toast clears
// a stale tick instead.
func (a *app) ensureToastTick() {
	if a.page == nil {
		return
	}
	if !a.toastOpen {
		if !a.toastUntil.IsZero() {
			a.toastUntil = time.Time{}
			a.page.SetTick(nil)
		}
		return
	}
	if !a.toastUntil.IsZero() {
		return
	}
	a.toastUntil = time.Now().Add(toastTTL)
	a.page.SetTick(a.tickToast)
	a.page.WantFrameAt(a.toastUntil)
}

// tickToast closes the toast once its deadline passes. The close goes
// through refresh, which sees a closed toast and arms no new tick, so the
// loop terminates.
func (a *app) tickToast(ctx context.Context) error {
	if a.toastOpen && time.Now().Before(a.toastUntil) {
		return nil
	}
	a.toastOpen = false
	a.toastUntil = time.Time{}
	if a.page != nil {
		a.page.SetTick(nil)
	}
	return a.refresh(ctx)
}
