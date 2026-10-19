package csspreview

func catalogTargets() []target {
	out := make([]target, 0, len(catalogProperties))
	for _, name := range catalogProperties {
		key, _ := catalogAction(name)
		if key == "" {
			continue
		}
		item := target{key: key, id: catalogTargetID(name), button: catalogButton(name)}
		if catalogPreviewClass(name) == "catalog-animation-standard" ||
			catalogPreviewClass(name) == "catalog-animation-webkit" {
			item.kind = "animation"
		} else {
			item.kind = "transition"
			item.property = transitionPreviewProperty(name)
		}
		out = append(out, item)
	}
	return out
}
