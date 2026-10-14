package element

func changedProps(a, b *Node) []string {
	var props []string
	if a == nil || b == nil {
		return props
	}
	for k, v := range b.Attrs {
		if av, ok := a.Attrs[k]; !ok || av != v {
			props = append(props, "attr:"+k)
		}
	}
	for k := range a.Attrs {
		if _, ok := b.Attrs[k]; !ok {
			props = append(props, "attr:"+k)
		}
	}
	for k, v := range b.Metadata {
		if av, ok := a.Metadata[k]; !ok || av != v {
			props = append(props, "metadata:"+k)
		}
	}
	for k := range a.Metadata {
		if _, ok := b.Metadata[k]; !ok {
			props = append(props, "metadata:"+k)
		}
	}
	if a.Kind != b.Kind {
		props = append(props, "kind")
	}
	if a.Key != b.Key {
		props = append(props, "key")
	}
	return props
}
