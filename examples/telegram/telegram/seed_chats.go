package telegram

// seedChats is the chat list, pinned first, then newest on top.
func seedChats() []Chat {
	return []Chat{
		{
			ID: "anna", Name: "Anna Petrova", Initials: "AP", Color: 0,
			Preview: "Perfect, see you at 7", Time: "12:04",
			Unread: unreadFew, Pinned: true,
		},
		{
			ID: "design", Name: "Design Team", Initials: "DT", Color: avatarColor2,
			Preview: "Mira: pushed the new icon set", Time: "11:47",
			Unread: unreadMany, Pinned: true, Group: true,
		},
		{
			ID: "max", Name: "Max Keller", Initials: "MK", Color: avatarColor3,
			Preview: "You: the build is green", Time: "10:15",
		},
		{
			ID: "mom", Name: "Mom", Initials: "M", Color: avatarColor4,
			Preview: "Call me when you're free", Time: "Yesterday",
			Unread: 1,
		},
		{
			ID: "devs", Name: "Dev Group", Initials: "DG", Color: avatarColor5,
			Preview: "Omar: merged #482", Time: "Yesterday",
			Unread: unreadMuted, Muted: true, Group: true,
		},
		{
			ID: "sofia", Name: "Sofia Lindqvist", Initials: "SL", Color: avatarColor6,
			Preview: "Sounds good 👍", Time: "Mon",
		},
		{
			ID: "alex", Name: "Alex Turner", Initials: "AT", Color: avatarColor7,
			Preview: "You: sent the file", Time: "Sun",
		},
	}
}
