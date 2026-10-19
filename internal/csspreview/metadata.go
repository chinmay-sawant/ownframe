package csspreview

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

func describe(p *ownframe.Page, kind string) string {
	var b strings.Builder
	for _, target := range targets(kind) {
		box, ok := findBox(p.Boxes(), target.id)
		if !ok {
			fmt.Fprintf(&b, "%s: metadata missing\n", target.id)
			continue
		}
		rules := []any{}
		if target.kind == "animation" {
			for _, a := range box.Animations {
				for _, k := range p.Display().Keyframes {
					if k.Name == a.Name {
						rules = append(rules, k)
					}
				}
			}
		}
		data, _ := json.Marshal(map[string]any{"box": box, "keyframes": rules})
		fmt.Fprintf(&b, "%s: %s\n", target.id, data)
	}
	b.WriteString("This readout shows the initial display-list snapshot. Press Start to play the visual sample on this page.\n")
	return b.String()
}

func findBox(boxes []ownframe.Box, id string) (ownframe.Box, bool) {
	for _, b := range boxes {
		if b.ID == id {
			return b, true
		}
	}
	return ownframe.Box{}, false
}
