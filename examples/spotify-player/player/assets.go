package player

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// registerImages supplies the bundled cover photos and interface symbols.
func registerImages(page *ownframe.Page) {
	coverFiles, err := audioFiles.ReadDir("assets/audio/covers")
	if err == nil {
		for _, cover := range coverFiles {
			name := strings.TrimSuffix(cover.Name(), ".jpg")
			body, readErr := audioFiles.ReadFile("assets/audio/covers/" + cover.Name())
			if readErr == nil {
				page.SetImage("cover-"+name, body)
			}
		}
	}

	entries, err := files.ReadDir("assets")
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), ".png")
		if name == entry.Name() || strings.HasPrefix(name, "cover-") {
			continue
		}

		page.SetImage(name, asset("assets/"+entry.Name()))
	}

	aliases := []string{"late-night", "spring", "funk", "focus", "chillhop", "soul"}
	for i, cover := range aliases {
		body, err := audioFiles.ReadFile("assets/audio/covers/" + cover + ".jpg")
		if err != nil {
			continue
		}

		page.SetImage("track-"+strconv.Itoa(i), body)
		page.SetImage("pick-"+strconv.Itoa(i), body)
		page.SetImage("card-"+strconv.Itoa(i), body)
	}
}
