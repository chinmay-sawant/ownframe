package main

// Server rail + channel sidebar.
//
// Click contract: every control below dispatches a data-action string (no
// OnClick handlers here). The dispatcher in actions.go owns box.Action:
//
//	server:select:<id>  rail home ("home") or server tile; selects the server
//	server:create       rail "+" button; opens the create-server flow
//	server:menu        sidebar guild header; opens the server menu
//	cat:toggle:<cat>   category header; <cat> is the Category name verbatim
//	chan:select:<id>   text/announce channel row; selects the channel
//	voice:join:<id>    voice channel row; joins (or focuses) the channel
//	voice:leave        leaves the joined voice channel
//	voice:mute         toggles self mute
//	voice:deafen       toggles self deafen
//	settings:open      opens settings (rail gear + footer gear)
//
// State comes from app.go/data.go: a.servers, a.channels keyed by server
// ID, a.serverID/a.channelID ("home" means the DM view), a.collapsed keyed
// by category name, and a.voiceID/a.voiceMuted/a.voiceDeafened.
//
// Styling notes: Discord chrome (rail tiles, guild header, categories,
// channel rows, voice panel, user bar) uses the .dc-* classes from
// styles.go on plain elements, because that sheet owns the Discord paint.
// chadcn carries the floating pieces: Tooltip for rail names, Badge for
// rail mention counts, Avatar for the footer presence. Channel rows stay
// raw .dc-chan buttons instead of SidebarNavItem so the row keeps the
// sheet's active/unread/badge treatment; unread channels get
// .dc-chan--unread (white hash + label per the sheet) and only mentioned
// channels get a .dc-chan-badge count.

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// dcMeName and dcMeInitials are the footer user-bar persona. app.go keeps
// no self user; the demo speaks as Ada (owner in members.go, author in the
// seeded messages).
const (
	dcMeName     = "Ada Kim"
	dcMeInitials = "AK"
)

// rail is the 72px server column: home, a separator, one tile per server,
// then add and settings. Tiles show the server Initial on its Color,
// the active ring, the unread pill (both via .dc-server modifiers), and a
// mention count badge.
func (a *app) rail() *ui.Element {
	rail := ui.Nav().ID("dc-rail").Class("dc-rail").Attr("aria-label", "Servers").
		Children(a.homeTile(), ui.Div().Class("dc-rail-sep").Attr("aria-hidden", "true"))
	for _, s := range a.servers {
		rail.Children(a.serverTile(s))
	}
	return rail.Children(a.addTile(), a.settingsTile())
}

// homeTile selects the DM view.
func (a *app) homeTile() *ui.Element {
	button := ui.Button("\u2302").ID("dc-server-home").
		Class(railTileClass("dc-server dc-server--home", a.serverID == "home", false)).
		Attr("data-action", "server:select:home").
		Attr("aria-label", "Home").
		Attr("title", "Home")
	if a.serverID == "home" {
		button.Attr("aria-current", "page")
	}
	return chadcn.Tooltip(chadcn.TooltipOptions{
		ID: "dc-server-tip-home", CSSHover: true, Content: "Home", Side: chadcn.PopupRight,
	}, button)
}

// serverTile is one per-server tile with tooltip and optional mention badge.
func (a *app) serverTile(s server) *ui.Element {
	active := s.ID == a.serverID
	button := ui.Button(s.Initial).ID("dc-server-"+s.ID).
		Class(railTileClass("dc-server", active, s.Unread)).
		Attr("data-action", "server:select:"+s.ID).
		Attr("aria-label", serverTileLabel(s)).
		Attr("title", s.Name)
	if s.Color != "" {
		button.BackgroundColor(ui.Hex(s.Color))
	}
	if active {
		button.Attr("aria-current", "page")
	}
	if s.Mentions > 0 {
		button.Children(
			chadcn.Badge(strconv.Itoa(s.Mentions), chadcn.BadgeOptions{Variant: chadcn.BadgeDanger}).
				ID("dc-server-badge-" + s.ID).
				Position(ui.PositionAbsolute).Bottom(ui.Px(-2)).Right(ui.Px(-2)),
		)
	}
	return chadcn.Tooltip(chadcn.TooltipOptions{
		ID: "dc-server-tip-" + s.ID, CSSHover: true, Content: s.Name, Side: chadcn.PopupRight,
	}, button)
}

// addTile opens the create-server flow; settingsTile opens settings.
func (a *app) addTile() *ui.Element {
	button := ui.Button("+").ID("dc-server-add").Class("dc-server").
		Attr("data-action", "server:create").
		Attr("aria-label", "Add a server").
		Attr("title", "Add a server")
	return chadcn.Tooltip(chadcn.TooltipOptions{
		ID: "dc-server-tip-add", CSSHover: true, Content: "Add a server", Side: chadcn.PopupRight,
	}, button)
}

func (a *app) settingsTile() *ui.Element {
	button := ui.Button("\u2699").ID("dc-server-settings").Class("dc-server").
		Attr("data-action", "settings:open").
		Attr("aria-label", "User settings").
		Attr("title", "User settings")
	return chadcn.Tooltip(chadcn.TooltipOptions{
		ID: "dc-server-tip-settings", CSSHover: true, Content: "User settings", Side: chadcn.PopupRight,
	}, button)
}

func railTileClass(base string, active, unread bool) string {
	if active {
		base += " dc-server--active"
	}
	if unread {
		base += " dc-server--unread"
	}
	return base
}

func serverTileLabel(s server) string {
	switch {
	case s.Mentions == 1:
		return s.Name + ", 1 mention"
	case s.Mentions > 1:
		return s.Name + ", " + strconv.Itoa(s.Mentions) + " mentions"
	case s.Unread:
		return s.Name + ", unread"
	default:
		return s.Name
	}
}

// sidebar is the 240px channel column: guild header, collapsible channel
// categories with text and voice rows, and the footer user bar.
func (a *app) sidebar() *ui.Element {
	return ui.Tag("aside").ID("dc-side").Class("dc-side").Attr("aria-label", "Channels").Children(
		a.guildHeader(),
		a.channelList(),
		a.userBar(),
	)
}

// guildHeader names the active server (or the DM view on "home") and opens
// the server menu.
func (a *app) guildHeader() *ui.Element {
	name := "Direct Messages"
	if a.serverID != "home" {
		name = a.serverByID(a.serverID).Name
	}
	return ui.Button(name).ID("dc-guild").Class("dc-guild").
		Attr("data-action", "server:menu").
		Attr("aria-label", "Server menu for "+name).
		Attr("aria-haspopup", "menu").
		Children(
			ui.Span().Class("dc-guild-chev").Attr("aria-hidden", "true").Text("\u25BE"),
		)
}

// channelList groups the active server's channels by Category in
// first-appearance order. The home view has no channel map entry and
// renders an empty list.
func (a *app) channelList() *ui.Element {
	list := ui.Div().Class("dc-channels")
	var order []string
	byCat := map[string][]dchannel{}
	for _, ch := range a.channels[a.serverID] {
		if _, ok := byCat[ch.Category]; !ok {
			order = append(order, ch.Category)
		}
		byCat[ch.Category] = append(byCat[ch.Category], ch)
	}
	for _, cat := range order {
		list.Children(a.categoryBlock(cat, byCat[cat]))
	}
	return list
}

// categoryBlock renders one collapsible category and, unless collapsed,
// its text and voice rows.
func (a *app) categoryBlock(cat string, chans []dchannel) *ui.Element {
	collapsed := a.collapsed[cat]
	chevron := "\u25BE"
	if collapsed {
		chevron = "\u25B8"
	}
	block := ui.Div().Class("dc-cat-block").Children(
		ui.Button(chevron+" "+cat).ID("dc-cat-"+cat).Class("dc-cat").
			Attr("data-action", "cat:toggle:"+cat).
			Attr("aria-expanded", strconv.FormatBool(!collapsed)),
	)
	if collapsed {
		return block
	}
	for _, ch := range chans {
		if ch.Kind == "voice" {
			block.Children(a.voiceBlock(ch))
		} else {
			block.Children(a.textRow(ch))
		}
	}
	return block
}

// textRow is one #name (or announce) row. The mention count wins the badge
// slot; unread state is the .dc-chan--unread treatment.
func (a *app) textRow(ch dchannel) *ui.Element {
	hash := "#"
	if ch.Kind == "announce" {
		hash = "»"
	}
	class := "dc-chan"
	if a.channelID == ch.ID {
		class += " dc-chan--active"
	}
	if ch.Unread {
		class += " dc-chan--unread"
	}
	row := ui.Button("").ID("dc-chan-"+ch.ID).Class(class).
		Attr("data-action", "chan:select:"+ch.ID).
		Attr("aria-label", channelRowLabel(ch)).
		Children(
			ui.Span().Class("dc-chan-hash").Attr("aria-hidden", "true").Text(hash),
			ui.Span().Class("dc-chan-label").Text(ch.Name),
		)
	if a.channelID == ch.ID {
		row.Attr("aria-current", "page")
	}
	if ch.Mentions > 0 {
		row.Children(ui.Span().Class("dc-chan-badge").Text(strconv.Itoa(ch.Mentions)))
	}
	return row
}

func channelRowLabel(ch dchannel) string {
	switch {
	case ch.Mentions == 1:
		return "#" + ch.Name + ", 1 mention"
	case ch.Mentions > 1:
		return "#" + ch.Name + ", " + strconv.Itoa(ch.Mentions) + " mentions"
	case ch.Unread:
		return "#" + ch.Name + ", unread"
	default:
		return "#" + ch.Name
	}
}

// voiceBlock is the voice channel row plus, while joined, the connection
// panel with mute, deafen, and leave controls.
func (a *app) voiceBlock(ch dchannel) *ui.Element {
	joined := a.voiceID == ch.ID
	class := "dc-chan"
	if joined {
		class += " dc-chan--active"
	}
	row := ui.Button("").ID("dc-chan-"+ch.ID).Class(class).
		Attr("data-action", "voice:join:"+ch.ID).
		Attr("aria-label", "Join voice channel "+ch.Name).
		Children(
			ui.Span().Class("dc-chan-hash").Attr("aria-hidden", "true").Text("♪"),
			ui.Span().Class("dc-chan-label").Text(ch.Name),
		)
	block := ui.Div().Children(row)
	if !joined {
		return block
	}
	return block.Children(a.voicePanel(ch))
}

// voicePanel is the joined-channel strip: connection state plus the mute,
// deafen, and leave buttons, with the occupant rows below. Icons match the
// voice.go strip. The first occupant row is always the local demo user
// (dcMeName, green presence dot via railCSS); up to two non-bot roster
// members from a.currentMembers() follow. Rows are display-only: no
// data-action. Paint lives in railCSS (examples/builder/discord/rail_css.go;
// the integrator appends railCSS to the theme sheet in app.go).
func (a *app) voicePanel(ch dchannel) *ui.Element {
	muteIcon, muteLabel := "◉", "Mute"
	if a.voiceMuted {
		muteIcon, muteLabel = "⊘", "Unmute"
	}
	deafenIcon, deafenLabel := "♪", "Deafen"
	if a.voiceDeafened {
		deafenIcon, deafenLabel = "⊘", "Undeafen"
	}
	strip := ui.Div().Class("dc-voice").
		Attr("aria-label", "Voice connected to "+ch.Name).
		Children(
			ui.Div().Class("dc-voice-info").Children(
				ui.Div().Class("dc-voice-state").Text("Voice connected"),
				ui.Div().Class("dc-voice-name").Text(ch.Name),
			),
			ui.Div().Class("dc-voice-btns").Children(
				ui.Button(muteIcon).ID("dc-voice-mute").Class("dc-iconbtn").
					Attr("data-action", "voice:mute").
					Attr("aria-label", muteLabel).
					Attr("title", muteLabel).
					Attr("aria-pressed", strconv.FormatBool(a.voiceMuted)),
				ui.Button(deafenIcon).ID("dc-voice-deafen").Class("dc-iconbtn").
					Attr("data-action", "voice:deafen").
					Attr("aria-label", deafenLabel).
					Attr("title", deafenLabel).
					Attr("aria-pressed", strconv.FormatBool(a.voiceDeafened)),
				ui.Button("\u23FB").ID("dc-voice-leave").Class("dc-iconbtn").
					Attr("data-action", "voice:leave").
					Attr("aria-label", "Leave voice channel").
					Attr("title", "Leave voice channel"),
			),
		)
	row := func(name, initials string) *ui.Element {
		return ui.Div().Class("dc-vc-user").Children(
			ui.Span().Class("dc-vc-ava").Text(initials),
			ui.Span().Class("dc-vc-name").Text(name),
		)
	}
	users := ui.Div().Class("dc-vc-users").
		Attr("aria-label", "Members in "+ch.Name).
		Children(row(dcMeName, dcMeInitials))
	n := 1
	for _, m := range a.currentMembers() {
		if m.Bot {
			continue
		}
		if n >= 3 {
			break
		}
		users.Children(row(m.Name, dcInitial(m.Name)))
		n++
	}
	return ui.Div().Children(strip, users)
}

// userBar is the sidebar footer: avatar with presence, name, voice status,
// and the self mute/deafen/settings controls.
func (a *app) userBar() *ui.Element {
	status := "Online"
	if a.voiceID != "" {
		status = "Voice connected"
	}
	return ui.Div().Class("dc-userbar").Children(
		chadcn.Avatar(chadcn.AvatarOptions{
			Initials: dcMeInitials,
			Size:     chadcn.AvatarSmall,
			Status:   chadcn.AvatarStatusOnline,
		}),
		ui.Div().Class("dc-userbar-names").Children(
			ui.Div().Class("dc-userbar-name").Text(dcMeName),
			ui.Div().Class("dc-userbar-sub").Text(status),
		),
		a.userToggle("dc-user-mute", "Mute microphone", "Unmute microphone", "voice:mute", a.voiceMuted),
		a.userToggle("dc-user-deafen", "Deafen audio", "Undeafen audio", "voice:deafen", a.voiceDeafened),
		ui.Button("\u2699").ID("dc-user-settings").Class("dc-iconbtn").
			Attr("data-action", "settings:open").
			Attr("aria-label", "Open user settings").
			Attr("title", "Open user settings"),
	)
}

// userToggle is one footer mute/deafen toggle with pressed state. Icons
// match the voice.go strip.
func (a *app) userToggle(id, offLabel, onLabel, action string, on bool) *ui.Element {
	icon, label := "◉", offLabel
	if id == "dc-user-deafen" {
		icon = "♪"
	}
	if on {
		label = onLabel
		if id == "dc-user-mute" {
			icon = "⊘"
		} else {
			icon = "⊘"
		}
	}
	return ui.Button(icon).ID(id).Class("dc-iconbtn").
		Attr("data-action", action).
		Attr("aria-label", label).
		Attr("title", label).
		Attr("aria-pressed", strconv.FormatBool(on))
}
