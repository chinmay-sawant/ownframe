// Search palette, user settings, create-server dialog, toast, and message
// menu for the discord demo.
//
// Mount: screen() includes a.overlays() once at the root; every builder
// returns nil while closed, so the aggregator paints nothing when all state
// is clear. a.overlays() owns the toast: it replaces shell dcToast, so
// remove that call when wiring or two toasts stack (dcToast also reads the
// retired a.toast field, which no longer exists on app).
//
// Actions emitted (handled by the host click dispatcher in actions.go):
// search:close, chan:select:<id>, msg:jump:<id>, settings:close,
// settings:tab:<name>, settings:save, settings:reset, members:toggle,
// voice:mute, voice:deafen, voice:leave, user:edit-profile,
// server:confirm-create, server:cancel-create, toast:close,
// msg:reply:<id>, msg:react:<id>:<emoji>, msg:pin:<id>, msg:delete:<id>,
// msg:copy:<id>.
// Gap: actions.go has no msg:copy: case yet. The Copy row dispatches it and
// needs a handler there (copy the message text, toast "Copied").
// Click-away and Escape already close these through overlay:dismiss (the
// shell dismiss layer) and closeOverlays, so no extra actions are needed.
//
// Form ids read by handleChange / handleClick:
// "dc-search" (palette query; the input id is exactly this, not
// "dc-search-input", because handleChange matches box.ID "dc-search"),
// "dc-new-server" (server name, read on server:confirm-create).
//
// The palette is hand-built from chadcn primitives instead of chadcn.Search
// or chadcn.Command: Search and Command rows bind Go callbacks and cannot
// carry data-action, while every result here must route through the click
// dispatcher as chan:select:<id> or msg:jump:<id>.
package main

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// Overlay form ids. handleChange matches the search id literally and
// confirmCreate reads the server name on submit.
const (
	overlaySearchInputID = "dc-search"
	overlayServerNameID  = "dc-new-server"
)

// Overlay geometry and result caps.
const (
	overlaySearchWidth      = 560
	overlaySearchTop        = 72
	overlaySearchZ          = 100
	overlaySearchChannelCap = 8
	overlaySearchMessageCap = 12
	overlayRowSnippetLen    = 72
)

// overlayMsgMenuFallbackX/Y is the menu spot when no click anchor exists,
// such as in tests that open the menu without geometry. Live clicks always
// carry the hovered row box, so this path stays out of the way.
const (
	overlayMsgMenuFallbackX = 660
	overlayMsgMenuFallbackY = 300
)

// overlaySettingsTabs orders the settings tabs. The suffix of
// settings:tab:<name> is stored verbatim as settingsTab, so these strings
// are both the action values and the active-tab keys.
var overlaySettingsTabs = []string{"My Account", "Profile", "Voice", "Appearance"}

// overlays aggregates the five overlay builders for one screen() call site.
// Nil children are ignored, so a quiet state renders an empty wrapper.
func (a *app) overlays() *ui.Element {
	return ui.Div().ID("dc-overlays").Children(
		a.searchOverlay(),
		a.settingsSheet(),
		a.createDialog(),
		a.messageMenu(),
		a.toast(),
	)
}

// searchOverlay is the quick-switcher palette: query field plus channel and
// message hits. Nil unless searchOpen. Result rows dispatch chan:select:<id>
// and msg:jump:<id>; the close control dispatches search:close, and clicks
// outside the panel reach the shell dismiss layer (overlay:dismiss).
func (a *app) searchOverlay() *ui.Element {
	if !a.searchOpen {
		return nil
	}
	query := strings.ToLower(strings.TrimSpace(a.searchQuery))
	channels, channelTotal := a.overlayChannelHits(query)
	messages, messageTotal := a.overlayMessageHits(query)

	var sections []*ui.Element
	if len(channels) > 0 {
		sections = append(sections,
			chadcn.Label("Channels"),
			ui.Div().Class("dc-search-results").Children(channels...),
		)
		sections = append(sections, overlayCappedNote(len(channels), channelTotal))
	}
	if len(messages) > 0 {
		if len(channels) > 0 {
			sections = append(sections, chadcn.Separator(chadcn.SeparatorOptions{}))
		}
		sections = append(sections,
			chadcn.Label("Messages"),
			ui.Div().Class("dc-search-results").Children(messages...),
		)
		sections = append(sections, overlayCappedNote(len(messages), messageTotal))
	}
	if len(sections) == 0 {
		hint := "No channels or messages yet."
		if strings.TrimSpace(a.searchQuery) != "" {
			hint = "No results for \"" + strings.TrimSpace(a.searchQuery) + "\""
		}
		sections = append(sections, chadcn.FieldDescription(hint))
	}

	panel := chadcn.Card(
		chadcn.CardHeader(
			ui.Div().Class("dc-search-head").Children(
				chadcn.CardTitle("Search"),
				chadcn.Button("×", chadcn.ButtonOptions{Variant: chadcn.ButtonText, Size: chadcn.ButtonSmall}).
					ID("dc-search-x").
					Attr("data-action", "search:close").
					Attr("aria-label", "Close search"),
			),
		),
		chadcn.CardContent(
			ui.Div().Class("dc-search-field").Children(
				chadcn.Field(chadcn.FieldOptions{
					Label:   "Search channels and messages",
					For:     overlaySearchInputID,
					Control: chadcn.InputWithOptions(overlaySearchInputID, a.searchQuery, chadcn.InputOptions{Type: "search", Placeholder: "Search channels and messages"}),
				}),
			),
			ui.Div().Class("dc-search-sections").Children(sections...),
		),
	)
	return ui.Div().ID("dc-search-panel").
		Attr("role", "dialog").
		Attr("aria-label", "Quick switcher").
		Position(ui.PositionFixed).
		Top(ui.Px(overlaySearchTop)).
		Left(ui.Percent(50)).
		Width(ui.Px(overlaySearchWidth)).
		MarginLeft(ui.Px(-overlaySearchWidth / 2)).
		ZIndex(ui.Number(overlaySearchZ)).
		Children(panel)
}

// dcSettingsSlug maps a settings tab name to its stable element id suffix.
func dcSettingsSlug(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "&", "")
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return strings.Trim(slug, "-")
}

// overlayChannelHits filters every channel in server order, capped for the
// palette. It returns the rows and the uncapped match total.
func (a *app) overlayChannelHits(query string) ([]*ui.Element, int) {
	rows := make([]*ui.Element, 0, overlaySearchChannelCap)
	total := 0
	for _, s := range a.servers {
		for _, ch := range a.channels[s.ID] {
			if !overlayMatchFold(ch.Name+" "+ch.Topic, query) {
				continue
			}
			total++
			if len(rows) >= overlaySearchChannelCap {
				continue
			}
			detail := s.Name
			if strings.TrimSpace(ch.Topic) != "" {
				detail += " · " + ch.Topic
			}
			rows = append(rows, ui.Button("").Attr("type", "button").
				ID("dc-hit-chan-"+ch.ID).
				Attr("data-action", "chan:select:"+ch.ID).
				Attr("title", detail).
				Children(
					ui.Text("# "+ch.Name).Class("dc-search-name"),
					ui.Text(detail).Class("dc-search-detail"),
				))
		}
	}
	return rows, total
}

// overlayMessageHits filters every message in server and channel order,
// capped for the palette. It returns the rows and the uncapped match total.
func (a *app) overlayMessageHits(query string) ([]*ui.Element, int) {
	rows := make([]*ui.Element, 0, overlaySearchMessageCap)
	total := 0
	for _, s := range a.servers {
		for _, ch := range a.channels[s.ID] {
			for _, m := range a.messages[ch.ID] {
				if !overlayMatchFold(m.Author+" "+m.Text, query) {
					continue
				}
				total++
				if len(rows) >= overlaySearchMessageCap {
					continue
				}
				rows = append(rows, ui.Button("").Attr("type", "button").
					ID("dc-hit-msg-"+m.ID).
					Attr("data-action", "msg:jump:"+m.ID).
					Attr("title", "#"+ch.Name+" · "+m.Time).
					Children(
						ui.Text(overlaySnippet(m.Author+": "+m.Text, overlayRowSnippetLen)).Class("dc-search-name"),
						ui.Text("#"+ch.Name+" · "+m.Time).Class("dc-search-detail"),
					))
			}
		}
	}
	return rows, total
}

// overlayCappedNote reports hidden hits when the cap clips the list.
func overlayCappedNote(shown, total int) *ui.Element {
	if total <= shown {
		return nil
	}
	return chadcn.FieldDescription("Showing " + strconv.Itoa(shown) + " of " + strconv.Itoa(total))
}

// settingsSheet is the fullscreen user-settings overlay, modeled on the real
// Discord client: a left nav rail with grouped sections on the left, the
// active page content on the right, and an ESC control top-right. Nil unless
// settingsOpen. Nav rows dispatch settings:tab:<name> (stored verbatim as
// settingsTab, so "Voice & Video" and "Game Overlay" need no dispatcher
// change); the ESC control dispatches settings:close. Escape already closes
// through closeOverlays, so no extra key handling is needed here. Voice rows
// reuse the voice:* actions, the member row reuses members:toggle, the edit
// control reuses the user:edit-profile stub, and overlay toggles reuse the
// generic flag:<key> action from actions.go, so no new contract is
// introduced.
//
// New dc-st-* classes used below (ZERO stylesheet rules are defined in this
// file; the integrator wires them in styles.go):
//
//	dc-st                 - inner content column centering the nav+main pair.
//	dc-st-nav             - left nav rail, fixed 218px wide.
//	dc-st-navsection      - small-caps group label (USER SETTINGS, ...).
//	dc-st-navitem         - nav row button dispatching settings:tab:<name>.
//	dc-st-navitem--active - modifier on the row matching the active tab.
//	dc-st-main            - right content column for the active tab.
//	dc-st-title           - tab heading text ("My Account", ...).
//	dc-st-card            - bordered content card grouping related rows.
//	dc-st-row             - one label/control line inside a card.
//	dc-st-esc             - top-right close control (dispatches settings:close).
//	dc-st-key             - keycap chip for static shortcut text (ESC, SHIFT).
//
// settingsSheet is the fullscreen user-settings page. It docks below the
// 48px custom topbar so the brand, search, settings, and window controls
// stay visible and clickable while settings is open.
func (a *app) settingsSheet() *ui.Element {
	if !a.settingsOpen {
		return nil
	}
	tab := a.settingsTab
	if tab == "" {
		tab = overlaySettingsTabs[0]
	}
	if tab == "Voice" {
		tab = "Voice & Video"
	}
	navItem := func(name string) *ui.Element {
		class := "dc-st-navitem"
		if name == tab {
			class += " dc-st-navitem--active"
		}
		return ui.Button(name).
			Attr("type", "button").
			ID("dc-st-tab-"+dcSettingsSlug(name)).
			Class(class).
			Attr("data-action", "settings:tab:"+name).
			Attr("aria-pressed", strconv.FormatBool(name == tab))
	}
	nav := ui.Div().Class("dc-st-nav").Children(
		ui.Text("USER SETTINGS").Class("dc-st-navsection"),
		navItem("My Account"),
		navItem("Profile"),
		ui.Text("APP SETTINGS").Class("dc-st-navsection"),
		navItem("Appearance"),
		navItem("Voice & Video"),
		navItem("Game Overlay"),
	)
	var body *ui.Element
	switch tab {
	case "Profile":
		body = a.settingsProfileSection()
	case "Voice & Video":
		body = a.settingsVoiceSection()
	case "Appearance":
		body = a.settingsAppearanceSection()
	case "Game Overlay":
		body = a.settingsOverlaySection()
	default:
		body = a.settingsAccountSection()
	}
	main := ui.Div().Class("dc-st-main").Children(
		ui.Text(tab).Class("dc-st-title"),
		body,
	)
	esc := ui.Div().Class("dc-st-escwrap").Children(
		ui.Button("×").
			Attr("type", "button").
			ID("dc-st-esc").
			Class("dc-st-esc").
			Attr("data-action", "settings:close").
			Attr("aria-label", "Close settings"),
		ui.Text("ESC").Class("dc-st-key"),
	)
	return ui.Div().ID("dc-settings").
		Attr("role", "dialog").
		Attr("aria-label", "User Settings").
		Position(ui.PositionFixed).
		Top(ui.Px(48)).
		Left(ui.Px(0)).
		Width(ui.Percent(100)).
		Height(ui.RawValue("calc(100vh - 48px)")).
		ZIndex(ui.Number(200)).
		Children(
			ui.Div().Class("dc-st").Children(nav, main, esc),
		)
}

// settingsAccountSection summarizes the demo account. The edit control
// reports user:edit-profile, which the dispatcher answers with a stub toast.
func (a *app) settingsAccountSection() *ui.Element {
	servers := "1 server"
	if len(a.servers) != 1 {
		servers = strconv.Itoa(len(a.servers)) + " servers"
	}
	return ui.Div().Class("dc-st-card").Children(
		ui.Div().Class("dc-st-row").Children(
			ui.Text("You are a member of "+servers+" in this demo."),
		),
		ui.Div().Class("dc-st-row").Children(
			ui.Button("Edit profile").
				Attr("type", "button").
				ID("dc-st-edit").
				Attr("data-action", "user:edit-profile"),
		),
	)
}

// settingsProfileSection holds the member list visibility toggle, reusing the
// members:toggle action the member rail already dispatches.
func (a *app) settingsProfileSection() *ui.Element {
	state := "OFF"
	if a.memberHidden {
		state = "ON"
	}
	return ui.Div().Class("dc-st-card").Children(
		ui.Div().Class("dc-st-row").Children(
			ui.Text("Hide member list"),
			ui.Text("Hide the member rail on the right."),
		),
		ui.Div().Class("dc-st-row").Children(
			ui.Button(state).
				Attr("type", "button").
				ID("dc-st-membertoggle").
				Attr("data-action", "members:toggle").
				Attr("aria-pressed", strconv.FormatBool(a.memberHidden)),
		),
	)
}

// settingsVoiceSection shows the current voice state with the same mute,
// deafen, and leave actions as the voice bar.
func (a *app) settingsVoiceSection() *ui.Element {
	if strings.TrimSpace(a.voiceID) == "" {
		return ui.Div().Class("dc-st-card").Children(
			ui.Div().Class("dc-st-row").Children(
				ui.Text("Join a voice channel from the sidebar to enable these controls."),
			),
			a.settingsVoiceButtons(true),
		)
	}
	state := "Connected to " + a.overlayChannelName(a.voiceID)
	status := "Neither muted nor deafened."
	switch {
	case a.voiceMuted && a.voiceDeafened:
		status = "Muted and deafened."
	case a.voiceMuted:
		status = "Muted."
	case a.voiceDeafened:
		status = "Deafened."
	}
	return ui.Div().Class("dc-st-card").Children(
		ui.Div().Class("dc-st-row").Children(ui.Text(state)),
		ui.Div().Class("dc-st-row").Children(ui.Text(status)),
		a.settingsVoiceButtons(false),
	)
}

// settingsVoiceButtons reuses the voice bar actions. Disabled while not in a
// voice channel so the panel never dispatches a meaningless toggle.
func (a *app) settingsVoiceButtons(disabled bool) *ui.Element {
	muteLabel := "Mute"
	if a.voiceMuted {
		muteLabel = "Unmute"
	}
	deafenLabel := "Deafen"
	if a.voiceDeafened {
		deafenLabel = "Undeafen"
	}
	voiceBtn := func(label, action string) *ui.Element {
		btn := ui.Button(label).
			Attr("type", "button").
			ID("dc-st-voice-"+strings.TrimPrefix(action, "voice:")).
			Attr("data-action", action)
		if disabled {
			btn.Attr("disabled", "true").Attr("aria-disabled", "true")
		}
		return btn
	}
	return ui.Div().Class("dc-st-row").Children(
		voiceBtn(muteLabel, "voice:mute"),
		voiceBtn(deafenLabel, "voice:deafen"),
		voiceBtn("Leave", "voice:leave"),
	)
}

// settingsAppearanceSection offers the Dark/Light choice like the mail
// settings card. The choice swaps the chadcn base sheet; the Discord shell
// follows through the .of-scope--light overrides. The member rail note
// mirrors the Profile toggle state so the page reads consistently.
func (a *app) settingsAppearanceSection() *ui.Element {
	members := "The member rail is currently shown."
	if a.memberHidden {
		members = "The member rail is currently hidden (see Profile)."
	}
	return ui.Div().Class("dc-st-card").Children(
		a.themeOption("dc-appearance-dark", "Dark", "Dim panels for low light.", "theme:dark", a.appearance != dcLight),
		a.themeOption("dc-appearance-light", "Light", "Bright panels for daytime reading.", "theme:light", a.appearance == dcLight),
		ui.Div().Class("dc-st-row").Children(ui.Text(members)),
	)
}

// themeOption is one Appearance row: a chadcn Choice that dispatches the
// theme action, mirroring the mail preview pattern.
func (a *app) themeOption(id, title, caption, action string, selected bool) *ui.Element {
	return chadcn.Choice(title, caption, selected).ID(id).Attr("data-action", action)
}

// settingsOverlaySection mirrors the real Discord Game Overlay page: the
// detected game plus overlay toggles. Toggles dispatch the generic
// flag:<key> action handled in actions.go; the lock shortcut row is static
// text. A nil flags map reads as all-off, so zero-value app structs render.
func (a *app) settingsOverlaySection() *ui.Element {
	toggle := func(label, desc, key string) *ui.Element {
		on := a.flags[key]
		state := "OFF"
		if on {
			state = "ON"
		}
		return ui.Div().Class("dc-st-row").Children(
			ui.Div().Children(
				ui.Text(label),
				ui.Text(desc),
			),
			ui.Button(state).
				Attr("type", "button").
				ID("dc-flag-"+key).
				Attr("data-action", "flag:"+key).
				Attr("aria-pressed", strconv.FormatBool(on)),
		)
	}
	return ui.Div().Class("dc-st-card").Children(
		ui.Div().Class("dc-st-row").Children(
			ui.Text("Currently Playing"),
			ui.Text("No game detected."),
		),
		toggle("Enable Overlay", "Show the overlay while in a game.", "overlay"),
		toggle("Enable Legacy Overlay", "Use the legacy overlay renderer.", "legacy"),
		ui.Div().Class("dc-st-row").Children(
			ui.Text("Toggle Overlay Lock"),
			ui.Text("SHIFT + `").Class("dc-st-keycap"),
		),
		toggle("Compatibility Mode", "Improve overlay support in older games.", "compat"),
		toggle("Clickable Regions", "Let overlay widgets receive pointer input.", "clickable"),
	)
}

// createDialog is the new-server dialog. Nil unless createOpen. The name
// field id is exactly "dc-new-server": server:create focuses it and
// server:confirm-create reads it. The close control and the backdrop both
// dispatch server:cancel-create; an empty name keeps the dialog open.
func (a *app) createDialog() *ui.Element {
	if !a.createOpen {
		return nil
	}
	return chadcn.Dialog(chadcn.DialogOptions{
		ID:          "dc-create",
		Open:        a.createOpen,
		Title:       "Create a server",
		Description: "Give your new community a name; it seeds a general text channel and a lobby voice channel.",
		CloseAction: "server:cancel-create",
		Width:       ui.Px(480),
		Footer: []*ui.Element{
			chadcn.Button("Cancel", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined}).
				ID("dc-create-cancel").
				Attr("data-action", "server:cancel-create"),
			chadcn.Button("Create", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).
				ID("dc-create-confirm").
				Attr("data-action", "server:confirm-create"),
		},
	},
		chadcn.Field(chadcn.FieldOptions{
			Label:       "Server name",
			For:         overlayServerNameID,
			Description: "Members see the first letter as the rail icon.",
			Control: chadcn.InputWithOptions(overlayServerNameID, "", chadcn.InputOptions{
				Placeholder: "e.g. Study Group",
			}),
		}),
	)
}

// toast is the transient notification. Nil unless toastOpen; the close
// control dispatches toast:close.
func (a *app) toast() *ui.Element {
	if !a.toastOpen {
		return nil
	}
	return chadcn.Toast(chadcn.ToastOptions{
		ID:          "dc-toast",
		Open:        true,
		Title:       a.toastTitle,
		Description: a.toastBody,
		CloseAction: "toast:close",
	})
}

// messageMenu is the context menu for the a.msgMenu message id (named
// messageMenu because app already owns the msgMenu state field; Go forbids
// a field and a method sharing one name). Nil while no menu is open. Items
// dispatch the msg:* actions handled in actions.go; the Copy row needs a
// msg:copy: case there (see the package note).
//
// Position follows Discord: the menu opens at the hovered row action and
// clamps to the viewport, flipping up near the bottom edge.
func (a *app) messageMenu() *ui.Element {
	if a.msgMenu == "" {
		return nil
	}
	label := "Message actions"
	pinLabel := "Pin message"
	if m, _, found := a.overlayFindMessage(a.msgMenu); found {
		if strings.TrimSpace(m.Author) != "" {
			label = "Message from " + m.Author
		}
		if m.Pinned {
			pinLabel = "Unpin message"
		}
	}
	id := a.msgMenu
	x, y := a.msgMenuPos()
	return chadcn.ContextMenu(chadcn.ContextMenuOptions{
		ID:    "dc-msg-menu",
		Label: label,
		Open:  true,
		X:     x,
		Y:     y,
		Items: []chadcn.MenuItem{
			{Label: "Reply", Action: "msg:reply:" + id},
			{Label: "React 👍", Action: "msg:react:" + id + ":👍"},
			{Label: pinLabel, Action: "msg:pin:" + id},
			{Label: "Copy text", Action: "msg:copy:" + id},
			{Separator: true},
			{Label: "Delete", Action: "msg:delete:" + id, Danger: true},
		},
	})
}

// msgMenuPos anchors the message menu like Discord: at the hovered row
// action when the click carried geometry, otherwise the fallback spot.
// The point clamps to the viewport and flips above the row near the
// bottom edge so the menu never clips.
func (a *app) msgMenuPos() (int, int) {
	const menuW, menuH = 240, 230
	const margin = 8
	winW, winH := a.windowSize()
	x, y := overlayMsgMenuFallbackX, overlayMsgMenuFallbackY
	if a != nil && a.msgAnchored {
		x, y = a.msgX, a.msgY
	}
	if x < margin {
		x = margin
	}
	if maxX := winW - menuW - margin; maxX > margin && x > maxX {
		x = maxX
	}
	if y < margin {
		y = margin
	}
	if y+menuH > winH-margin {
		flipped := y - menuH - 24
		if flipped >= margin {
			y = flipped
		} else if maxY := winH - menuH - margin; maxY >= margin {
			y = maxY
		}
	}
	return x, y
}

// overlayFindMessage locates a message by id across every server, returning
// the message and its channel name.
func (a *app) overlayFindMessage(id string) (dmessage, string, bool) {
	for _, s := range a.servers {
		for _, ch := range a.channels[s.ID] {
			for _, m := range a.messages[ch.ID] {
				if m.ID == id {
					return m, ch.Name, true
				}
			}
		}
	}
	return dmessage{}, "", false
}

// overlayChannelName resolves a channel id to its display name in server
// order, falling back to a generic label for unknown ids.
func (a *app) overlayChannelName(id string) string {
	for _, s := range a.servers {
		for _, ch := range a.channels[s.ID] {
			if ch.ID == id {
				return ch.Name
			}
		}
	}
	return "Voice channel"
}

// overlayMatchFold reports whether haystack contains query, case-insensitively.
// An empty query matches everything, so the palette browses while unfiltered.
func overlayMatchFold(haystack, query string) bool {
	if query == "" {
		return true
	}
	return strings.Contains(strings.ToLower(haystack), query)
}

// overlaySnippet truncates s to n runes for one-line result rows.
func overlaySnippet(s string, n int) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= n {
		return string(runes)
	}
	return string(runes[:n]) + "…"
}
