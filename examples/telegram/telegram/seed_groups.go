package telegram

// seedGroups fills the group chats' history. Group bubbles carry the
// author's name and color.
func seedGroups(m map[string][]Message) {
	m["design"] = []Message{
		{ID: "design-1", Text: "Pushed the new icon set, have a look when you can", Time: "10:31",
			Author: "Mira", Initials: "MI", Color: avatarColor5},
		{ID: "design-2", Text: "The rounded ones look great", Time: "10:44",
			Own: true, Read: true},
		{ID: "design-3", Text: "Can we try a lighter stroke on the back arrow?", Time: "11:02",
			Author: "Jonas", Initials: "JW", Color: avatarColor4},
		{ID: "design-4", Text: "Second that, it reads heavy on the dark theme", Time: "11:05",
			Author: "Priya", Initials: "PN", Color: avatarColor2},
		{ID: "design-5", Text: "Pushed the new icon set", Time: "11:47",
			Author: "Mira", Initials: "MI", Color: avatarColor5},
	}
	m["devs"] = []Message{
		{ID: "devs-1", Text: "The parser fix is in review", Time: "09:14",
			Author: "Omar", Initials: "OH", Color: avatarColor5},
		{ID: "devs-2", Text: "Nice, I'll take a look this morning", Time: "09:20",
			Own: true, Read: true},
		{ID: "devs-3", Text: "We can merge it today", Time: "09:36",
			Author: "Kate", Initials: "KM", Color: avatarColor6},
		{ID: "devs-4", Text: "Merged #482", Time: "13:02",
			Author: "Omar", Initials: "OH", Color: avatarColor5},
	}
}
