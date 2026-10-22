package ownframe

import (
	"context"
	"testing"
)

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
