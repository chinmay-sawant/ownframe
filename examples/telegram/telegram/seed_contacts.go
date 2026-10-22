package telegram

// seedContacts is the contacts tab. A contact with a ChatID opens that
// thread; an empty ChatID starts a new one.
func seedContacts() []Contact {
	return []Contact{
		{ID: "anna", Name: "Anna Petrova", Initials: "AP", Color: 0,
			Status: "online", ChatID: "anna"},
		{ID: "max", Name: "Max Keller", Initials: "MK", Color: avatarColor3,
			Status: "last seen 2 minutes ago", ChatID: "max"},
		{ID: "sofia", Name: "Sofia Lindqvist", Initials: "SL", Color: avatarColor6,
			Status: "last seen recently", ChatID: "sofia"},
		{ID: "alex", Name: "Alex Turner", Initials: "AT", Color: avatarColor7,
			Status: "last seen yesterday at 21:14", ChatID: "alex"},
		{ID: "nina", Name: "Nina Rossi", Initials: "NR", Color: 1,
			Status: "last seen 4 hours ago"},
		{ID: "omar", Name: "Omar Haddad", Initials: "OH", Color: avatarColor5,
			Status: "last seen recently"},
		{ID: "priya", Name: "Priya Nair", Initials: "PN", Color: avatarColor2,
			Status: "last seen 11 minutes ago"},
		{ID: "jonas", Name: "Jonas Weber", Initials: "JW", Color: avatarColor4,
			Status: "last seen yesterday"},
		{ID: "kate", Name: "Kate Morgan", Initials: "KM", Color: avatarColor6,
			Status: "last seen a long time ago"},
		{ID: "leo", Name: "Leo Fontaine", Initials: "LF", Color: avatarColor3,
			Status: "last seen 3 days ago"},
	}
}
