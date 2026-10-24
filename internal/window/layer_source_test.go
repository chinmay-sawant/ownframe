package window

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/replay"
)

// replay.LayerSource and host.LayerSource must stay one type. The shell asserts
// the replay role on the screen it holds, and that screen is *page.Page.
var _ replay.LayerSource = (*page.Page)(nil)

// windowPageLayerHTML animates one element's transform, the entry the page
// promotes today.
const windowPageLayerHTML = `<style>
@keyframes slide { 0% { transform: translateX(0px); } 100% { transform: translateX(120px); } }
#target { width: 40px; height: 20px; background-color: rgb(20,40,60);
	animation-name: slide; animation-duration: 2s;
	animation-timing-function: linear; animation-iteration-count: infinite;
	animation-fill-mode: both; }
</style><div id="target"></div>`

// TestWindowFindsPageLayerSource proves the window reaches the page through its
// LayerSource assertion after the type move, not through a wrapper.
func TestWindowFindsPageLayerSource(t *testing.T) {
	t.Parallel()

	p, err := page.New(page.Config{HTML: windowPageLayerHTML, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}

	shell := &shell{app: p}
	if _, ok := shell.app.(replay.LayerSource); !ok {
		t.Fatal("the shell's screen does not satisfy replay.LayerSource")
	}
	specs := shell.layerSpecs(p.Display())
	if len(specs) != 1 {
		t.Fatalf("window layer specs = %d, want one", len(specs))
	}
	if specs[0].ElementID != "target" {
		t.Fatalf("element = %q, want target", specs[0].ElementID)
	}
	if specs[0].Matrix.IsIdentity() {
		t.Fatal("the window received an identity layer matrix")
	}
}
