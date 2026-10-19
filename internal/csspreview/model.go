package csspreview

type view struct {
	Started  map[string]bool
	Catalog  []catalogRow
	Readout  string
	Selected string
}

type target struct{ key, id, button, property, kind string }

func targets(kind string) []target {
	switch kind {
	case "transition":
		return transitionTargets()
	case "all":
		return catalogTargets()
	default:
		return animationTargets()
	}
}

func animationTargets() []target {
	return []target{
		{key: "standard", id: "animation-standard-target", button: "start-standard-animation", kind: "animation"},
		{key: "webkit", id: "animation-webkit-target", button: "start-webkit-animation", kind: "animation"},
	}
}

func transitionTargets() []target {
	out := make([]target, 0, 8)
	for _, engine := range []string{"standard", "webkit"} {
		for _, property := range []string{"left", "width", "opacity", "background"} {
			key := engine + "-" + property
			cssProperty := property
			if property == "background" {
				cssProperty = "background-color"
			}
			out = append(out, target{
				key: key, id: key + "-target", button: "start-" + key + "-transition",
				property: cssProperty, kind: "transition",
			})
		}
	}
	return out
}

func targetByButton(kind, id string) (target, bool) {
	for _, t := range targets(kind) {
		if t.button == id {
			return t, true
		}
	}
	return target{}, false
}

func targetByKey(kind, key string) (target, bool) {
	for _, t := range targets(kind) {
		if t.key == key {
			return t, true
		}
	}
	return target{}, false
}
