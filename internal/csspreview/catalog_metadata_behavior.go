package csspreview

func animationMetadataBehavior(name string) string {
	switch name {
	case "animation-range":
		return "The blue tile glides; the configured start and end range appear above."
	case "animation-range-center":
		return "The blue tile glides; the configured center range appears above."
	case "animation-range-start":
		return "The blue tile glides; the configured start point appears above."
	case "animation-range-end":
		return "The blue tile glides; the configured end point appears above."
	case "animation-timeline":
		return "The blue tile glides on the preview clock; the configured timeline appears above."
	default:
		return "Start launches the blue glide; the configured trigger appears above."
	}
}

func otherMetadataBehavior(name string) string {
	switch name {
	case "trigger-scope":
		return "Start launches the blue glide; the configured trigger scope appears above."
	case "view-transition-class":
		return "The blue tile glides; its configured view-transition class appears above."
	case "view-transition-group":
		return "The blue tile glides; its configured view-transition group appears above."
	case "view-transition-name":
		return "The blue tile glides; its configured view-transition name appears above."
	default:
		return "The blue tile glides; its configured view-transition scope appears above."
	}
}
