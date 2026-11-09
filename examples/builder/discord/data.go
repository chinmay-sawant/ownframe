//nolint:mnd // Discord demo fixtures are literal display content: unread counts, reactions, and member lists are fixed demo values.
package main

// This file seeds the Discord-clone demo. Types server, dchannel,
// reaction, dmessage, dmember, and struct app live in app.go; only seed
// functions and lookup helpers are defined here.

func defaultServers() []server {
	return []server{
		{ID: "ownframe", Name: "OwnFrame HQ", Initial: "O", Color: "#5865F2", Unread: true, Mentions: 1},
		{ID: "rustaceans", Name: "Rustaceans", Initial: "R", Color: "#DEA584", Unread: true},
		{ID: "gophers", Name: "Go Gophers", Initial: "G", Color: "#00ADD8"},
		{ID: "design", Name: "Design Club", Initial: "D", Color: "#EB459E", Unread: true, Mentions: 2},
		{ID: "gaming", Name: "Gaming Lounge", Initial: "L", Color: "#57F287", Unread: true},
		{ID: "music", Name: "Music Lab", Initial: "M", Color: "#FEE75C"},
	}
}

func defaultChannels() map[string][]dchannel {
	chans := map[string][]dchannel{
		"ownframe": {
			{ID: "ownframe-announcements", Name: "announcements", Kind: "announce", Category: "INFORMATION", Topic: "Release notes and big news"},
			{ID: "ownframe-rules", Name: "rules", Kind: "text", Category: "INFORMATION", Topic: "Be kind, stay on topic"},
			{ID: "ownframe-dev-log", Name: "dev-log", Kind: "text", Category: "INFORMATION", Topic: "Nightly build notes", Unread: true},
			{ID: "ownframe-general", Name: "general", Kind: "text", Category: "TEXT CHANNELS", Topic: "Anything about OwnFrame", Unread: true, Mentions: 1},
			{ID: "ownframe-help", Name: "help", Kind: "text", Category: "TEXT CHANNELS", Topic: "Ask for help here", Unread: true, Mentions: 1},
			{ID: "ownframe-media", Name: "media", Kind: "text", Category: "TEXT CHANNELS", Topic: "Screenshots and clips"},
			{ID: "ownframe-lobby", Name: "Lobby", Kind: "voice", Category: "VOICE CHANNELS"},
			{ID: "ownframe-pairing", Name: "Pairing 1:1", Kind: "voice", Category: "VOICE CHANNELS"},
		},
		"rustaceans": {
			{ID: "rustaceans-announcements", Name: "announcements", Kind: "announce", Category: "INFORMATION", Topic: "Releases and events"},
			{ID: "rustaceans-rules", Name: "rules", Kind: "text", Category: "INFORMATION", Topic: "Read before posting"},
			{ID: "rustaceans-general", Name: "general", Kind: "text", Category: "TEXT CHANNELS", Topic: "Ferris fan club", Unread: true},
			{ID: "rustaceans-help", Name: "help", Kind: "text", Category: "TEXT CHANNELS", Topic: "Borrow checker support group"},
			{ID: "rustaceans-showcase", Name: "showcase", Kind: "text", Category: "TEXT CHANNELS", Topic: "Show what you built"},
			{ID: "rustaceans-lobby", Name: "Lobby", Kind: "voice", Category: "VOICE CHANNELS"},
		},
		"gophers": {
			{ID: "gophers-announcements", Name: "announcements", Kind: "announce", Category: "INFORMATION", Topic: "Go releases and meetups"},
			{ID: "gophers-general", Name: "general", Kind: "text", Category: "TEXT CHANNELS", Topic: "Gopher chat"},
			{ID: "gophers-help", Name: "help", Kind: "text", Category: "TEXT CHANNELS", Topic: "Goroutine debugging"},
			{ID: "gophers-media", Name: "media", Kind: "text", Category: "TEXT CHANNELS", Topic: "Talks and diagrams"},
			{ID: "gophers-lobby", Name: "Lobby", Kind: "voice", Category: "VOICE CHANNELS"},
			{ID: "gophers-pairing", Name: "Pairing 1:1", Kind: "voice", Category: "VOICE CHANNELS"},
		},
		"design": {
			{ID: "design-announcements", Name: "announcements", Kind: "announce", Category: "INFORMATION", Topic: "Crit nights and challenges"},
			{ID: "design-general", Name: "general", Kind: "text", Category: "TEXT CHANNELS", Topic: "Design talk", Unread: true, Mentions: 2},
			{ID: "design-feedback", Name: "feedback", Kind: "text", Category: "TEXT CHANNELS", Topic: "Post work for critique"},
			{ID: "design-media", Name: "media", Kind: "text", Category: "TEXT CHANNELS", Topic: "Moodboards and refs"},
			{ID: "design-lobby", Name: "Lobby", Kind: "voice", Category: "VOICE CHANNELS"},
		},
		"gaming": {
			{ID: "gaming-announcements", Name: "announcements", Kind: "announce", Category: "INFORMATION", Topic: "Game nights and patches"},
			{ID: "gaming-general", Name: "general", Kind: "text", Category: "TEXT CHANNELS", Topic: "Lobby chat", Unread: true},
			{ID: "gaming-media", Name: "media", Kind: "text", Category: "TEXT CHANNELS", Topic: "Clips and screenshots"},
			{ID: "gaming-lfg", Name: "looking-for-group", Kind: "text", Category: "TEXT CHANNELS", Topic: "Find a squad"},
			{ID: "gaming-lobby", Name: "Lobby", Kind: "voice", Category: "VOICE CHANNELS"},
			{ID: "gaming-squad", Name: "Squad Voice", Kind: "voice", Category: "VOICE CHANNELS"},
		},
		"music": {
			{ID: "music-announcements", Name: "announcements", Kind: "announce", Category: "INFORMATION", Topic: "Listening parties"},
			{ID: "music-general", Name: "general", Kind: "text", Category: "TEXT CHANNELS", Topic: "What are you spinning?"},
			{ID: "music-share", Name: "share", Kind: "text", Category: "TEXT CHANNELS", Topic: "Drop a track"},
			{ID: "music-party", Name: "Listening Party", Kind: "voice", Category: "VOICE CHANNELS"},
		},
	}
	for serverID, list := range chans {
		for i := range list {
			list[i].ServerID = serverID
		}
	}
	return chans
}

func defaultMessages() map[string][]dmessage {
	msgs := map[string][]dmessage{
		"ownframe-general": {
			{ID: "ownframe-general-m01", Author: "Mira", AuthorColor: "#5865F2", Time: "Today at 08:02", Text: "Morning all! v0.0.3 milestone board is up — grab a ticket if you want one."},
			{ID: "ownframe-general-m02", Author: "Theo", AuthorColor: "#57F287", Time: "Today at 08:15", Text: "I'll take the tile-cache eviction bug, been staring at it anyway.", ReplyTo: "ownframe-general-m01", Reactions: []reaction{{Emoji: "👀", Count: 3}}},
			{ID: "ownframe-general-m03", Author: "Priya", AuthorColor: "#EB459E", Time: "Today at 08:31", Text: "Heads up: the login example flickers on resize. Recording attached in #media if anyone wants a repro."},
			{ID: "ownframe-general-m04", Author: "Sam", AuthorColor: "#FEE75C", Time: "Today at 08:44", Text: "Saw that too — only when DevTools is docked. I think the blit clip rect is one frame behind.", ReplyTo: "ownframe-general-m03"},
			{ID: "ownframe-general-m05", Author: "Ada", AuthorColor: "#00ADD8", Time: "Today at 09:03", Text: "@here nightly build is green: ebiten draw path down to 1.2ms avg on the sprite scene.", Mention: true, Reactions: []reaction{{Emoji: "🎉", Count: 8}, {Emoji: "🚀", Count: 5}}},
			{ID: "ownframe-general-m06", Author: "Mira", AuthorColor: "#5865F2", Time: "Today at 09:10", Text: "Huge. That was the retained-layer promotion fix, right?", ReplyTo: "ownframe-general-m05"},
			{ID: "ownframe-general-m07", Author: "Ada", AuthorColor: "#00ADD8", Time: "Today at 09:12", Text: "Exactly — layers now survive their rest pose instead of re-baking every frame.", ReplyTo: "ownframe-general-m06", Reactions: []reaction{{Emoji: "🧠", Count: 4}}},
			{ID: "ownframe-general-m08", Author: "Bot", AuthorColor: "#ED4245", Time: "Today at 09:30", Text: "📌 Pinned: release checklist — changelog, migration notes, tag library, bump examples require.", Pinned: true},
			{ID: "ownframe-general-m09", Author: "Jon", AuthorColor: "#9B59B6", Time: "Today at 10:01", Text: "Quick poll: keep the 125ms keep-alive tick or drop to 100? Laptop fans say drop it."},
			{ID: "ownframe-general-m10", Author: "Priya", AuthorColor: "#EB459E", Time: "Today at 10:07", Text: "Keep 125. Below that the idle-park never engages and we burn battery for nothing.", ReplyTo: "ownframe-general-m09", Reactions: []reaction{{Emoji: "🔋", Count: 6}}},
			{ID: "ownframe-general-m11", Author: "Theo", AuthorColor: "#57F287", Time: "Today at 10:20", Text: "dev-log has the flame graphs if anyone wants numbers before voting."},
			{ID: "ownframe-general-m12", Author: "Sam", AuthorColor: "#FEE75C", Time: "Today at 10:41", Text: "Can someone review my AT-SPI patch? Bus registration works but the children-changed event fires twice."},
			{ID: "ownframe-general-m13", Author: "Mira", AuthorColor: "#5865F2", Time: "Today at 10:55", Text: "@Sam on it this afternoon — probably the cache re-announce on Registry acquire.", Mention: true, ReplyTo: "ownframe-general-m12"},
			{ID: "ownframe-general-m14", Author: "Lena", AuthorColor: "#E67E22", Time: "Today at 11:12", Text: "New to OwnFrame — what's the fastest way to get a window up? The login example?"},
			{ID: "ownframe-general-m15", Author: "Theo", AuthorColor: "#57F287", Time: "Today at 11:15", Text: "@Lena yep! `go run ./examples/login` then read documentation/features.md for the call shapes.", Mention: true, ReplyTo: "ownframe-general-m14"},
			{ID: "ownframe-general-m16", Author: "Lena", AuthorColor: "#E67E22", Time: "Today at 11:18", Text: "That worked in like 30 seconds, this is very cool.", ReplyTo: "ownframe-general-m15", Reactions: []reaction{{Emoji: "❤️", Count: 7}}},
			{ID: "ownframe-general-m17", Author: "Ada", AuthorColor: "#00ADD8", Time: "Today at 11:40", Text: "Reminder: motion rules — one op, one driver. Turn off the CSS switch before hand-driving ops from SetTick."},
			{ID: "ownframe-general-m18", Author: "Priya", AuthorColor: "#EB459E", Time: "Today at 11:52", Text: "Edited: crit moved to Friday 3pm (was Thursday).", Edited: true},
			{ID: "ownframe-general-m19", Author: "Jon", AuthorColor: "#9B59B6", Time: "Today at 12:05", Text: "Lunch poll: tacos vs ramen. React 🌮 or 🍜.", Reactions: []reaction{{Emoji: "🌮", Count: 5}, {Emoji: "🍜", Count: 9, Reacted: true}}},
			{ID: "ownframe-general-m20", Author: "Mira", AuthorColor: "#5865F2", Time: "Today at 12:11", Text: "Ramen wins. Ordering at 12:30, drop requests in thread.", ReplyTo: "ownframe-general-m19"},
		},
		"ownframe-dev-log": {
			{ID: "ownframe-dev-log-m01", Author: "Bot", AuthorColor: "#ED4245", Time: "Yesterday at 18:00", Text: "Build #421 passed: 212 tests green, lint clean."},
			{ID: "ownframe-dev-log-m02", Author: "Ada", AuthorColor: "#00ADD8", Time: "Yesterday at 18:20", Text: "Landed tile-grid repaint: damage now replays per 256px tile, max 24 rects per frame."},
			{ID: "ownframe-dev-log-m03", Author: "Theo", AuthorColor: "#57F287", Time: "Yesterday at 18:35", Text: "Nice — did the >3-viewport fallback (single window-sized buffer) survive review?", ReplyTo: "ownframe-dev-log-m02"},
			{ID: "ownframe-dev-log-m04", Author: "Ada", AuthorColor: "#00ADD8", Time: "Yesterday at 18:41", Text: "Yes, with the 12MP scratch warmup on first viewport frame. Raster + chunk + text caches all hot.", ReplyTo: "ownframe-dev-log-m03", Reactions: []reaction{{Emoji: "🔥", Count: 4}}},
			{ID: "ownframe-dev-log-m05", Author: "Sam", AuthorColor: "#FEE75C", Time: "Yesterday at 19:02", Text: "Shipped SetScrollWindow pinning for chat top bar + composer. Scrolled thread can't cover them now."},
			{ID: "ownframe-dev-log-m06", Author: "Mira", AuthorColor: "#5865F2", Time: "Yesterday at 19:15", Text: "Confirmed on Telegram demo — composer stays put while messages scroll under it.", ReplyTo: "ownframe-dev-log-m05"},
			{ID: "ownframe-dev-log-m07", Author: "Priya", AuthorColor: "#EB459E", Time: "Yesterday at 20:03", Text: "Question: should retained layers cover elements with opacity tracks only, or transforms too?"},
			{ID: "ownframe-dev-log-m08", Author: "Ada", AuthorColor: "#00ADD8", Time: "Yesterday at 20:11", Text: "Both — visible 2D transform or opacity track gets promoted. SVG-paint writers stay direct.", ReplyTo: "ownframe-dev-log-m07"},
			{ID: "ownframe-dev-log-m09", Author: "Jon", AuthorColor: "#9B59B6", Time: "Yesterday at 21:30", Text: "Demotion check is in: content still changing every frame demotes after 2 frames, re-promotes after 60 calm ones."},
			{ID: "ownframe-dev-log-m10", Author: "Theo", AuthorColor: "#57F287", Time: "Yesterday at 21:44", Text: "That's the anti-thrash rule. Equalizer bars will love this.", ReplyTo: "ownframe-dev-log-m09"},
			{ID: "ownframe-dev-log-m11", Author: "Bot", AuthorColor: "#ED4245", Time: "Today at 06:00", Text: "Nightly perf: p50 frame 3.1ms, p99 9.8ms, zero fallback-badge frames."},
			{ID: "ownframe-dev-log-m12", Author: "Mira", AuthorColor: "#5865F2", Time: "Today at 07:20", Text: "📌 Pinned: perf budget — p99 must stay under 16ms on the sprite scene before we tag v0.0.3.", Pinned: true},
			{ID: "ownframe-dev-log-m13", Author: "Sam", AuthorColor: "#FEE75C", Time: "Today at 07:45", Text: "WSL snap-maximize path fixed: drag to monitor top edge now maximizes via the manual workarea path."},
			{ID: "ownframe-dev-log-m14", Author: "Ada", AuthorColor: "#00ADD8", Time: "Today at 07:58", Text: "@Sam nice, was that the same code path as the double-click header toggle?", Mention: true, ReplyTo: "ownframe-dev-log-m13"},
			{ID: "ownframe-dev-log-m15", Author: "Sam", AuthorColor: "#FEE75C", Time: "Today at 08:05", Text: "Sibling path — double-click toggles, drag-to-top snaps. Both go through the workarea helper now (edited for clarity).", Edited: true, ReplyTo: "ownframe-dev-log-m14"},
		},
		"ownframe-help": {
			{ID: "ownframe-help-m01", Author: "Lena", AuthorColor: "#E67E22", Time: "Today at 11:20", Text: "@Mira SetTick callback isn't firing — do I need RetainState on? I'm on v0.0.2.", Mention: true},
			{ID: "ownframe-help-m02", Author: "Mira", AuthorColor: "#5865F2", Time: "Today at 11:33", Text: "@Lena no — SetTick runs every frame regardless. Is your window parked? It wakes on input or WantFrameAt.", Mention: true, ReplyTo: "ownframe-help-m01"},
			{ID: "ownframe-help-m03", Author: "Lena", AuthorColor: "#E67E22", Time: "Today at 11:40", Text: "That was it — I never called WantFrameIn so it parked after 5 clean frames. Thanks!"},
		},
		"ownframe-media": {
			{ID: "ownframe-media-m01", Author: "Priya", AuthorColor: "#EB459E", Time: "Today at 08:33", Text: "Resize flicker repro (DevTools docked, drag right edge)."},
			{ID: "ownframe-media-m02", Author: "Jon", AuthorColor: "#9B59B6", Time: "Today at 09:02", Text: "60fps sprite scene after the layer fix — buttery.", Reactions: []reaction{{Emoji: "🧈", Count: 6}}},
		},
		"ownframe-announcements": {
			{ID: "ownframe-announcements-m01", Author: "Mira", AuthorColor: "#5865F2", Time: "Monday at 09:00", Text: "@everyone v0.0.3-DEV is open: motion system, retained layers, and the new Telegram demo. Changelog in #dev-log.", Mention: true, Reactions: []reaction{{Emoji: "🎉", Count: 21}}},
			{ID: "ownframe-announcements-m02", Author: "Mira", AuthorColor: "#5865F2", Time: "Today at 09:05", Text: "Reminder: breaking changes allowed until 1.0 — speak now on the Page-wrap API names."},
		},
		"ownframe-rules": {
			{ID: "ownframe-rules-m01", Author: "Mira", AuthorColor: "#5865F2", Time: "Monday at 09:00", Text: "1) Be kind. 2) Stay on topic. 3) One op, one driver.", Pinned: true},
		},
		"rustaceans-general": {
			{ID: "rustaceans-general-m01", Author: "Ferris", AuthorColor: "#DEA584", Time: "Today at 09:00", Text: "Daily borrow-checker tip: clone only at the boundary, borrow inside."},
			{ID: "rustaceans-general-m02", Author: "Mara", AuthorColor: "#EB459E", Time: "Today at 09:20", Text: "My async runtime comparison post is up in #showcase — tokio vs smol numbers inside.", Reactions: []reaction{{Emoji: "🦀", Count: 9}}},
			{ID: "rustaceans-general-m03", Author: "Owen", AuthorColor: "#00ADD8", Time: "Today at 10:11", Text: "Fought the borrow checker for 2h. The checker won. Again."},
		},
		"rustaceans-help": {
			{ID: "rustaceans-help-m01", Author: "Nina", AuthorColor: "#57F287", Time: "Today at 08:12", Text: "Why does my `Send` bound fail when I wrap the future in Rc?"},
			{ID: "rustaceans-help-m02", Author: "Ferris", AuthorColor: "#DEA584", Time: "Today at 08:25", Text: "Rc is !Send — switch to Arc if it crosses threads, or keep it on one thread.", ReplyTo: "rustaceans-help-m01"},
		},
		"gophers-general": {
			{ID: "gophers-general-m01", Author: "Gus", AuthorColor: "#00ADD8", Time: "Today at 07:55", Text: "Go 1.26 range-over-func finally clicked for me. Iterators everywhere now."},
			{ID: "gophers-general-m02", Author: "Ada", AuthorColor: "#00ADD8", Time: "Today at 08:30", Text: "Reminder: share the generic utility that saved you this week in thread."},
		},
		"gophers-help": {
			{ID: "gophers-help-m01", Author: "Gus", AuthorColor: "#00ADD8", Time: "Today at 08:00", Text: "Goroutine leak? `go vet` first, pprof second, panic third."},
			{ID: "gophers-help-m02", Author: "Nadia", AuthorColor: "#FEE75C", Time: "Today at 08:40", Text: "Adding: check for unbuffered channels with no receiver before anything else.", ReplyTo: "gophers-help-m01"},
		},
		"design-general": {
			{ID: "design-general-m01", Author: "Iris", AuthorColor: "#EB459E", Time: "Today at 09:40", Text: "@here crit night moved to Friday 3pm — bring one WIP and one question.", Mention: true, Reactions: []reaction{{Emoji: "🎨", Count: 11}}},
			{ID: "design-general-m02", Author: "Ken", AuthorColor: "#FEE75C", Time: "Today at 09:55", Text: "Bringing the new sidebar. Still torn between 28% and fixed 340px under 1180.", ReplyTo: "design-general-m01"},
			{ID: "design-general-m03", Author: "Iris", AuthorColor: "#EB459E", Time: "Today at 10:02", Text: "@Ken bring both, we'll A/B it live on the call.", Mention: true, ReplyTo: "design-general-m02"},
		},
		"design-feedback": {
			{ID: "design-feedback-m01", Author: "Ken", AuthorColor: "#FEE75C", Time: "Today at 10:30", Text: "Sidebar v3 posted in #media — roast it (kindly)."},
			{ID: "design-feedback-m02", Author: "Iris", AuthorColor: "#EB459E", Time: "Today at 10:44", Text: "Contrast on the active row needs +10%. Otherwise ship it.", ReplyTo: "design-feedback-m01"},
		},
		"gaming-general": {
			{ID: "gaming-general-m01", Author: "Zoe", AuthorColor: "#57F287", Time: "Today at 20:15", Text: "Game night Friday: co-op roguelike, 4 slots, who is in? React ✅."},
			{ID: "gaming-general-m02", Author: "Max", AuthorColor: "#ED4245", Time: "Today at 20:22", Text: "In. Bringing snacks (virtual).", ReplyTo: "gaming-general-m01", Reactions: []reaction{{Emoji: "✅", Count: 4, Reacted: true}}},
		},
		"gaming-lfg": {
			{ID: "gaming-lfg-m01", Author: "Max", AuthorColor: "#ED4245", Time: "Today at 19:00", Text: "Need 1 more for ranked, diamond+, mic required."},
			{ID: "gaming-lfg-m02", Author: "Zoe", AuthorColor: "#57F287", Time: "Today at 19:12", Text: "Me! Warming up now.", ReplyTo: "gaming-lfg-m01"},
		},
		"music-general": {
			{ID: "music-general-m01", Author: "Nadia", AuthorColor: "#FEE75C", Time: "Today at 18:00", Text: "Spinning: Khruangbin all evening. Join the Listening Party voice at 9."},
			{ID: "music-general-m02", Author: "Leo", AuthorColor: "#9B59B6", Time: "Today at 18:20", Text: "Counter-offer: ambient techno set. Winner decided by emoji vote.", Reactions: []reaction{{Emoji: "🎧", Count: 5}}},
		},
	}
	for channelID, list := range msgs {
		for i := range list {
			list[i].ChannelID = channelID
		}
	}
	return msgs
}

func defaultMembers() map[string][]dmember {
	return map[string][]dmember{
		"ownframe": {
			{ID: "u-mira", Name: "Mira", Color: "#5865F2", Status: "online", Activity: "Running the milestone board", Role: "Owner"},
			{ID: "u-ada", Name: "Ada", Color: "#00ADD8", Status: "online", Activity: "Profiling the draw path", Role: "Admin"},
			{ID: "u-theo", Name: "Theo", Color: "#57F287", Status: "online", Activity: "Hunting the tile-cache bug", Role: "Mod"},
			{ID: "u-priya", Name: "Priya", Color: "#EB459E", Status: "idle", Activity: "Away — lunch", Role: "Mod"},
			{ID: "u-sam", Name: "Sam", Color: "#FEE75C", Status: "online", Activity: "Pinning the composer bar", Role: "Member"},
			{ID: "u-jon", Name: "Jon", Color: "#9B59B6", Status: "online", Activity: "Tuning the keep-alive tick", Role: "Member"},
			{ID: "u-lena", Name: "Lena", Color: "#E67E22", Status: "online", Activity: "Reading features.md", Role: "Member"},
			{ID: "u-omar", Name: "Omar", Color: "#1ABC9C", Status: "idle", Activity: "Away — coffee", Role: "Member"},
			{ID: "u-rui", Name: "Rui", Color: "#E91E63", Status: "dnd", Activity: "Do not disturb", Role: "Member"},
			{ID: "u-kim", Name: "Kim", Color: "#95A5A6", Status: "offline", Activity: "Offline", Role: "Member"},
			{ID: "u-ci", Name: "CI Bot", Color: "#ED4245", Status: "online", Activity: "Watching build #422", Role: "Bot", Bot: true},
			{ID: "u-modmail", Name: "ModMail", Color: "#7289DA", Status: "idle", Activity: "Idle", Role: "Bot", Bot: true},
		},
		"rustaceans": {
			{ID: "u-ferris", Name: "Ferris", Color: "#DEA584", Status: "online", Activity: "Crabbing", Role: "Owner"},
			{ID: "u-mara", Name: "Mara", Color: "#EB459E", Status: "online", Activity: "Benchmarking runtimes", Role: "Admin"},
			{ID: "u-owen", Name: "Owen", Color: "#00ADD8", Status: "idle", Activity: "Away — lunch", Role: "Mod"},
			{ID: "u-nina", Name: "Nina", Color: "#57F287", Status: "online", Activity: "Fighting Rc", Role: "Member"},
			{ID: "u-pete", Name: "Pete", Color: "#FEE75C", Status: "dnd", Activity: "Do not disturb", Role: "Member"},
			{ID: "u-quinn", Name: "Quinn", Color: "#9B59B6", Status: "offline", Activity: "Offline", Role: "Member"},
			{ID: "u-crates", Name: "Crates Bot", Color: "#ED4245", Status: "online", Activity: "Indexing docs.rs", Role: "Bot", Bot: true},
		},
		"gophers": {
			{ID: "u-gus", Name: "Gus", Color: "#00ADD8", Status: "online", Activity: "Ranging over funcs", Role: "Owner"},
			{ID: "u-ada", Name: "Ada", Color: "#00ADD8", Status: "online", Activity: "Reviewing generics", Role: "Admin"},
			{ID: "u-nadia", Name: "Nadia", Color: "#FEE75C", Status: "idle", Activity: "Away — lunch", Role: "Mod"},
			{ID: "u-finn", Name: "Finn", Color: "#57F287", Status: "online", Activity: "Fixing a data race", Role: "Member"},
			{ID: "u-gail", Name: "Gail", Color: "#EB459E", Status: "offline", Activity: "Offline", Role: "Member"},
			{ID: "u-hank", Name: "Hank", Color: "#95A5A6", Status: "offline", Activity: "Offline", Role: "Member"},
			{ID: "u-gobot", Name: "GoBot", Color: "#ED4245", Status: "online", Activity: "Vetting modules", Role: "Bot", Bot: true},
		},
		"design": {
			{ID: "u-iris", Name: "Iris", Color: "#EB459E", Status: "online", Activity: "Prepping crit night", Role: "Owner"},
			{ID: "u-ken", Name: "Ken", Color: "#FEE75C", Status: "dnd", Activity: "Deep in Figma", Role: "Admin"},
			{ID: "u-luna", Name: "Luna", Color: "#9B59B6", Status: "online", Activity: "Picking type pairs", Role: "Mod"},
			{ID: "u-moe", Name: "Moe", Color: "#1ABC9C", Status: "idle", Activity: "Away — lunch", Role: "Member"},
			{ID: "u-nora", Name: "Nora", Color: "#E67E22", Status: "online", Activity: "Building a moodboard", Role: "Member"},
			{ID: "u-otto", Name: "Otto", Color: "#95A5A6", Status: "offline", Activity: "Offline", Role: "Member"},
		},
		"gaming": {
			{ID: "u-zoe", Name: "Zoe", Color: "#57F287", Status: "online", Activity: "Hosting game night", Role: "Owner"},
			{ID: "u-max", Name: "Max", Color: "#ED4245", Status: "online", Activity: "Warming up for ranked", Role: "Admin"},
			{ID: "u-ivy", Name: "Ivy", Color: "#9B59B6", Status: "idle", Activity: "Away — dinner", Role: "Mod"},
			{ID: "u-jay", Name: "Jay", Color: "#00ADD8", Status: "online", Activity: "Clipping highlights", Role: "Member"},
			{ID: "u-kai", Name: "Kai", Color: "#FEE75C", Status: "dnd", Activity: "In a match", Role: "Member"},
			{ID: "u-liv", Name: "Liv", Color: "#EB459E", Status: "offline", Activity: "Offline", Role: "Member"},
			{ID: "u-ryan", Name: "Ryan", Color: "#1ABC9C", Status: "offline", Activity: "Offline", Role: "Member"},
			{ID: "u-lfg", Name: "LFG Bot", Color: "#ED4245", Status: "online", Activity: "Matching squads", Role: "Bot", Bot: true},
		},
		"music": {
			{ID: "u-nadia", Name: "Nadia", Color: "#FEE75C", Status: "online", Activity: "Spinning Khruangbin", Role: "Owner"},
			{ID: "u-leo", Name: "Leo", Color: "#9B59B6", Status: "idle", Activity: "Queuing ambient techno", Role: "Admin"},
			{ID: "u-mia", Name: "Mia", Color: "#EB459E", Status: "online", Activity: "Listening along", Role: "Member"},
			{ID: "u-nate", Name: "Nate", Color: "#00ADD8", Status: "offline", Activity: "Offline", Role: "Member"},
			{ID: "u-dj", Name: "DJ Bot", Color: "#ED4245", Status: "online", Activity: "On the decks", Role: "Bot", Bot: true},
		},
	}
}

// unreadCounts keeps the per-channel badge numbers shown in the sidebar.
// Keys stay in sync with the Unread/Mentions flags in defaultChannels.
func unreadCounts() map[string]int {
	return map[string]int{
		"ownframe-general":   3,
		"ownframe-dev-log":   3,
		"ownframe-help":      1,
		"rustaceans-general": 5,
		"design-general":     12,
		"gaming-general":     7,
	}
}

func seedData(a *app) {
	a.servers = defaultServers()
	a.channels = defaultChannels()
	a.messages = defaultMessages()
	a.members = defaultMembers()
	a.drafts = map[string]string{}
	a.collapsed = map[string]bool{}
	a.unread = unreadCounts()
	if len(a.servers) == 0 {
		return
	}
	a.serverID = a.servers[0].ID
	a.channelID = ""
	for _, ch := range a.channels[a.serverID] {
		if ch.Name == "general" && (ch.Kind == "text" || ch.Kind == "") {
			a.channelID = ch.ID
			break
		}
	}
	if a.channelID == "" {
		for _, ch := range a.channels[a.serverID] {
			if ch.Kind != "voice" {
				a.channelID = ch.ID
				break
			}
		}
	}
}

func (a *app) currentChannels() []dchannel {
	if a == nil {
		return nil
	}
	if list, ok := a.channels[a.serverID]; ok && len(list) > 0 {
		return list
	}
	for _, list := range a.channels {
		if len(list) > 0 {
			return list
		}
	}
	return nil
}

func (a *app) currentMessages() []dmessage {
	if a == nil {
		return nil
	}
	if msgs, ok := a.messages[a.channelID]; ok {
		return msgs
	}
	return nil
}

func (a *app) currentMembers() []dmember {
	if a == nil {
		return nil
	}
	if list, ok := a.members[a.serverID]; ok {
		return list
	}
	return nil
}

func (a *app) currentChannel() dchannel {
	if a == nil {
		return dchannel{}
	}
	return a.channelByID(a.serverID, a.channelID)
}

func (a *app) serverByID(id string) server {
	if a != nil {
		for _, s := range a.servers {
			if s.ID == id {
				return s
			}
		}
		if len(a.servers) > 0 {
			return a.servers[0]
		}
	}
	for _, s := range defaultServers() {
		if s.ID == id {
			return s
		}
	}
	def := defaultServers()
	if len(def) > 0 {
		return def[0]
	}
	return server{}
}

func (a *app) channelByID(serverID, channelID string) dchannel {
	lookup := func(list []dchannel) (dchannel, bool) {
		for _, ch := range list {
			if ch.ID == channelID {
				return ch, true
			}
		}
		return dchannel{}, false
	}
	if a != nil {
		if list, ok := a.channels[serverID]; ok {
			if ch, found := lookup(list); found {
				return ch
			}
		}
		for _, list := range a.channels {
			if ch, found := lookup(list); found {
				return ch
			}
		}
		if list := a.currentChannels(); len(list) > 0 {
			return list[0]
		}
	}
	def := defaultChannels()
	if list, ok := def[serverID]; ok {
		if ch, found := lookup(list); found {
			return ch
		}
	}
	for _, list := range def {
		if ch, found := lookup(list); found {
			return ch
		}
		if len(list) > 0 {
			return list[0]
		}
	}
	return dchannel{}
}
