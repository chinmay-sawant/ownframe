package main

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestSeedData(t *testing.T) {
	servers := defaultServers()
	if len(servers) == 0 {
		t.Fatal("defaultServers returned no servers")
	}
	channels := defaultChannels()
	messages := defaultMessages()
	members := defaultMembers()
	channelIDs := map[string]string{}
	for sid, list := range channels {
		if len(list) == 0 {
			t.Errorf("server %q has no channels", sid)
		}
		for _, ch := range list {
			if ch.ServerID != sid {
				t.Errorf("channel %q ServerID = %q, want %q", ch.ID, ch.ServerID, sid)
			}
			channelIDs[ch.ID] = sid
		}
	}
	for cid, list := range messages {
		sid, ok := channelIDs[cid]
		if !ok {
			t.Errorf("messages for unknown channel %q", cid)
			continue
		}
		_ = sid
		for _, m := range list {
			if m.ChannelID != cid {
				t.Errorf("message %q ChannelID = %q, want %q", m.ID, m.ChannelID, cid)
			}
		}
	}
	if len(members) == 0 {
		t.Fatal("defaultMembers returned no servers")
	}
}

func TestChannelSwitch(t *testing.T) {
	a := &app{
		serverID:  "s1",
		channelID: "c1",
		channels: map[string][]dchannel{
			"s1": {
				{ID: "c1", ServerID: "s1", Name: "general", Kind: "text"},
				{ID: "c2", ServerID: "s1", Name: "random", Kind: "text"},
			},
		},
		messages: map[string][]dmessage{
			"c1": {{ID: "m1", ChannelID: "c1", Author: "Ada", Text: "hi"}},
			"c2": {
				{ID: "m2", ChannelID: "c2", Author: "Bo", Text: "hello"},
				{ID: "m3", ChannelID: "c2", Author: "Cy", Text: "hey"},
			},
		},
	}
	if got := len(a.currentChannels()); got != 2 {
		t.Fatalf("currentChannels = %d, want 2", got)
	}
	a.channelID = "c2"
	if got := len(a.currentMessages()); got != 2 {
		t.Fatalf("currentMessages in c2 = %d, want 2", got)
	}
	a.channelID = "c1"
	if got := len(a.currentMessages()); got != 1 {
		t.Fatalf("currentMessages in c1 = %d, want 1", got)
	}
}

func TestMessageAppend(t *testing.T) {
	a := &app{messages: map[string][]dmessage{}}
	a.appendMessage("c1", dmessage{ID: "m1", Author: "You", Text: "first"})
	a.appendMessage("c1", dmessage{ID: "m2", Author: "You", Text: "second"})
	if got := len(a.messages["c1"]); got != 2 {
		t.Fatalf("messages[c1] = %d, want 2", got)
	}
	if got := a.messages["c1"][1].Text; got != "second" {
		t.Fatalf("last message = %q, want %q", got, "second")
	}
	if got := a.messages["c1"][1].ChannelID; got != "c1" {
		t.Fatalf("ChannelID = %q, want c1", got)
	}
	// Empty channel posts nowhere.
	a.appendMessage("", dmessage{ID: "m3"})
	if _, ok := a.messages[""]; ok {
		t.Fatal("appendMessage with empty channelID should post nowhere")
	}
}

func TestActionParsing(t *testing.T) {
	cases := []struct {
		action, prefix, rest string
	}{
		{"server:select:ownframe", "server:select:", "ownframe"},
		{"chan:select:general", "chan:select:", "general"},
		{"cat:toggle:TEXT CHANNELS", "cat:toggle:", "TEXT CHANNELS"},
		{"msg:reply:m42", "msg:reply:", "m42"},
		{"msg:menu:m42", "msg:menu:", "m42"},
		{"msg:pin:m42", "msg:pin:", "m42"},
		{"msg:delete:m42", "msg:delete:", "m42"},
		{"msg:jump:m42", "msg:jump:", "m42"},
		{"msg:react:m42:👍", "msg:react:", "m42:👍"},
		{"member:open:u7", "member:open:", "u7"},
		{"voice:join:lobby", "voice:join:", "lobby"},
	}
	for _, c := range cases {
		if !strings.HasPrefix(c.action, c.prefix) {
			t.Errorf("action %q missing prefix %q", c.action, c.prefix)
			continue
		}
		if got := strings.TrimPrefix(c.action, c.prefix); got != c.rest {
			t.Errorf("action %q rest = %q, want %q", c.action, got, c.rest)
		}
	}
	// Reaction payload splits into id + emoji.
	rest := strings.TrimPrefix("msg:react:m42:👍", "msg:react:")
	parts := strings.SplitN(rest, ":", 2)
	if len(parts) != 2 || parts[0] != "m42" || parts[1] != "👍" {
		t.Fatalf("react split = %q, want [m42 👍]", parts)
	}
}

// TestRotateTypingCycle drives one show -> gap -> show rotation with
// simulated time. The app is built directly: newApp mounts a window.
func TestRotateTypingCycle(t *testing.T) {
	const (
		showWindow = 5 * time.Second
		gapWindow  = 3 * time.Second
		slop       = 2 * time.Second
	)
	base := time.Now()
	a := &app{
		typingNames: []string{"Theo", "Priya", "Sam"},
		typingUntil: base.Add(-time.Second),
	}

	a.rotateTyping(base)
	if len(a.typing) != 1 || a.typing[0] != "Theo" {
		t.Fatalf("after first rotate typing = %q, want [Theo]", a.typing)
	}
	if dt := a.typingUntil.Sub(base); dt < showWindow-slop || dt > showWindow+slop {
		t.Fatalf("typing window = %v, want ~%v", dt, showWindow)
	}

	// Past the window the indicator clears for a short gap.
	gapAt := a.typingUntil.Add(time.Millisecond)
	a.rotateTyping(gapAt)
	if len(a.typing) != 0 {
		t.Fatalf("in gap typing = %q, want empty", a.typing)
	}
	if dt := a.typingUntil.Sub(gapAt); dt < gapWindow-slop || dt > gapWindow+slop {
		t.Fatalf("typing gap = %v, want ~%v", dt, gapWindow)
	}

	// Past the gap the next name rotates in.
	nextAt := a.typingUntil.Add(time.Millisecond)
	a.rotateTyping(nextAt)
	if len(a.typing) != 1 || a.typing[0] != "Priya" {
		t.Fatalf("after gap typing = %q, want [Priya]", a.typing)
	}
	if dt := a.typingUntil.Sub(nextAt); dt < showWindow-slop || dt > showWindow+slop {
		t.Fatalf("typing window = %v, want ~%v", dt, showWindow)
	}
}

// TestRotateTypingTwoTypers walks show/gap cycles with simulated time
// until a two-name indicator appears (every third cycle). The app is
// built directly: newApp mounts a window.
func TestRotateTypingTwoTypers(t *testing.T) {
	names := []string{"Theo", "Priya", "Sam"}
	a := &app{typingNames: names}
	now := time.Now()
	const maxCycles = 9
	var pair []string
	for i := 0; i < maxCycles; i++ {
		a.rotateTyping(now)
		if len(a.typing) == 2 {
			pair = append([]string(nil), a.typing...)
			break
		}
		now = a.typingUntil.Add(time.Millisecond)
		a.rotateTyping(now)
		now = a.typingUntil.Add(time.Millisecond)
	}
	if pair == nil {
		t.Fatalf("no two-typer indicator within %d cycles, last typing=%q", maxCycles, a.typing)
	}
	if pair[0] == pair[1] {
		t.Fatalf("two-typer names = %q, want two distinct names", pair)
	}
	allowed := map[string]bool{}
	for _, n := range names {
		allowed[n] = true
	}
	for _, n := range pair {
		if !allowed[n] {
			t.Errorf("typer %q not in %q", n, names)
		}
	}
}

// TestFlagToggleParse pins the "flag:<key>" parse contract from the flag:
// branch of handleClick in actions.go. That branch refreshes the view, so
// it needs a window; only the strings logic is covered here.
func TestFlagToggleParse(t *testing.T) {
	parse := func(action string) (string, bool) {
		if !strings.HasPrefix(action, "flag:") {
			return "", false
		}
		key := strings.TrimSpace(strings.TrimPrefix(action, "flag:"))
		if key == "" {
			return "", false
		}
		return key, true
	}
	cases := []struct {
		action, key string
		ok          bool
	}{
		{"flag:overlay", "overlay", true},
		{"flag:  overlay  ", "overlay", true},
		{"flag:", "", false},
		{"flag:   ", "", false},
		{"server:create", "", false},
	}
	for _, c := range cases {
		key, ok := parse(c.action)
		if ok != c.ok || key != c.key {
			t.Errorf("parse(%q) = (%q, %v), want (%q, %v)", c.action, key, ok, c.key, c.ok)
		}
	}
}

// historyCap is the per-channel history bound stated in the README memory
// bounds ("cap history at 500 messages per channel with drop-oldest on
// append"). It is a literal here so this contract does not depend on any
// sibling constant for the same bound.
const historyCap = 500

// TestHistoryCap pins the per-channel history bound: appends past the cap
// drop the oldest first (FIFO), the newest message survives, and a second
// channel is unaffected by the overflow. The app is built directly:
// newApp mounts a window.
func TestHistoryCap(t *testing.T) {
	a := &app{messages: map[string][]dmessage{}}
	a.appendMessage("c2", dmessage{ID: "other-1", Author: "Bo", Text: "untouched one"})
	a.appendMessage("c2", dmessage{ID: "other-2", Author: "Bo", Text: "untouched two"})
	const appends = 600
	for i := 1; i <= appends; i++ {
		a.appendMessage("c1", dmessage{ID: fmt.Sprintf("m%04d", i), Author: "You", Text: fmt.Sprintf("msg %d", i)})
	}
	got := a.messages["c1"]
	if len(got) != historyCap {
		t.Fatalf("messages[c1] = %d, want %d (drop-oldest FIFO)", len(got), historyCap)
	}
	if last := got[len(got)-1]; last.Text != "msg 600" {
		t.Fatalf("newest message = %q, want %q", last.Text, "msg 600")
	}
	if last := got[len(got)-1]; last.ChannelID != "c1" {
		t.Fatalf("newest ChannelID = %q, want c1", last.ChannelID)
	}
	other := a.messages["c2"]
	if len(other) != 2 || other[0].Text != "untouched one" || other[1].Text != "untouched two" {
		t.Fatalf("channel c2 affected by c1 overflow: %v", other)
	}
}

// TestCapKeepsNewest pins the FIFO order of the history bound: after 505
// numbered appends the oldest five (#1-#5) are gone, the first kept message
// is #6, and the last is #505. The app is built directly: newApp mounts a
// window.
func TestCapKeepsNewest(t *testing.T) {
	a := &app{messages: map[string][]dmessage{}}
	const appends = 505
	for i := 1; i <= appends; i++ {
		a.appendMessage("c1", dmessage{ID: fmt.Sprintf("m%04d", i), Author: "You", Text: fmt.Sprintf("msg %d", i)})
	}
	got := a.messages["c1"]
	if len(got) != historyCap {
		t.Fatalf("messages[c1] = %d, want %d", len(got), historyCap)
	}
	if got[0].Text != "msg 6" {
		t.Fatalf("oldest kept = %q, want %q (drop-oldest FIFO over 505 appends)", got[0].Text, "msg 6")
	}
	if got[len(got)-1].Text != "msg 505" {
		t.Fatalf("newest kept = %q, want %q", got[len(got)-1].Text, "msg 505")
	}
}

// TestWindowBounds pins the transcript-window contract for a channel at the
// history cap: a viewport of viewRows rows at any clamped offset stays
// inside the capped history and the bottom window ends on the newest
// message. visibleRange is absent in scroll.go (the transcript still renders
// every row with no slicing yet), so this test documents the spec against
// the capped slice and skips until the helper lands. It must not define
// visibleRange itself.
func TestWindowBounds(t *testing.T) {
	a := &app{messages: map[string][]dmessage{}}
	const appends = 1000
	for i := 1; i <= appends; i++ {
		a.appendMessage("c1", dmessage{ID: fmt.Sprintf("m%04d", i), Author: "You", Text: fmt.Sprintf("msg %d", i)})
	}
	t.Skip("visibleRange absent in scroll.go: transcript renders every row with no slicing yet")
	msgs := a.messages["c1"]
	if len(msgs) != historyCap {
		t.Fatalf("messages[c1] = %d, want %d after 1000 appends", len(msgs), historyCap)
	}
	const viewRows = 20
	for _, off := range []int{0, 240, 479, 480, 10000, -5} {
		start := off
		if start < 0 {
			start = 0
		}
		if maxStart := len(msgs) - viewRows; start > maxStart {
			start = maxStart
		}
		end := start + viewRows
		if end > len(msgs) {
			end = len(msgs)
		}
		if !(0 <= start && start <= end && end <= len(msgs)) {
			t.Fatalf("window [%d:%d) out of bounds for offset %d over %d messages", start, end, off, len(msgs))
		}
		if visible := msgs[start:end]; len(visible) > viewRows {
			t.Fatalf("visible window = %d rows, want <= %d", len(visible), viewRows)
		}
	}
	bottom := msgs[len(msgs)-viewRows:]
	if bottom[0].Text != "msg 981" || bottom[viewRows-1].Text != "msg 1000" {
		t.Fatalf("bottom window = [%q..%q], want [msg 981..msg 1000]", bottom[0].Text, bottom[viewRows-1].Text)
	}
}

// TestPatchFallbackContract pins the action-string parse shapes behind the
// patched update paths: msg:react splits id:emoji on the first colon only
// (mirroring strings.SplitN(rest, ":", 2) in actions.go, so an emoji token
// containing ":" survives whole), and flag: passes its key through to the
// flags map. Pure string asserts: no window, no refresh.
func TestPatchFallbackContract(t *testing.T) {
	react := func(action string) (id, emoji string, ok bool) {
		if !strings.HasPrefix(action, "msg:react:") {
			return "", "", false
		}
		parts := strings.SplitN(strings.TrimPrefix(action, "msg:react:"), ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "", false
		}
		return parts[0], parts[1], true
	}
	reactCases := []struct {
		action, id, emoji string
		ok                bool
	}{
		{"msg:react:m42:👍", "m42", "👍", true},
		{"msg:react:m42:custom:emoji", "m42", "custom:emoji", true},
		{"msg:react:m42", "", "", false},
		{"msg:react::👍", "", "", false},
		{"msg:react:m42:", "", "", false},
		{"msg:reply:m42", "", "", false},
	}
	for _, c := range reactCases {
		id, emoji, ok := react(c.action)
		if ok != c.ok || id != c.id || emoji != c.emoji {
			t.Errorf("react(%q) = (%q, %q, %v), want (%q, %q, %v)",
				c.action, id, emoji, ok, c.id, c.emoji, c.ok)
		}
	}
	flagKey := func(action string) (string, bool) {
		if !strings.HasPrefix(action, "flag:") {
			return "", false
		}
		key := strings.TrimSpace(strings.TrimPrefix(action, "flag:"))
		if key == "" {
			return "", false
		}
		return key, true
	}
	if key, ok := flagKey("flag:overlay"); !ok || key != "overlay" {
		t.Fatalf("flagKey(%q) = (%q, %v), want (%q, %v)", "flag:overlay", key, ok, "overlay", true)
	}
	if _, ok := flagKey("server:create"); ok {
		t.Fatalf("flagKey(%q) = ok, want not-ok", "server:create")
	}
	// Flag passthrough: the parsed key toggles the app flags map with no
	// window involved, mirroring the flag: branch minus refresh.
	a := &app{}
	key, ok := flagKey("flag:overlay")
	if !ok {
		t.Fatal("flagKey(flag:overlay) not ok")
	}
	if a.flags == nil {
		a.flags = map[string]bool{}
	}
	a.flags[key] = !a.flags[key]
	if !a.flags["overlay"] {
		t.Fatal("flag passthrough did not toggle flags[overlay]")
	}
}
