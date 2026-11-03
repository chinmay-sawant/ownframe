package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// AvatarSize selects the avatar diameter. The zero value is medium.
type AvatarSize uint8

const (
	// AvatarMedium is the default 40px avatar.
	AvatarMedium AvatarSize = iota
	// AvatarSmall is the 32px avatar.
	AvatarSmall
	// AvatarLarge is the 56px avatar.
	AvatarLarge
)

// AvatarStatus selects an optional presence dot. The zero value hides it.
type AvatarStatus uint8

const (
	// AvatarStatusNone hides the presence dot.
	AvatarStatusNone AvatarStatus = iota
	// AvatarStatusOnline marks the person as online.
	AvatarStatusOnline
	// AvatarStatusAway marks the person as away.
	AvatarStatusAway
	// AvatarStatusBusy marks the person as busy.
	AvatarStatusBusy
	// AvatarStatusOffline marks the person as offline.
	AvatarStatusOffline
)

// AvatarOptions describes the image, fallback text, size, and presence dot.
// Source is an image name registered on the page with SetImage; the
// constructor never reads the filesystem.
type AvatarOptions struct {
	Source   string
	Initials string
	Fallback string
	Alt      string
	Size     AvatarSize
	Status   AvatarStatus
}

// Avatar shows a page image, initials, or a fallback character inside a
// circle. With no source, the fallback is the initials, then Fallback, then
// "?".
func Avatar(options AvatarOptions) *ui.Element {
	size := "of-avatar--medium"
	switch options.Size {
	case AvatarSmall:
		size = "of-avatar--small"
	case AvatarLarge:
		size = "of-avatar--large"
	}

	frame := ui.Span().Class("of-avatar-frame")
	accessibleName := options.Alt
	if accessibleName == "" {
		accessibleName = options.Initials
	}
	if options.Source != "" {
		frame.Children(ui.Tag("img").Class("of-avatar-image").
			Attr("src", options.Source).Attr("alt", accessibleName))
	} else {
		text := options.Initials
		if text == "" {
			text = options.Fallback
		}
		if text == "" {
			text = "?"
		}
		if accessibleName == "" {
			accessibleName = text
		}
		frame.Children(ui.Span().Class("of-avatar-fallback").
			Attr("aria-hidden", ariaBool(true)).Text(text))
	}

	root := ui.Span().Class(classes("of-avatar", size)).Children(frame)
	if options.Source == "" {
		root.Attr("role", "img").Attr("aria-label", accessibleName)
	}
	if options.Status != AvatarStatusNone {
		root.Children(ui.Span().Class(classes("of-avatar-status", avatarStatusClass(options.Status))).
			Children(ui.Span().Class("of-visually-hidden").Text(avatarStatusLabel(options.Status))))
	}
	return root
}

func avatarStatusClass(status AvatarStatus) string {
	switch status {
	case AvatarStatusOnline:
		return "of-avatar-status--online"
	case AvatarStatusAway:
		return "of-avatar-status--away"
	case AvatarStatusBusy:
		return "of-avatar-status--busy"
	case AvatarStatusOffline:
		return "of-avatar-status--offline"
	default:
		return "of-avatar-status--online"
	}
}

func avatarStatusLabel(status AvatarStatus) string {
	switch status {
	case AvatarStatusOnline:
		return "Online"
	case AvatarStatusAway:
		return "Away"
	case AvatarStatusBusy:
		return "Busy"
	case AvatarStatusOffline:
		return "Offline"
	default:
		return "Online"
	}
}
