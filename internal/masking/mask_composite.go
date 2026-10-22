package masking

func combineMask(source, destination uint8, mode string) uint8 {
	s, d := float64(source)/channelMax, float64(destination)/channelMax
	switch mode {
	case "subtract":
		return channel(s * (1 - d))
	case "intersect":
		return channel(s * d)
	case "exclude":
		return channel(s*(1-d) + d*(1-s))
	default:
		return channel(s + d*(1-s))
	}
}
