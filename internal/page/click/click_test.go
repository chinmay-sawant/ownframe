package click

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/fetch"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

var (
	errClick = errors.New("click")
	errPaint = errors.New("paint")
	errOpen  = errors.New("open")
	errLoad  = errors.New("load")
	errDraw  = errors.New("draw")
)

// TestClickSequenceOrder checks decision order on a fake surface.
// The page click tests still call Page.Click.
func TestClickSequenceOrder(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		f       fake
		want    string
		wantErr error
	}{
		{
			name: "miss redraws when changed",
			want: "HitBox HasOpenLink RedrawIfChanged",
		},
		{
			name: "consumed press redraws and stops",
			f: fake{
				hitOK: true,
				hit:   layout.Box{ID: "a"},
				press: true,
			},
			want: "HitBox ConsumePress:a Redraw",
		},
		{
			name: "blank id uses the inner id before the route",
			f: fake{
				hitOK: true,
				hit:   layout.Box{ID: ""},
				idAt:  "icon",
				byID:  layout.Box{ID: "icon", Action: "go"},
				routes: map[string]string{
					"go": "<b>n</b>",
				},
			},
			want: "HitBox BoxIDAt BoxByID:icon ConsumePress:icon FormControl:icon BlurForm Lookup:go Load:<b>n</b>",
		},
		{
			name: "blank id with no inner id stays blank",
			f: fake{
				hitOK: true,
				hit:   layout.Box{},
			},
			want: "HitBox BoxIDAt ConsumePress: FormControl: BlurForm HasOpenLink HasClick ClickPainted: RedrawIfChanged",
		},
		{
			name: "enabled button loads its route",
			f: fake{
				hitOK: true,
				hit:   layout.Box{ID: "b", Action: "next"},
				form:  true,
				ctrl:  form.Control{Tag: "button"},
				routes: map[string]string{
					"next": "<p>x</p>",
				},
			},
			want: "HitBox ConsumePress:b FormControl:b Control:b Lookup:next Load:<p>x</p>",
		},
		{
			name: "disabled button does not route or blur",
			f: fake{
				hitOK: true,
				hit:   layout.Box{ID: "b", Action: "next"},
				form:  true,
				ctrl:  form.Control{Tag: "button", Disabled: true},
				routes: map[string]string{
					"next": "<p>x</p>",
				},
			},
			want: "HitBox ConsumePress:b FormControl:b Control:b ClickControl:b",
		},
		{
			name: "button without a route activates the control",
			f: fake{
				hitOK: true,
				hit:   layout.Box{ID: "b", Action: "next"},
				form:  true,
				ctrl:  form.Control{Tag: "button"},
			},
			want: "HitBox ConsumePress:b FormControl:b Control:b Lookup:next ClickControl:b",
		},
		{
			name: "button without an action activates the control",
			f: fake{
				hitOK: true,
				hit:   layout.Box{ID: "b"},
				form:  true,
				ctrl:  form.Control{Tag: "button"},
			},
			want: "HitBox ConsumePress:b FormControl:b Control:b ClickControl:b",
		},
		{
			name: "text control activates and does not blur",
			f: fake{
				hitOK: true,
				hit:   layout.Box{ID: "t"},
				form:  true,
				ctrl:  form.Control{Tag: "input", Type: "text"},
			},
			want: "HitBox ConsumePress:t FormControl:t Control:t ClickControl:t",
		},
		{
			name: "other hit blurs and loads the route",
			f: fake{
				hitOK: true,
				hit:   layout.Box{ID: "n", Action: "go"},
				routes: map[string]string{
					"go": "html",
				},
			},
			want: "HitBox ConsumePress:n FormControl:n BlurForm Lookup:go Load:html",
		},
		{
			name: "route miss still skips the link and returns the handler error",
			f: fake{
				hitOK:    true,
				hit:      layout.Box{ID: "n", Action: "go"},
				open:     true,
				click:    true,
				clickErr: errClick,
			},
			want:    "HitBox ConsumePress:n FormControl:n BlurForm Lookup:go HasOpenLink HasClick HandlerClick:n",
			wantErr: errClick,
		},
		{
			name: "paint claim skips the changed redraw",
			f: fake{
				hitOK:   true,
				hit:     layout.Box{ID: "n"},
				painted: true,
			},
			want: "HitBox ConsumePress:n FormControl:n BlurForm HasOpenLink HasClick ClickPainted:n",
		},
		{
			name: "paint error skips the changed redraw",
			f: fake{
				hitOK:    true,
				hit:      layout.Box{ID: "n"},
				paintErr: errPaint,
			},
			want:    "HitBox ConsumePress:n FormControl:n BlurForm HasOpenLink HasClick ClickPainted:n",
			wantErr: errPaint,
		},
		{
			name: "miss opens a link and stops",
			f: fake{
				open:    true,
				handled: true,
				openErr: errOpen,
			},
			want:    "HitBox HasOpenLink OpenLinkAt",
			wantErr: errOpen,
		},
		{
			name: "declined link still redraws when changed",
			f:    fake{open: true},
			want: "HitBox HasOpenLink OpenLinkAt RedrawIfChanged",
		},
		{
			name: "load error is returned",
			f: fake{
				hitOK: true,
				hit:   layout.Box{ID: "n", Action: "go"},
				routes: map[string]string{
					"go": "html",
				},
				loadErr: errLoad,
			},
			want:    "HitBox ConsumePress:n FormControl:n BlurForm Lookup:go Load:html",
			wantErr: errLoad,
		},
		{
			name:    "changed redraw error is returned",
			f:       fake{changedErr: errDraw},
			want:    "HitBox HasOpenLink RedrawIfChanged",
			wantErr: errDraw,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := tc.f
			err := Run(context.Background(), &f, 1, 1)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}

			got := strings.Join(f.calls, " ")
			if got != tc.want {
				t.Fatalf("calls = %s", got)
			}

			if f.hitX != 1 || f.hitY != 1 {
				t.Fatalf("hit at %v,%v", f.hitX, f.hitY)
			}

			if strings.Contains(tc.want, "OpenLinkAt") && (f.openX != 1 || f.openY != 1) {
				t.Fatalf("open at %v,%v", f.openX, f.openY)
			}
		})
	}
}

func TestClickRejectsNilContext(t *testing.T) {
	t.Parallel()

	f := &fake{}
	//nolint:staticcheck // reason: intentionally passes nil to verify ErrNilContext.
	err := Run(nil, f, 1, 1)
	if !errors.Is(err, fetch.ErrNilContext) {
		t.Fatalf("err = %v", err)
	}

	if len(f.calls) != 0 {
		t.Fatalf("calls = %v", f.calls)
	}
}

func TestClickStopsOnCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	f := &fake{}
	err := Run(ctx, f, 1, 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}

	if len(f.calls) != 0 {
		t.Fatalf("calls = %v", f.calls)
	}
}

type fake struct {
	hit        layout.Box
	hitOK      bool
	idAt       string
	byID       layout.Box
	press      bool
	form       bool
	ctrl       form.Control
	routes     map[string]string
	open       bool
	handled    bool
	openErr    error
	click      bool
	clickErr   error
	painted    bool
	paintErr   error
	loadErr    error
	redrawErr  error
	changedErr error
	calls      []string
	hitX       float64
	hitY       float64
	openX      float64
	openY      float64
}

func (f *fake) note(step string) {
	f.calls = append(f.calls, step)
}

func (f *fake) HitBox(x, y float64) (layout.Box, bool) {
	f.note("HitBox")
	f.hitX, f.hitY = x, y

	return f.hit, f.hitOK
}

func (f *fake) BoxIDAt(float64, float64) string {
	f.note("BoxIDAt")

	return f.idAt
}

func (f *fake) BoxByID(id string) layout.Box {
	f.note("BoxByID:" + id)

	return f.byID
}

func (f *fake) ConsumePress(id string) bool {
	f.note("ConsumePress:" + id)

	return f.press
}

func (f *fake) FormControl(id string) bool {
	f.note("FormControl:" + id)

	return f.form
}

func (f *fake) Control(id string) (form.Control, bool) {
	f.note("Control:" + id)

	return f.ctrl, true
}

func (f *fake) Lookup(action string) (string, bool) {
	f.note("Lookup:" + action)
	html, ok := f.routes[action]

	return html, ok
}

func (f *fake) Load(_ context.Context, html string) error {
	f.note("Load:" + html)

	return f.loadErr
}

func (f *fake) ClickControl(_ context.Context, box layout.Box) error {
	f.note("ClickControl:" + box.ID)

	return nil
}

func (f *fake) BlurForm() {
	f.note("BlurForm")
}

func (f *fake) HasOpenLink() bool {
	f.note("HasOpenLink")

	return f.open
}

func (f *fake) OpenLinkAt(_ context.Context, x, y float64) (bool, error) {
	f.note("OpenLinkAt")
	f.openX, f.openY = x, y

	return f.handled, f.openErr
}

func (f *fake) HasClick() bool {
	f.note("HasClick")

	return f.click
}

func (f *fake) HandlerClick(_ context.Context, box layout.Box) error {
	f.note("HandlerClick:" + box.ID)

	return f.clickErr
}

func (f *fake) ClickPainted(_ context.Context, box layout.Box) (bool, error) {
	f.note("ClickPainted:" + box.ID)

	return f.painted, f.paintErr
}

func (f *fake) Redraw(context.Context) error {
	f.note("Redraw")

	return f.redrawErr
}

func (f *fake) RedrawIfChanged(context.Context) error {
	f.note("RedrawIfChanged")

	return f.changedErr
}

var _ Surface = (*fake)(nil)
