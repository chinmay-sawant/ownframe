//go:build linux && !android

package clipboard

func pickAuth(b []byte, host, num string) (string, []byte) {
	var name string
	var data []byte
	best := 0

	for len(b) > 0 {
		fam, addr, disp, nm, dat, rest, ok := nextAuth(b)
		if !ok {
			break
		}

		b = rest
		if disp != "" && disp != num {
			continue
		}

		if nm != "" && nm != "MIT-MAGIC-COOKIE-1" {
			continue
		}

		score := 1
		if fam == 256 && addr == host {
			score = 3
		} else if fam == 65535 || fam == 252 {
			score = 2
		}

		if score > best {
			best = score
			name = nm
			data = dat
		}
	}

	return name, data
}

func nextAuth(b []byte) (uint16, string, string, string, []byte, []byte, bool) {
	if len(b) < xAuthHdr {
		return 0, "", "", "", nil, nil, false
	}

	fam := uint16(b[0])<<xShift8 | uint16(b[1])
	addr, b, ok := cutStr(b[2:])
	if !ok {
		return 0, "", "", "", nil, nil, false
	}

	disp, b, ok := cutStr(b)
	if !ok {
		return 0, "", "", "", nil, nil, false
	}

	name, b, ok := cutStr(b)
	if !ok {
		return 0, "", "", "", nil, nil, false
	}

	raw, b, ok := cutStr(b)
	if !ok {
		return 0, "", "", "", nil, nil, false
	}

	return fam, addr, disp, name, []byte(raw), b, true
}

func cutStr(b []byte) (string, []byte, bool) {
	if len(b) < xAuthLen {
		return "", nil, false
	}

	n := int(b[0])<<xShift8 | int(b[1])
	b = b[2:]

	if n < 0 || len(b) < n {
		return "", nil, false
	}

	return string(b[:n]), b[n:], true
}
