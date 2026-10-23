package draw

import "testing"

func TestImageStoreClonesAndDeletes(t *testing.T) {
	t.Parallel()

	data := []byte{1, 2}
	images := Store(nil, "logo", data)
	data[0] = 9
	got, ok := Lookup(images, "logo")
	if !ok || got[0] != 1 {
		t.Fatal("clone")
	}

	images = Store(images, "logo", nil)
	if _, ok := Lookup(images, "logo"); ok {
		t.Fatal("deleted")
	}
	if Store(nil, "missing", nil) == nil {
		t.Fatal("nil delete")
	}
}

func TestImageLookupEmoji(t *testing.T) {
	t.Parallel()

	if _, ok := Lookup(nil, "emoji/1f602"); !ok {
		t.Fatal("bundled")
	}
	if _, ok := Lookup(nil, "emoji/1fae0"); ok {
		t.Fatal("missing")
	}
	if _, ok := Lookup(nil, "logo"); ok {
		t.Fatal("logo")
	}

	registered := map[string][]byte{"emoji/1f602": {7}}
	got, ok := Lookup(registered, "emoji/1f602")
	if !ok || len(got) != 1 || got[0] != 7 {
		t.Fatalf("registered %v %v", got, ok)
	}
}
