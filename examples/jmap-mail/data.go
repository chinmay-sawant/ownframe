package main

import "time"

// This example models a tiny subset of RFC 8621 JMAP Email data. The IDs,
// mailbox membership, and keyword flags are shaped like JMAP, but these
// records are local fixtures rather than results from a JMAP server.
type contact struct {
	Name  string
	Email string
}

type mail struct {
	ID         string
	ThreadID   string
	From       contact
	To         contact
	Subject    string
	Preview    string
	Body       string
	ReceivedAt time.Time
	MailboxIDs map[string]bool
	Keywords   map[string]bool
}

type mailbox struct {
	ID   string
	Name string
	Role string
}

func mustMailTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return t
}

func sampleMailboxes() []mailbox {
	return []mailbox{
		{ID: "inbox", Name: "Inbox", Role: "inbox"},
		{ID: "starred", Name: "Starred", Role: "virtual"},
		{ID: "sent", Name: "Sent", Role: "sent"},
		{ID: "drafts", Name: "Drafts", Role: "drafts"},
		{ID: "archive", Name: "Archive", Role: "archive"},
		{ID: "trash", Name: "Trash", Role: "trash"},
	}
}

func sampleMessages() []mail {
	return []mail{
		{
			ID: "m-101", ThreadID: "t-101",
			From:       contact{Name: "Maya Patel", Email: "maya.patel@example.com"},
			To:         contact{Name: "Alex Morgan", Email: "alex@example.com"},
			Subject:    "Design review — OwnFrame UI builder",
			Preview:    "I added notes on the navigation, spacing, and the new fluent component API.",
			Body:       "Hi Alex,\n\nI went through the latest OwnFrame prototype. The navigation hierarchy looks clean, and the builder approach could make our desktop tooling much easier to maintain.\n\nCould we review the sidebar interactions and keyboard shortcuts on Monday?\n\nThanks,\nMaya",
			ReceivedAt: mustMailTime("2026-10-31T17:25:00+05:30"),
			MailboxIDs: map[string]bool{"inbox": true},
			Keywords:   map[string]bool{"$seen": false, "$flagged": true},
		},
		{
			ID: "m-102", ThreadID: "t-102",
			From:       contact{Name: "JMAP Working Group", Email: "updates@example.net"},
			To:         contact{Name: "Alex Morgan", Email: "alex@example.com"},
			Subject:    "JMAP synchronization notes",
			Preview:    "A quick overview of Email/query, Email/get, and Email/changes for the demo.",
			Body:       "Hello,\n\nFor this example, the mailbox list and messages are local sample data. A real JMAP client would discover the session, obtain account capabilities, and use Email/query, Email/get, and changes to keep the UI synchronized.\n\nNo credentials or network connection are required for this preview.",
			ReceivedAt: mustMailTime("2026-10-31T15:12:00+05:30"),
			MailboxIDs: map[string]bool{"inbox": true},
			Keywords:   map[string]bool{"$seen": false},
		},
		{
			ID: "m-103", ThreadID: "t-103",
			From:       contact{Name: "Nora Singh", Email: "nora.singh@example.org"},
			To:         contact{Name: "Alex Morgan", Email: "alex@example.com"},
			Subject:    "Coffee and roadmap discussion",
			Preview:    "We can go over the next milestone after the team stand-up.",
			Body:       "Hey Alex,\n\nLet's catch up after the stand-up to discuss the next milestone. I've put the sketch of the mail-client layout in our notes.\n\nSee you then,\nNora",
			ReceivedAt: mustMailTime("2026-10-31T11:40:00+05:30"),
			MailboxIDs: map[string]bool{"inbox": true},
			Keywords:   map[string]bool{"$seen": true},
		},
		{
			ID: "m-104", ThreadID: "t-104",
			From:       contact{Name: "Operations", Email: "ops@example.com"},
			To:         contact{Name: "Alex Morgan", Email: "alex@example.com"},
			Subject:    "October release checklist",
			Preview:    "The CI smoke tests and cross-platform packaging review are available.",
			Body:       "Hello,\n\nThe release checklist now includes Linux, Windows, and macOS smoke tests, a package size review, and accessibility checks.\n\nPlease look through the completed items and add any final notes.",
			ReceivedAt: mustMailTime("2026-10-30T18:05:00+05:30"),
			MailboxIDs: map[string]bool{"inbox": true},
			Keywords:   map[string]bool{"$seen": true, "$flagged": true},
		},
		{
			ID: "m-105", ThreadID: "t-105",
			From:       contact{Name: "Kai Chen", Email: "kai.chen@example.net"},
			To:         contact{Name: "Alex Morgan", Email: "alex@example.com"},
			Subject:    "A small typography experiment",
			Preview:    "I've shared a palette and a few spacing values that fit the native UI.",
			Body:       "Hi Alex,\n\nFor the mail reader, consider a quiet neutral palette, generous whitespace, and typography that stays readable at laptop resolutions.\n\nI attached the proposed measurements to our design notes.\n\nKai",
			ReceivedAt: mustMailTime("2026-10-29T16:22:00+05:30"),
			MailboxIDs: map[string]bool{"inbox": true},
			Keywords:   map[string]bool{"$seen": true},
		},
		{
			ID: "m-106", ThreadID: "t-106",
			From:       contact{Name: "Alex Morgan", Email: "alex@example.com"},
			To:         contact{Name: "Maya Patel", Email: "maya.patel@example.com"},
			Subject:    "Re: Layout review",
			Preview:    "Thanks, I will bring the updated preview to the meeting.",
			Body:       "Hi Maya,\n\nThanks for the feedback! I'll bring a revised preview to the review meeting.\n\nAlex",
			ReceivedAt: mustMailTime("2026-10-29T14:15:00+05:30"),
			MailboxIDs: map[string]bool{"sent": true},
			Keywords:   map[string]bool{"$seen": true},
		},
		{
			ID: "m-107", ThreadID: "t-107",
			From:       contact{Name: "Alex Morgan", Email: "alex@example.com"},
			To:         contact{Name: "Nora Singh", Email: "nora.singh@example.org"},
			Subject:    "Draft: Mail client interaction plan",
			Preview:    "We should keep the first demo local and let the UI builder drive all views.",
			Body:       "Hi Nora,\n\nA first iteration could have folders, message selection, search, archive, and star actions using local state. Actual JMAP transport and synchronization can come later.\n\nAlex",
			ReceivedAt: mustMailTime("2026-10-28T10:00:00+05:30"),
			MailboxIDs: map[string]bool{"drafts": true},
			Keywords:   map[string]bool{"$draft": true, "$seen": true},
		},
		{
			ID: "m-108", ThreadID: "t-108",
			From:       contact{Name: "Community Updates", Email: "community@example.org"},
			To:         contact{Name: "Alex Morgan", Email: "alex@example.com"},
			Subject:    "Native Go apps showcase",
			Preview:    "Thanks for sharing a rendering-engine prototype with the community.",
			Body:       "Hi Alex,\n\nWe enjoyed the showcase. We look forward to seeing more examples of native apps created with Go.\n\nCommunity Team",
			ReceivedAt: mustMailTime("2026-10-24T09:45:00+05:30"),
			MailboxIDs: map[string]bool{"archive": true},
			Keywords:   map[string]bool{"$seen": true},
		},
	}
}
