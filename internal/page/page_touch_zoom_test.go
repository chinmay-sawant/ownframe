package page

import "testing"

func TestTouchZoomConfigAndOverride(t *testing.T) {
	p, err := New(Config{HTML: "<main>ok</main>", Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if !p.TouchZoomAllowed() {
		t.Fatal("touch zoom should remain enabled by default")
	}
	p.SetTouchZoomAllowed(false)
	if p.TouchZoomAllowed() {
		t.Fatal("runtime override did not disable touch zoom")
	}
	p, err = New(Config{HTML: "<main>ok</main>", Width: 100, Height: 100, DisableTouchZoom: true})
	if err != nil {
		t.Fatal(err)
	}
	if p.TouchZoomAllowed() {
		t.Fatal("config did not disable touch zoom")
	}
}

func TestWheelZoomConfig(t *testing.T) {
	p, err := New(Config{HTML: "<main>ok</main>", Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	if !p.WheelZoomAllowed() {
		t.Fatal("wheel zoom should remain enabled by default")
	}

	p, err = New(Config{HTML: "<main>ok</main>", Width: 100, Height: 100, DisableWheelZoom: true})
	if err != nil {
		t.Fatal(err)
	}

	if p.WheelZoomAllowed() {
		t.Fatal("config did not disable wheel zoom")
	}
}

func TestKeyZoomConfig(t *testing.T) {
	p, err := New(Config{HTML: "<main>ok</main>", Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	if !p.KeyZoomAllowed() {
		t.Fatal("key zoom should remain enabled by default")
	}

	p, err = New(Config{HTML: "<main>ok</main>", Width: 100, Height: 100, DisableKeyZoom: true})
	if err != nil {
		t.Fatal(err)
	}

	if p.KeyZoomAllowed() {
		t.Fatal("config did not disable key zoom")
	}

	if !p.WheelZoomAllowed() {
		t.Fatal("key zoom config also disabled wheel zoom")
	}
}
