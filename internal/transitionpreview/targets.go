package transitionpreview

type target struct{ key, id, button, property string }

var targets = makeTargets()

func makeTargets() []target {
	out := make([]target, 0, targetCap)
	for _, engine := range []string{"standard", "webkit"} {
		for _, property := range []string{"left", "width", "opacity", "background"} {
			key := engine + "-" + property
			name := property
			if property == "background" {
				name = "background-color"
			}
			out = append(out, target{key: key, id: key + "-target",
				button: "start-" + key + "-transition", property: name})
		}
	}
	return out
}

func byButton(id string) (target, bool) {
	for _, item := range targets {
		if item.button == id {
			return item, true
		}
	}
	return target{}, false
}

func byKey(key string) (target, bool) {
	for _, item := range targets {
		if item.key == key {
			return item, true
		}
	}
	return target{}, false
}
