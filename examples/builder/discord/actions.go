package main

import (
	"context"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// msgReactParts is the id:emoji pair in a msg:react: action.
const msgReactParts = 2

// handleClick dispatches every data-action in the discord demo. Panels only
// emit action strings; all state lives on app and every change goes through
// refresh, so the screen stays a pure function of that state.
func (a *app) handleClick(ctx context.Context, box ownframe.Box) error {
	if a.page == nil {
		return nil
	}
	// findMessage locates a message by ID across every channel. It returns
	// the channel ID and the slice index, or -1 when unknown.
	findMessage := func(id string) (string, int) {
		for channelID, list := range a.messages {
			for i := range list {
				if list[i].ID == id {
					return channelID, i
				}
			}
		}
		return "", -1
	}
	// selectServer switches server, drops channel-scoped UI, and opens the
	// server's first text channel. "home" and unknown IDs fall back to the
	// first server; with no servers there is nothing to select.
	selectServer := func(ctx context.Context, id string) error {
		if id == "home" {
			if len(a.servers) == 0 {
				return nil
			}
			id = a.servers[0].ID
		}
		known := false
		for _, s := range a.servers {
			if s.ID == id {
				known = true
				break
			}
		}
		if !known {
			return nil
		}
		a.serverID = id
		a.channelID = ""
		for _, ch := range a.channels[id] {
			if ch.Kind != "voice" {
				a.channelID = ch.ID
				break
			}
		}
		a.msgMenu = ""
		a.replyTo = ""
		a.page.ScrollTo(0, 0)
		return a.refresh(ctx)
	}
	// selectChannel switches channel (and server when the channel belongs
	// to another server), clears its unread state, restores the draft into
	// the composer, and drops message-scoped UI.
	selectChannel := func(ctx context.Context, id string) error {
		serverID := ""
		for sid, list := range a.channels {
			for i := range list {
				if list[i].ID == id {
					serverID = sid
					a.channels[sid][i].Unread = false
					a.channels[sid][i].Mentions = 0
				}
			}
		}
		if serverID == "" {
			return nil
		}
		a.serverID = serverID
		a.channelID = id
		delete(a.unread, id)
		a.msgMenu = ""
		a.replyTo = ""
		if a.drafts == nil {
			a.drafts = map[string]string{}
		}
		a.page.Form().SetFormValue("dc-composer", a.drafts[id])
		return a.refresh(ctx)
	}
	// sendComposer posts the composer text to the active channel. Empty
	// text and no active channel send nothing.
	sendComposer := func(ctx context.Context) error {
		if a.channelID == "" {
			return nil
		}
		text := strings.TrimSpace(a.page.Form().FormValue("dc-composer"))
		if text == "" {
			return nil
		}
		now := time.Now()
		a.appendMessage(a.channelID, dmessage{
			ID:          "local-" + strings.ReplaceAll(now.Format("20060102150405.000000000"), ".", ""),
			ChannelID:   a.channelID,
			Author:      "You",
			AuthorColor: "#5865F2",
			Time:        "Today at " + now.Format("15:04"),
			Text:        text,
			ReplyTo:     a.replyTo,
		})
		a.page.Form().SetFormValue("dc-composer", "")
		if a.drafts != nil {
			delete(a.drafts, a.channelID)
		}
		a.replyTo = ""
		return a.refresh(ctx)
	}
	// confirmCreateServer builds a server from the dc-new-server form. An
	// empty name keeps the dialog open; the new server seeds a general
	// text channel and a lobby voice channel and becomes active.
	confirmCreate := func(ctx context.Context) error {
		name := strings.TrimSpace(a.page.Form().FormValue("dc-new-server"))
		if name == "" {
			return nil
		}
		var slug strings.Builder
		for _, r := range strings.ToLower(name) {
			switch {
			case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
				slug.WriteRune(r)
			default:
				if s := slug.String(); s != "" && !strings.HasSuffix(s, "-") {
					slug.WriteByte('-')
				}
			}
		}
		id := strings.Trim(slug.String(), "-")
		if id == "" {
			id = "server"
		}
		for _, s := range a.servers {
			if s.ID == id {
				id += "-" + time.Now().Format("150405")
				break
			}
		}
		initial := "?"
		for _, r := range name {
			initial = strings.ToUpper(string(r))
			break
		}
		palette := []string{"#5865F2", "#57F287", "#EB459E", "#00ADD8", "#FEE75C", "#9B59B6", "#E67E22"}
		if a.channels == nil {
			a.channels = map[string][]dchannel{}
		}
		general := id + "-general"
		a.servers = append(a.servers, server{
			ID:      id,
			Name:    name,
			Initial: initial,
			Color:   palette[len(a.servers)%len(palette)],
		})
		a.channels[id] = []dchannel{
			{ID: general, ServerID: id, Name: "general", Kind: "text", Category: "TEXT CHANNELS", Topic: "Anything and everything"},
			{ID: id + "-lobby", ServerID: id, Name: "Lobby", Kind: "voice", Category: "VOICE CHANNELS"},
		}
		a.serverID = id
		a.channelID = general
		a.createOpen = false
		a.msgMenu = ""
		a.replyTo = ""
		a.page.Form().SetFormValue("dc-new-server", "")
		a.page.ScrollTo(0, 0)
		return a.showToast(ctx, "Server created", name+" is ready. Start chatting in #general.")
	}
	switch {
	case box.Action == "server:create":
		a.createOpen = true
		a.page.Form().RequestFocus("dc-new-server")
		return a.refresh(ctx)
	case box.Action == "server:cancel-create":
		a.createOpen = false
		return a.refresh(ctx)
	case box.Action == "server:confirm-create":
		return confirmCreate(ctx)
	case box.Action == "server:menu":
		return a.showToast(ctx, "Server menu", "Server settings are not in this demo.")
	case strings.HasPrefix(box.Action, "server:select:"):
		return selectServer(ctx, strings.TrimPrefix(box.Action, "server:select:"))
	case strings.HasPrefix(box.Action, "nav:"):
		return selectServer(ctx, strings.TrimPrefix(box.Action, "nav:"))
	case strings.HasPrefix(box.Action, "chan:select:"):
		return selectChannel(ctx, strings.TrimPrefix(box.Action, "chan:select:"))
	case strings.HasPrefix(box.Action, "cat:toggle:"):
		cat := strings.TrimPrefix(box.Action, "cat:toggle:")
		if cat == "" {
			return nil
		}
		if a.collapsed == nil {
			a.collapsed = map[string]bool{}
		}
		a.collapsed[cat] = !a.collapsed[cat]
		return a.refresh(ctx)
	case box.Action == "search:open":
		a.searchOpen = true
		a.page.Form().RequestFocus("dc-search")
		return a.refresh(ctx)
	case box.Action == "search:close":
		a.searchOpen = false
		a.searchQuery = ""
		a.page.Form().SetFormValue("dc-search", "")
		return a.refresh(ctx)
	case box.Action == "settings:open":
		a.settingsOpen = true
		return a.refresh(ctx)
	case box.Action == "settings:close":
		a.settingsOpen = false
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "settings:tab:"):
		a.settingsTab = strings.TrimPrefix(box.Action, "settings:tab:")
		return a.refresh(ctx)
	case box.Action == "settings:save":
		return a.showToast(ctx, "Settings saved", "Your preferences apply to this demo session.")
	case box.Action == "settings:reset":
		return a.showToast(ctx, "Settings reset", "Defaults restored for this demo session.")
	case box.Action == "toast:close":
		a.toastOpen = false
		a.toastTitle = ""
		a.toastBody = ""
		return a.refresh(ctx)
	case box.Action == "overlay:dismiss":
		if a.closeOverlays() {
			return a.refresh(ctx)
		}
		return nil
	case box.Action == "msg:send":
		return sendComposer(ctx)
	case strings.HasPrefix(box.Action, "msg:react:"):
		rest := strings.TrimPrefix(box.Action, "msg:react:")
		parts := strings.SplitN(rest, ":", msgReactParts)
		if len(parts) != msgReactParts || parts[0] == "" || parts[1] == "" {
			return nil
		}
		channelID, idx := findMessage(parts[0])
		if idx < 0 {
			return nil
		}
		emoji := parts[1]
		reactions := &a.messages[channelID][idx].Reactions
		toggled := false
		for i := range *reactions {
			if (*reactions)[i].Emoji != emoji {
				continue
			}
			if (*reactions)[i].Reacted {
				(*reactions)[i].Reacted = false
				(*reactions)[i].Count--
				if (*reactions)[i].Count <= 0 {
					*reactions = append((*reactions)[:i], (*reactions)[i+1:]...)
				}
			} else {
				(*reactions)[i].Reacted = true
				(*reactions)[i].Count++
			}
			toggled = true
			break
		}
		if !toggled {
			*reactions = append(*reactions, reaction{Emoji: emoji, Count: 1, Reacted: true})
		}
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "msg:reply:"):
		id := strings.TrimPrefix(box.Action, "msg:reply:")
		if _, idx := findMessage(id); idx < 0 {
			return nil
		}
		a.replyTo = id
		a.msgMenu = ""
		a.page.Form().RequestFocus("dc-composer")
		return a.refresh(ctx)
	case box.Action == "msg:reply-clear":
		a.replyTo = ""
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "msg:pin:"):
		id := strings.TrimPrefix(box.Action, "msg:pin:")
		channelID, idx := findMessage(id)
		if idx < 0 {
			return nil
		}
		a.messages[channelID][idx].Pinned = !a.messages[channelID][idx].Pinned
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "msg:copy:"):
		id := strings.TrimPrefix(box.Action, "msg:copy:")
		if _, idx := findMessage(id); idx < 0 {
			return nil
		}
		a.msgMenu = ""
		return a.showToast(ctx, "Copied", "Message text copied to the demo clipboard.")
	case strings.HasPrefix(box.Action, "msg:delete:"):
		id := strings.TrimPrefix(box.Action, "msg:delete:")
		channelID, idx := findMessage(id)
		if idx < 0 {
			return nil
		}
		list := a.messages[channelID]
		a.messages[channelID] = append(list[:idx], list[idx+1:]...)
		if a.replyTo == id {
			a.replyTo = ""
		}
		if a.msgMenu == id {
			a.msgMenu = ""
		}
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "msg:jump:"):
		id := strings.TrimPrefix(box.Action, "msg:jump:")
		channelID, idx := findMessage(id)
		if idx < 0 {
			return nil
		}
		for sid, list := range a.channels {
			for _, ch := range list {
				if ch.ID == channelID {
					a.serverID = sid
				}
			}
		}
		a.channelID = channelID
		a.searchOpen = false
		a.msgMenu = ""
		a.page.ScrollTo(0, 0)
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "msg:menu:"):
		id := strings.TrimPrefix(box.Action, "msg:menu:")
		if id == "" {
			return nil
		}
		if a.msgMenu == id {
			a.msgMenu = ""
		} else {
			a.msgMenu = id
		}
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "member:open:"):
		a.profileOpen = strings.TrimPrefix(box.Action, "member:open:")
		return a.refresh(ctx)
	case box.Action == "member:close":
		a.profileOpen = ""
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "member:msg:"):
		id := strings.TrimPrefix(box.Action, "member:msg:")
		if id == "" {
			return nil
		}
		name := id
		for _, list := range a.members {
			for _, m := range list {
				if m.ID == id {
					name = m.Name
				}
			}
		}
		a.profileOpen = ""
		return a.showToast(ctx, "Direct messages", "DMs with "+name+" are not in this demo.")
	case box.Action == "members:toggle":
		a.memberHidden = !a.memberHidden
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "voice:join:"):
		id := strings.TrimPrefix(box.Action, "voice:join:")
		if strings.TrimSpace(id) == "" {
			return nil
		}
		a.voiceID = id
		a.voiceMuted = false
		a.voiceDeafened = false
		return a.refresh(ctx)
	case box.Action == "voice:leave":
		a.voiceID = ""
		a.voiceMuted = false
		a.voiceDeafened = false
		return a.refresh(ctx)
	case box.Action == "voice:mute":
		a.voiceMuted = !a.voiceMuted
		if !a.voiceMuted {
			a.voiceDeafened = false
		}
		return a.refresh(ctx)
	case box.Action == "voice:deafen":
		a.voiceDeafened = !a.voiceDeafened
		if a.voiceDeafened {
			a.voiceMuted = true
		}
		return a.refresh(ctx)
	case box.Action == "win:min":
		a.page.Window().Minimize()
		return nil
	case box.Action == "win:max":
		a.page.Window().ToggleMaximize()
		return nil
	case box.Action == "win:close":
		a.page.Window().Close()
		return nil
	case strings.HasPrefix(box.Action, "flag:"):
		key := strings.TrimSpace(strings.TrimPrefix(box.Action, "flag:"))
		if key == "" {
			return nil
		}
		if a.flags == nil {
			a.flags = map[string]bool{}
		}
		a.flags[key] = !a.flags[key]
		return a.refresh(ctx)
	case strings.HasPrefix(box.Action, "user:"):
		label := strings.ReplaceAll(strings.TrimSpace(strings.TrimPrefix(box.Action, "user:")), "-", " ")
		if label == "" {
			label = "This control"
		}
		return a.showToast(ctx, "Not in this demo", label+" is a stub in this demo.")
	default:
		return nil
	}
}

// handleChange mirrors the composer draft, the search query, and nothing
// else. The create-server input reads on confirm, so it is ignored here.
func (a *app) handleChange(ctx context.Context, box ownframe.Box) error {
	if a.page == nil {
		return nil
	}
	switch box.ID {
	case "dc-composer":
		if a.channelID == "" {
			return nil
		}
		if a.drafts == nil {
			a.drafts = map[string]string{}
		}
		a.drafts[a.channelID] = a.page.Form().FormValue(box.ID)
		return a.refresh(ctx)
	case "dc-search":
		a.searchQuery = a.page.Form().FormValue(box.ID)
		return a.refresh(ctx)
	default:
		return nil
	}
}

// handleKey closes overlays on escape and sends the composer on enter. Other
// keys use engine behavior only.
func (a *app) handleKey(ctx context.Context, key string) error {
	if a.page == nil {
		return nil
	}
	switch key {
	case "escape":
		if a.closeOverlays() {
			return a.refresh(ctx)
		}
		return nil
	case "enter":
		if a.page.Form().FocusID() == "dc-composer" {
			return a.handleClick(ctx, ownframe.Box{Action: "msg:send"})
		}
		return nil
	default:
		return nil
	}
}

// handleHover reports no change: hover states in this demo are CSS only.
func (a *app) handleHover(_ context.Context, _, _ ownframe.Box) (bool, error) {
	return false, nil
}

// showToast opens the transient notification. It closes through toast:close
// or overlay:dismiss.
func (a *app) showToast(ctx context.Context, title, body string) error {
	a.toastOpen = true
	a.toastTitle = title
	a.toastBody = body
	return a.refresh(ctx)
}

// appendMessage posts msg to a channel. An empty channel ID posts nowhere.
func (a *app) appendMessage(channelID string, msg dmessage) {
	if channelID == "" {
		return
	}
	if a.messages == nil {
		a.messages = map[string][]dmessage{}
	}
	msg.ChannelID = channelID
	a.messages[channelID] = append(a.messages[channelID], msg)
}

// closeOverlays clears every popup: search, profile, message menu, create
// dialog, settings, and toast. It reports whether anything was open.
func (a *app) closeOverlays() bool {
	changed := a.searchOpen || a.profileOpen != "" || a.msgMenu != "" ||
		a.createOpen || a.settingsOpen || a.toastOpen
	a.searchOpen = false
	a.searchQuery = ""
	a.profileOpen = ""
	a.msgMenu = ""
	a.createOpen = false
	a.settingsOpen = false
	a.toastOpen = false
	a.toastTitle = ""
	a.toastBody = ""
	if a.page != nil {
		a.page.Form().SetFormValue("dc-search", "")
	}
	return changed
}
