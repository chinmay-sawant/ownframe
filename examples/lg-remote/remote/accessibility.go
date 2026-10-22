package remote

import (
	"encoding/json"
	"strings"

	"github.com/chinmay-sawant/ownframe/examples/lg-remote/bridge"
)

func (a *App) publishAccessibility() {
	if !a.phone {
		return
	}
	x, y := a.page.ScrollOffset()
	now := accessState{gen: a.page.Generation(), x: x, y: y}
	if now == a.access {
		return
	}
	a.access = now
	w, h := a.page.Size()
	scale, ox, oy := 1.0, 0.0, 0.0
	if a.page.ViewLocked() {
		scale, ox, oy = a.phoneFit()
		x, y = 0, 0
	}
	keys := a.controlNames()
	nodes := []accessibleNode{}
	seen := map[string]bool{}
	for _, box := range a.page.Boxes() {
		if box.ID == "" || seen[box.ID] {
			continue
		}
		seen[box.ID] = true
		key, button := keys[box.ID]
		if !button && box.ID != "status" && box.ID != "hint" {
			continue
		}
		bx, by := (box.X-float64(x))*scale+ox, (box.Y-float64(y))*scale+oy
		bw, bh := box.W*scale, box.H*scale
		if bx+bw <= 0 || by+bh <= 0 || bx >= float64(w) || by >= float64(h) {
			continue
		}
		name := key.Name
		if !button {
			name = box.Text
		}
		node := accessibleNode{ID: box.ID, Name: name, X: bx, Y: by,
			W: bw, H: bh, Enabled: !key.Disabled, Button: button}
		node.Editable = box.ID == "host"
		if box.ID == "sensitivity" {
			node.Slider = true
			node.Button = false
			node.Progress = a.view.Sensitivity
			node.Value = a.view.SensitivityLabel
		}
		if node.Editable {
			node.Value = a.page.FormValue("host")
			node.Button = false
		}
		node.Selected = box.ID == "tab-"+a.view.Panel
		if strings.TrimSpace(name) != "" {
			nodes = append(nodes, node)
		}
	}
	data, err := json.Marshal(struct {
		Width, Height int
		Panel         string
		Scrollable    bool
		Nodes         []accessibleNode
	}{w, h, a.view.Panel, a.page.AllowScroll() && !a.page.ViewLocked(), nodes})
	if err == nil {
		bridge.SetAccessibility(string(data))
	}
}
