package main

import (
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
