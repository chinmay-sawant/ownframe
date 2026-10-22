package ownframe

import (
	"context"
	"testing"
)

func repaintCases() []repaintCase {
	return []repaintCase{
		loginCase(),
		platformCase(),
		statesCase(),
		formsCase(),
		scrollCase(),
		themeCase(),
	}
}

const loginRepaintHTML = `<html><body style="margin:0;background:#f4f1ea">
<div style="width:280px;margin:20px;padding:16px;background:#fff">
<h1>Sign in</h1><p id="msg">{{.Msg}}</p>
<input id="email" style="display:block;width:240px;border:1px solid #999;padding:6px">
<input id="password" style="display:block;width:240px;border:1px solid #999;padding:6px">
<button id="login">Sign in</button></div></body></html>`

// loginCase types the example's secret pair and submits.
func loginCase() repaintCase {
	return repaintCase{
		name: "login",
		newPage: func(t *testing.T) *Page {
			return newPage(t, Config{Title: "Sign in", HTML: loginRepaintHTML, Width: 320, Height: 260},
				func(p *Page) {
					p.SetData(map[string]string{"Msg": ""})
					p.Handle(Handlers{Click: func(_ context.Context, box Box) error {
						if box.ID != "login" {
							return nil
						}

						msg := "Denied"
						if p.FormValue("email") == "secret" && p.FormValue("password") == "secret" {
							msg = "Welcome"
						}

						p.SetData(map[string]string{"Msg": msg})

						return nil
					}})
				})
		},
		interact: func(t *testing.T, p *Page, ctx context.Context) {
			clickID(t, p, ctx, "email")
			typeText(t, p, ctx, "secret")
			clickID(t, p, ctx, "password")
			typeText(t, p, ctx, "secret")
			clickID(t, p, ctx, "login")
		},
	}
}

func typeText(t *testing.T, p *Page, ctx context.Context, text string) {
	t.Helper()

	if err := p.Testing().Type(ctx, text); err != nil {
		t.Fatal(err)
	}
}

const formsRepaintHTML = `<html><head><style>
#agree { width:18px;height:18px }
#agree:checked { outline:3px solid #176b45 }
#email { display:block;width:240px;border:1px solid #999;padding:6px }
</style></head><body style="margin:0;background:#f4f1ea">
<div style="width:280px;margin:20px;padding:16px;background:#fff">
<input id="agree" type="checkbox"><label for="agree">Agree</label>
<input id="email">
</div></body></html>`

// formsCase toggles the checkbox and types into the field.
func formsCase() repaintCase {
	return repaintCase{
		name: "forms",
		newPage: func(t *testing.T) *Page {
			return newPage(t, Config{Title: "Forms", HTML: formsRepaintHTML, Width: 320, Height: 220}, nil)
		},
		interact: func(t *testing.T, p *Page, ctx context.Context) {
			clickID(t, p, ctx, "agree")
			clickID(t, p, ctx, "email")
			typeText(t, p, ctx, "note")
		},
	}
}

const scrollRepaintHTML = `<html><body style="margin:0;background:#f4f1ea">
<div style="height:800px"></div>
<div style="width:300px;margin:20px;padding:16px;background:#fff">
<p id="status">{{.Status}}</p><button id="far">Far</button></div></body></html>`

// scrollCase clicks a control below the fold of a page taller than its frame.
func scrollCase() repaintCase {
	return repaintCase{
		name: "scrolling",
		newPage: func(t *testing.T) *Page {
			return newPage(t, Config{Title: "Scrolling", HTML: scrollRepaintHTML, Width: 320, Height: 200},
				func(p *Page) {
					p.SetData(map[string]string{"Status": ""})
					p.Handle(Handlers{Click: func(_ context.Context, box Box) error {
						if box.ID == "far" {
							p.SetData(map[string]string{"Status": "reached"})
						}

						return nil
					}})
				})
		},
		interact: func(t *testing.T, p *Page, ctx context.Context) {
			clickID(t, p, ctx, "far")
		},
	}
}

const themeRepaintHTML = `<html><body style="margin:0;background:var(--bg,#f4f1ea)">
<div style="width:280px;margin:20px;padding:16px;background:var(--card,#fff)">
<h1 style="color:var(--ink,#1c1915)">{{.Title}}</h1>
<div id="toggle">Switch</div></div></body></html>`

const themeRepaintCSS = `body { background: #102030 } div { background: #203040 } h1 { color: #ffffff }`

// themeCase switches the stylesheet from a click handler.
func themeCase() repaintCase {
	return repaintCase{
		name: "theme",
		newPage: func(t *testing.T) *Page {
			return newPage(t, Config{Title: "Theme", HTML: themeRepaintHTML, Width: 320, Height: 200},
				func(p *Page) {
					p.SetData(map[string]string{"Title": "Live theme"})
					p.Handle(Handlers{Click: func(_ context.Context, box Box) error {
						if box.ID == "toggle" {
							return p.SetTheme(themeRepaintCSS)
						}

						return nil
					}})
				})
		},
		interact: func(t *testing.T, p *Page, ctx context.Context) {
			clickID(t, p, ctx, "toggle")
		},
	}
}

const statesRepaintHTML = `<html><head><style>
#hover { width:200px;height:60px;margin:20px;background:#e5e7eb }
#hover:hover { background:#1a56db }
</style></head><body style="margin:0;background:#f4f1ea">
<div id="hover"><span>Hover me</span></div></body></html>`

// statesCase moves the pointer onto the hover target.
func statesCase() repaintCase {
	return repaintCase{
		name: "states",
		newPage: func(t *testing.T) *Page {
			return newPage(t, Config{Title: "States", HTML: statesRepaintHTML, Width: 320, Height: 160}, nil)
		},
		interact: func(t *testing.T, p *Page, ctx context.Context) {
			x, y := boxCenter(t, p, "hover")
			if err := p.Testing().Hover(ctx, x, y); err != nil {
				t.Fatal(err)
			}
		},
	}
}

const platformRepaintHTML = `<html><body style="margin:0;background:#f4f1ea">
<div style="width:300px;margin:20px;padding:16px;background:#fff">
<h1>Platform</h1><p id="count">{{.Count}}</p>
<button id="inc">Increment</button></div></body></html>`

// platformCase clicks the counter twice.
func platformCase() repaintCase {
	return repaintCase{
		name: "platform",
		newPage: func(t *testing.T) *Page {
			return newPage(t, Config{Title: "Platform", HTML: platformRepaintHTML, Width: 320, Height: 200},
				func(p *Page) {
					count := 0
					p.SetData(map[string]int{"Count": count})
					p.Handle(Handlers{Click: func(_ context.Context, box Box) error {
						if box.ID == "inc" {
							count++
							p.SetData(map[string]int{"Count": count})
						}

						return nil
					}})
				})
		},
		interact: func(t *testing.T, p *Page, ctx context.Context) {
			clickID(t, p, ctx, "inc")
			clickID(t, p, ctx, "inc")
		},
	}
}
