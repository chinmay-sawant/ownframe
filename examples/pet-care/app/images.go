package app

import (
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/assets"
)

func prepareImages(page *ownframe.Page) {
	if data, err := assets.Files.ReadFile("empty-care.png"); err == nil {
		page.SetImage("empty-care", data)
	}
	if data, err := assets.Files.ReadFile("dog-profile.png"); err == nil {
		page.SetImage("dog-profile", data)
	}
}
