package ui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/store"
)

// rig drives a headless Quarry: manual clock, memory clipboard, memory store.
type rig struct {
	t   *testing.T
	a   *App
	ctx context.Context
	now time.Time
}

func newRig(t *testing.T, w, h int, db string) *rig {
	t.Helper()
	ownframe.UseMemory(true)
	ownframe.Clear()
	a, err := NewApp(context.Background(), LaunchOptions{DBPath: db, DataDir: store.Memory, DemoDir: t.TempDir(), Width: w, Height: h})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(a.Close)
	r := &rig{t: t, a: a, ctx: context.Background(), now: time.Unix(1_700_000_000, 0)}
	a.clock = func() time.Time { return r.now }
	a.SetModifiers(func() (bool, bool) { return false, false })
	if err := a.Page().Redraw(r.ctx); err != nil {
		t.Fatal(err)
	}
	return r
}

func (r *rig) must(err error) {
	r.t.Helper()
	if err != nil {
		r.t.Fatal(err)
	}
}

// box returns the laid-out box with id.
func (r *rig) box(id string) ownframe.Box {
	r.t.Helper()
	for _, b := range r.a.Page().Boxes() {
		if b.ID == id {
			return b
		}
	}
	r.t.Fatalf("no box %q", id)
	return ownframe.Box{}
}

func (r *rig) has(id string) bool {
	for _, b := range r.a.Page().Boxes() {
		if b.ID == id {
			return true
		}
	}
	return false
}

// tap presses and releases the center of box id.
func (r *rig) tap(id string) {
	r.t.Helper()
	b := r.box(id)
	x, y := float64(b.X+b.W/2), float64(b.Y+b.H/2)
	p := r.a.Page()
	r.must(p.Edit().Hover(r.ctx, x, y))
	r.must(p.Edit().Press(r.ctx, x, y))
	r.must(p.Click(r.ctx, x, y))
	r.must(p.Edit().Release(r.ctx))
	r.tick()
}

// tick advances the page and drains async completions.
func (r *rig) tick() {
	r.t.Helper()
	r.now = r.now.Add(150 * time.Millisecond)
	r.must(r.a.tickHook(r.ctx))
}

// settle waits for the running query to finish and drains it.
func (r *rig) settle() {
	r.t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for r.a.running != nil {
		if time.Now().After(deadline) {
			r.t.Fatal("query did not finish")
		}
		time.Sleep(5 * time.Millisecond)
		r.tick()
	}
	r.tick()
}

func (r *rig) key(k string) {
	r.t.Helper()
	r.must(r.a.onKeyDown(r.ctx, k))
}

func (r *rig) chord(k string, shift bool) {
	r.t.Helper()
	r.a.SetModifiers(func() (bool, bool) { return shift, true })
	r.key(k)
	r.a.SetModifiers(func() (bool, bool) { return false, false })
}

// typ types text into the focused surface.
func (r *rig) typ(s string) {
	r.t.Helper()
	r.must(r.a.onType(r.ctx, s))
}

// texts returns every text run painted on the page.
func (r *rig) texts() string {
	var sb strings.Builder
	for _, op := range r.a.Page().Probe().Display().Ops {
		if op.Kind == ownframe.DisplayOpText {
			sb.WriteString(op.Text)
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

func (r *rig) png(name string) {
	// Screenshots land in the test artifact directory (go test -artifacts).
	_ = os.WriteFile(filepath.Join(r.t.ArtifactDir(), name+".png"), r.a.Page().Probe().PNGWithContext(r.ctx), 0o600)
}

func (r *rig) openTableByName(name string) {
	r.t.Helper()
	for _, s := range r.a.schema {
		for _, g := range s.Children {
			for _, n := range g.Children {
				if n.Name == name {
					r.must(r.a.openTable(r.ctx, n))
					r.settle()
					return
				}
			}
		}
	}
	r.t.Fatalf("no table %q", name)
}
