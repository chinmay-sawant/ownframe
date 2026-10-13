package masking

func combineMask(source, destination uint8, mode string) uint8 {
	s, d := int(source), int(destination)
	switch mode {
	case "subtract":
		return uint8((s*(255-d) + 127) / 255)
	case "intersect":
		return uint8((s*d + 127) / 255)
	case "exclude":
		return uint8((s*(255-d) + d*(255-s) + 127) / 255)
	default:
		return uint8((s*255 + d*(255-s) + 127) / 255)
	}
}
