package bitmap

func inRound(x, y, w, h float64, rx, ry [4]float64) bool {
	if x < 0 || y < 0 || x >= w || y >= h {
		return false
	}
	centers := [4][2]float64{{rx[0], ry[0]}, {w - rx[1], ry[1]},
		{w - rx[2], h - ry[2]}, {rx[3], h - ry[3]}}
	for i, c := range centers {
		cornerX := x < c[0]
		if i == 1 || i == 2 {
			cornerX = x > c[0]
		}
		cornerY := y < c[1]
		if i >= 2 {
			cornerY = y > c[1]
		}
		if cornerX && cornerY && rx[i] > 0 && ry[i] > 0 {
			dx, dy := (x-c[0])/rx[i], (y-c[1])/ry[i]
			return dx*dx+dy*dy <= 1
		}
	}
	return true
}
