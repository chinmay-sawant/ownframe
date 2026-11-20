package transitionpreview

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

func describe(page *ownframe.Page) string {
	var out strings.Builder
	for _, target := range targets {
		box, ok := findBox(page.Boxes(), target.id)
		if !ok {
			fmt.Fprintf(&out, "%s: metadata missing\n", target.id)
			continue
		}
		data, _ := json.Marshal(map[string]any{"box": box, "transitions": box.Transitions})
		fmt.Fprintf(&out, "%s: %s\n", target.id, data)
	}
	return out.String()
}
