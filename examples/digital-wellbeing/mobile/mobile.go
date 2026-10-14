// Package mobile binds the Mochi usage companion to Android.
package mobile

import (
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/digital-wellbeing/dashboard"
)

var (
	app       *dashboard.App
	inputPage *ownframe.Page
)

// Dummy keeps the package visible to ebitenmobile bind.
func Dummy() {}
