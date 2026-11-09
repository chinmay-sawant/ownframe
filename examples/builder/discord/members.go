// Members rail and profile popover for the discord demo.
//
// Actions emitted (handled by the host click dispatcher in another file):
// members:toggle, member:open:<id>, member:close, member:msg:<id>.
package main

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// dcMemberGroups orders the role sections in the rail.
func dcMemberGroups() []string {
	return []string{"Owner", "Admin", "Mods", "Bots", "Online", "Offline"}
}

// dcMemberGroup maps a roster row to its rail section. Bots always group
// together; every other member groups by role, with plain members split by
// presence into Online and Offline.
func dcMemberGroup(m dmember) string {
	if m.Bot {
		return "Bots"
	}
	switch m.Role {
	case "Owner":
		return "Owner"
	case "Admin":
		return "Admin"
	case "Mod":
		return "Mods"
	}
	if strings.EqualFold(strings.TrimSpace(m.Status), "offline") {
		return "Offline"
	}
	return "Online"
}

// dcAvatarStatus maps the demo presence string to a chadcn status dot.
func dcAvatarStatus(status string) chadcn.AvatarStatus {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "online":
		return chadcn.AvatarStatusOnline
	case "idle":
		return chadcn.AvatarStatusAway
	case "dnd":
		return chadcn.AvatarStatusBusy
	default:
		return chadcn.AvatarStatusOffline
	}
}

// dcStatusLabel renders the demo presence string as plain text.
func dcStatusLabel(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "online":
		return "Online"
	case "idle":
		return "Idle"
	case "dnd":
		return "Do not disturb"
	default:
		return "Offline"
	}
}

// dcInitial derives avatar initials from a display name.
func dcInitial(name string) string {
	for _, r := range name {
		return strings.ToUpper(string(r))
	}
	return "?"
}

// membersPanel is the 240px right rail grouped by role. It returns nil while
// the rail is hidden. Each row opens the profile popover for that member.
func (a *app) membersPanel() *ui.Element {
	if a.memberHidden {
		return nil
	}
	members := a.currentMembers()
	groups := dcMemberGroups()
	sections := make([]*ui.Element, 0, len(groups))
	for _, group := range groups {
		rows := make([]*ui.Element, 0, len(members))
		for _, m := range members {
			if dcMemberGroup(m) == group {
				rows = append(rows, a.dcMemberRow(m))
			}
		}
		if len(rows) == 0 {
			continue
		}
		header := ui.Text(group + " — " + strconv.Itoa(len(rows))).Class("dc-role")
		sections = append(sections, ui.Div().Class("dc-role-group").Children(
			append([]*ui.Element{header}, rows...)...,
		))
	}
	return ui.Tag("aside").Class("dc-members").
		Width(ui.Px(240)).
		Attr("aria-label", "Members").
		Children(
			ui.Div().Class("dc-members-list").Children(sections...),
		)
}

// dcMemberRow is one avatar-plus-name row that opens the profile popover.
// Bots carry a BOT tag via chadcn.Badge. In light mode the name stays dark
// text with a role-color dot, since pale role colors don't read on light
// backgrounds; in dark mode the name itself carries the role color.
func (a *app) dcMemberRow(m dmember) *ui.Element {
	name := ui.Text(m.Name).Class("dc-member-name")
	rowName := ui.Div().Class("dc-member-namerow")
	if a != nil && a.appearance == dcLight {
		if strings.TrimSpace(m.Color) != "" {
			dot := ui.Span().Class("dc-member-dot")
			dot.BackgroundColor(ui.Hex(m.Color))
			rowName.Children(dot)
		}
	} else if strings.TrimSpace(m.Color) != "" {
		name.Color(ui.Hex(m.Color))
	}
	rowName.Children(name)
	if m.Bot {
		rowName.Children(chadcn.Badge("BOT", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}))
	}
	activity := strings.TrimSpace(m.Activity)
	if activity == "" {
		activity = dcStatusLabel(m.Status)
	}
	return ui.Div().Class("dc-member").
		ID("dc-member-"+m.ID).
		Attr("data-action", "member:open:"+m.ID).
		Attr("role", "button").
		Attr("aria-label", "Open profile for "+m.Name).
		Children(
			chadcn.Avatar(chadcn.AvatarOptions{
				Initials: dcInitial(m.Name),
				Size:     chadcn.AvatarSmall,
				Status:   dcAvatarStatus(m.Status),
			}),
			ui.Div().Class("dc-member-meta").Children(
				rowName,
				ui.Text(activity).Class("dc-member-activity"),
			),
		)
}

// dcProfileMember resolves the open profile ID against the current server
// first, then every other server.
func (a *app) dcProfileMember() *dmember {
	for _, m := range a.currentMembers() {
		if m.ID == a.profileOpen {
			m := m
			return &m
		}
	}
	for _, list := range a.members {
		for _, m := range list {
			if m.ID == a.profileOpen {
				m := m
				return &m
			}
		}
	}
	return nil
}

// profileCard is the member popover, dressed like a real Discord user card:
// a role-tinted banner, a large overlapping avatar, the display name plus a
// lowercase handle, ABOUT ME activity text, a static MEMBER SINCE row, ROLES
// chips with color dots (plus the BOT tag for bots), and Message/Close
// actions. It returns nil while no profile is open. The message action opens
// a DM note via toast; both dismiss controls dispatch member:close. All are
// handled by the host click dispatcher.
//
// Position follows Discord: the card sits left of the member rail,
// vertically anchored to the clicked row and clamped to the viewport so a
// row near the bottom shifts the card up instead of clipping.
func (a *app) profileCard() *ui.Element {
	if a.profileOpen == "" {
		return nil
	}
	m := a.dcProfileMember()
	if m == nil {
		return nil
	}
	banner := ui.Div().Class("dc-profile-banner").Children(
		chadcn.Button("×", chadcn.ButtonOptions{
			Variant: chadcn.ButtonText,
			Size:    chadcn.ButtonSmall,
		}).ID("dc-profile-x").Attr("data-action", "member:close").
			Attr("aria-label", "Close profile"),
	)
	if strings.TrimSpace(m.Color) != "" {
		banner.BackgroundColor(ui.Hex(m.Color))
	}
	nameRow := ui.Div().Class("dc-profile-namerow").Children(
		ui.Text(m.Name).Class("dc-profile-name"),
	)
	if m.Bot {
		nameRow.Children(chadcn.Badge("BOT", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}))
	}
	roles := []*ui.Element{dcProfileRoleChip(m.Role, m.Color)}
	if m.Bot {
		roles = append(roles, chadcn.Badge("BOT", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}))
	}
	left, top := a.profileCardPos()
	return ui.Div().ID("dc-profile").Class("dc-profile").
		Attr("role", "dialog").
		Attr("aria-label", "Profile for "+m.Name).
		Position(ui.PositionFixed).
		Left(ui.Px(float64(left))).
		Top(ui.Px(float64(top))).
		ZIndex(ui.Number(150)).
		Children(
			banner,
			ui.Div().Class("dc-profile-ava").Children(
				chadcn.Avatar(chadcn.AvatarOptions{
					Initials: dcInitial(m.Name),
					Size:     chadcn.AvatarLarge,
					Status:   dcAvatarStatus(m.Status),
				}),
			),
			ui.Div().Class("dc-profile-body").Children(
				nameRow,
				ui.Text(dcProfileHandle(m.Name)).Class("dc-profile-user"),
				ui.Div().Class("dc-profile-sec").Children(
					ui.Text("ABOUT ME").Class("dc-profile-sechead"),
					ui.Text(dcProfileAbout(*m)).Class("dc-profile-text"),
				),
				ui.Div().Class("dc-profile-sec").Children(
					ui.Text("MEMBER SINCE").Class("dc-profile-sechead"),
					ui.Text("Nov 2026").Class("dc-profile-text"),
				),
				ui.Div().Class("dc-profile-sec").Children(
					ui.Text("ROLES").Class("dc-profile-sechead"),
					ui.Div().Class("dc-profile-roles").Children(roles...),
				),
				ui.Div().Class("dc-profile-actions").Children(
					chadcn.Button("Message", chadcn.ButtonOptions{
						Variant: chadcn.ButtonFilled,
						Size:    chadcn.ButtonSmall,
					}).ID("dc-profile-msg").Attr("data-action", "member:msg:"+m.ID),
					chadcn.Button("Close", chadcn.ButtonOptions{
						Variant: chadcn.ButtonOutlined,
						Size:    chadcn.ButtonSmall,
					}).ID("dc-profile-close").Attr("data-action", "member:close"),
				),
			),
		)
}

// profileCardPos anchors the profile popover like Discord: left of the
// member rail, top aligned to the clicked row, clamped so the card stays
// inside the viewport. Without an anchor it falls back to the old static
// spot so tests without click geometry still render.
func (a *app) profileCardPos() (int, int) {
	const cardW, cardH = 300, 480
	const membersW, gap = 240, 12
	const topMin, margin = 56, 12
	// edgeInset keeps the card clear of the viewport sides so it never
	// touches the rail or the window frame.
	const edgeInset = 8
	winW, winH := a.windowSize()
	left := winW - membersW - gap - cardW
	if left < edgeInset {
		left = edgeInset
	}
	if maxLeft := winW - cardW - edgeInset; left > maxLeft && maxLeft > edgeInset {
		left = maxLeft
	}
	top := 64
	if a != nil && a.profileAnchored {
		top = int(a.profileY)
	}
	if top < topMin {
		top = topMin
	}
	if maxTop := winH - cardH - margin; maxTop > topMin && top > maxTop {
		top = maxTop
	}
	return left, top
}

// dcProfileHandle renders the username handle: @ plus the lowercase name
// with spaces stripped.
func dcProfileHandle(name string) string {
	handle := strings.ToLower(strings.TrimSpace(name))
	handle = strings.ReplaceAll(handle, " ", "")
	if handle == "" {
		handle = "member"
	}
	return "@" + handle
}

// dcProfileAbout returns the ABOUT ME line: the live activity when set,
// otherwise the plain presence label.
func dcProfileAbout(m dmember) string {
	if activity := strings.TrimSpace(m.Activity); activity != "" {
		return activity
	}
	return dcStatusLabel(m.Status)
}

// dcProfileRoleChip is one ROLES chip: a color dot plus the role name. An
// empty role falls back to Member so the row never renders blank.
func dcProfileRoleChip(role, color string) *ui.Element {
	if strings.TrimSpace(role) == "" {
		role = "Member"
	}
	dot := ui.Span().Class("dc-profile-dot")
	if strings.TrimSpace(color) != "" {
		dot.BackgroundColor(ui.Hex(color))
	}
	return ui.Div().Class("dc-profile-rolechip").Children(
		dot,
		ui.Text(role).Class("dc-profile-chiptext"),
	)
}
