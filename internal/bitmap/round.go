package bitmap

func inRound(x, y, w, h float64, rx, ry [4]float64) bool {
	if x < 0 || y < 0 || x >= w || y >= h {
		return false
	}
	if rx[0] > 0 && ry[0] > 0 && x < rx[0] && y < ry[0] {
		dx, dy := (x-rx[0])/rx[0], (y-ry[0])/ry[0]
		return dx*dx+dy*dy <= 1
	}
	if rx[1] > 0 && ry[1] > 0 && x > w-rx[1] && y < ry[1] {
		dx, dy := (x-(w-rx[1]))/rx[1], (y-ry[1])/ry[1]
		return dx*dx+dy*dy <= 1
	}
	if rx[2] > 0 && ry[2] > 0 && x > w-rx[2] && y > h-ry[2] {
		dx, dy := (x-(w-rx[2]))/rx[2], (y-(h-ry[2]))/ry[2]
		return dx*dx+dy*dy <= 1
	}
	if rx[3] > 0 && ry[3] > 0 && x < rx[3] && y > h-ry[3] {
		dx, dy := (x-rx[3])/rx[3], (y-(h-ry[3]))/ry[3]
		return dx*dx+dy*dy <= 1
	}
	return true
}
