package main

func animationMetadataBehavior(name string) string {
	switch name {
	case "animation-range":
		return "An S-curve runs while the configured entry and exit range appears above."
	case "animation-range-center":
		return "A green tile fades through three levels; its center range appears above."
	case "animation-range-start":
		return "An orange tile steps across the track; its start range appears above."
	case "animation-range-end":
		return "A violet tile moves backward; its end range appears above."
	case "animation-timeline":
		return "A blue tile follows a broken diagonal on the preview clock."
	default:
		return "Start launches a dotted yellow sweep; the trigger appears above."
	}
}

func otherMetadataBehavior(name string) string {
	switch name {
	case "trigger-scope":
		return "A mint tile zigzags when Start fires the scoped trigger."
	case "view-transition-class":
		return "A striped-color tile sweeps right; its transition class appears above."
	case "view-transition-group":
		return "An amber tile traces a V; its transition group appears above."
	case "view-transition-name":
		return "A red tile follows a long diagonal; its transition name appears above."
	default:
		return "A lilac tile crosses left; its transition scope appears above."
	}
}
