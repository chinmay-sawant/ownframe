package page

import (
	"context"
	"testing"
)

func TestTickManualOverlapDiagnostic(t *testing.T) {
	html := `<style>
@keyframes pulse { from { opacity: 0.2; } to { opacity: 1; } }
#anim { width: 40px; height: 40px; background: #ff0000; animation: pulse 2s linear infinite; }
#other { width: 40px; height: 40px; background: #0000ff; }
</style><div id="anim"></div><div id="other"></div>`

	p, err := New(Config{HTML: html, Width: 200, Height: 120})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	// Mutating an element owned by an active CSS animation triggers ManualOverlap.
	p.SetTick(func(context.Context) error {
		d := p.Display()
		for i := range d.Ops {
			if d.Ops[i].R > 0.5 && d.Ops[i].B < 0.5 {
				d.Ops[i].X += 5
				break
			}
		}
		return nil
	})
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	st := p.Motion()
	if st.ManualOverlap != 1 || st.ManualOverlapID != "anim" {
		t.Fatalf("after overlapping tick: ManualOverlap=%d ID=%q, want 1 / %q", st.ManualOverlap, st.ManualOverlapID, "anim")
	}

	// Mutating a non-animated element does not increment ManualOverlap.
	p.SetTick(func(context.Context) error {
		d := p.Display()
		for i := range d.Ops {
			if d.Ops[i].B > 0.5 && d.Ops[i].R < 0.5 {
				d.Ops[i].X += 5
				break
			}
		}
		return nil
	})
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().ManualOverlap; got != 1 {
		t.Fatalf("non-overlapping tick changed ManualOverlap to %d, want 1", got)
	}

	// Turning animations off prevents overlap diagnostics when driving #anim manually.
	p.SetAnimations(false)
	p.SetTick(func(context.Context) error {
		d := p.Display()
		for i := range d.Ops {
			if d.Ops[i].R > 0.5 && d.Ops[i].B < 0.5 {
				d.Ops[i].X += 5
				break
			}
		}
		return nil
	})
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().ManualOverlap; got != 1 {
		t.Fatalf("after SetAnimations(false), ManualOverlap=%d, want 1 (unchanged)", got)
	}
}
