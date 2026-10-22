package inputlab

import "fmt"

// send snapshots the plain form, like the old forms demo.
func (a *App) send() error {
	a.view.Status = fmt.Sprintf(
		"email=%s remember=%t plan=%s color=%s file=%s",
		a.page.Form().FormValue("f-email"),
		a.page.Form().FormChecked("f-remember"),
		a.fplan(),
		a.page.Form().FormValue("f-color"),
		a.page.Form().FormValue("f-doc"),
	)
	a.page.SetData(&a.view)
	return nil
}

func (a *App) fplan() string {
	if a.page.Form().FormChecked("f-pro") {
		return "pro"
	}
	if a.page.Form().FormChecked("f-free") {
		return "free"
	}
	return ""
}

// refresh prints focus plus the plain, edit, clip, and login state.
func (a *App) refresh() {
	a.view.Status = fmt.Sprintf(
		"focus=%s form=%s|%s plan=%s color=%s doc=%s note=%q clip=%s|%s login=%s",
		a.page.Edit().FocusedField(),
		a.page.Form().FormValue("f-email"),
		a.page.Form().FormValue("f-note"),
		a.fplan(),
		a.page.Form().FormValue("f-color"),
		a.page.Form().FormValue("f-doc"),
		a.page.Form().FormValue("e-note"),
		a.page.Form().FormValue("c-left"),
		a.page.Form().FormValue("c-right"),
		a.page.Form().FormValue("si-email"),
	)
	a.page.SetData(&a.view)
}

// ready reports whether both sign-in fields carry text.
func (a *App) ready() bool {
	return a.page.Form().FormValue("si-email") != "" &&
		a.page.Form().FormValue("si-password") != ""
}

// syncReady syncs View.Ready with the sign-in fields.
func (a *App) syncReady() {
	a.view.Ready = a.ready()
}

// signIn checks the demo secret/secret pair. Demo only: no auth
// boundary. DO NOT COPY into production; verify server side.
func (a *App) signIn() {
	if a.page.Form().FormValue("si-email") == "secret" &&
		a.page.Form().FormValue("si-password") == "secret" {
		a.view.LoginOk = "Signed in."
		a.view.LoginEr = ""
		return
	}
	a.view.LoginOk = ""
	a.view.LoginEr = "Unknown email or password."
}

// scrollTo moves the page; the window consumes ScrollTo.
