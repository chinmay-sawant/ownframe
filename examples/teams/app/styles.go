package app

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/examples/teams/activity"
	"github.com/chinmay-sawant/ownframe/examples/teams/calendar"
	"github.com/chinmay-sawant/ownframe/examples/teams/calls"
	"github.com/chinmay-sawant/ownframe/examples/teams/channels"
	"github.com/chinmay-sawant/ownframe/examples/teams/chat"
	"github.com/chinmay-sawant/ownframe/examples/teams/files"
)

// appStyles are the shared stylesheets, in order.
var appStyles = []string{"base.css", "ui.css", "rail.css", "flyout.css"}

// styleSheets concatenates the shared styles. A menu's stylesheet rides on
// the page theme, so only the menu on screen is matched.
func styleSheets() string {
	var b strings.Builder

	for _, name := range appStyles {
		b.WriteString(file("components/" + name))
	}

	return b.String()
}

// sectionCSS returns the stylesheet of one menu.
func sectionCSS(name string) string {
	switch name {
	case "activity":
		return activity.CSS()
	case "chat":
		return chat.CSS()
	case "channels":
		return channels.CSS()
	case "calendar":
		return calendar.CSS()
	case "calls":
		return calls.CSS()
	case "files":
		return files.CSS()
	}

	return ""
}
