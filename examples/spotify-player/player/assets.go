package player

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// registerImages supplies pre-rasterized images and aliases the cover slots.
func registerImages(page *ownframe.Page) {
	entries, err := files.ReadDir("assets")
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := strings.TrimSuffix(entry.Name(), ".png")
		if name == entry.Name() {
			continue
		}

		page.SetImage(name, asset("assets/"+entry.Name()))
	}

	for i := 0; i < 6; i++ {
		cover := "cover-" + strconv.Itoa(i+1)
		body := asset("assets/" + cover + ".png")
		page.SetImage("track-"+strconv.Itoa(i), body)
		page.SetImage("pick-"+strconv.Itoa(i), body)
		page.SetImage("card-"+strconv.Itoa(i), body)
	}
}
