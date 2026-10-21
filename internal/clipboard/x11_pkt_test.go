//go:build linux && !android

package clipboard

import "testing"

func TestX11PropertyBytes(t *testing.T) {
	got := prop8(1, 2, 3, []byte("hi"))
	want := []byte{
		18, 0, 7, 0,
		1, 0, 0, 0,
		2, 0, 0, 0,
		3, 0, 0, 0,
		8, 0, 0, 0,
		2, 0, 0, 0,
		'h', 'i', 0, 0,
	}

	if len(got) != len(want) {
		t.Fatalf("len %d", len(got))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("byte %d = %d, want %d", i, got[i], want[i])
		}
	}

	atom := internReq("UTF8_STRING")
	if atom[0] != 16 || get16(atom[4:]) != 11 {
		t.Fatalf("intern %v", atom)
	}

	if string(atom[8:19]) != "UTF8_STRING" || len(atom) != 20 {
		t.Fatalf("intern name %q len %d", atom[8:], len(atom))
	}
}

func TestSetupRoot(t *testing.T) {
	buf := make([]byte, 120)
	buf[0] = 1
	put32(buf[12:], 0x100)
	put32(buf[16:], 0x1f)
	buf[28] = 2
	put32(buf[40:], 7)
	put32(buf[80:], 99)

	base, mask, _, ok := setupIDs(buf)
	if !ok || base != 0x100 || mask != 0x1f {
		t.Fatalf("ids %x %x %v", base, mask, ok)
	}

	root, ok := rootAt(buf, 1)
	if !ok || root != 99 {
		t.Fatalf("root %d %v", root, ok)
	}
}

func TestDisplayAndAuth(t *testing.T) {
	host, num, scr, ok := displaySpec("unix:1.2")
	if !ok || host != "" || num != "1" || scr != 2 {
		t.Fatalf("display %q %q %d %v", host, num, scr, ok)
	}

	raw := authEntry(256, "box", "1", "MIT-MAGIC-COOKIE-1", "0123456789abcdef")
	name, data := pickAuth(raw, "box", "1")
	if name != "MIT-MAGIC-COOKIE-1" || string(data) != "0123456789abcdef" {
		t.Fatalf("auth %q %q", name, data)
	}
}

func authEntry(fam int, addr, disp, name, data string) []byte {
	var b []byte
	b = append(b, byte(fam>>8), byte(fam)) //nolint:gosec // reason: big-endian family byte split.
	b = appendStr(b, addr)
	b = appendStr(b, disp)
	b = appendStr(b, name)
	b = appendStr(b, data)

	return b
}

func appendStr(b []byte, s string) []byte {
	b = append(b, byte(len(s)>>8), byte(len(s))) //nolint:gosec // reason: big-endian length byte split.

	return append(b, s...)
}
