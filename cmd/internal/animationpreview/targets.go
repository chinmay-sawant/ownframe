package animationpreview

type target struct{ key, id, button, trigger string }

var targets = []target{
	{key: "standard", id: "animation-standard-target", button: "start-standard-animation", trigger: "--standard-start"},
	{key: "webkit", id: "animation-webkit-target", button: "start-webkit-animation"},
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
