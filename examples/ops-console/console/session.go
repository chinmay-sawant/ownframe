package console

func (a *App) signIn() {
	if !a.ready() {
		return
	}

	if a.view.Name != operatorName || !pinMatch(a.view.PIN.Reveal()) {
		a.view.Error = "Unknown operator or pin."
		a.view.Notice = ""

		return
	}

	a.view.Operator = a.view.Name
	a.view.Error = ""
	a.view.Notice = "Shift opened."
	a.view.Name = ""
	a.view.PIN = ""
	a.view.Ready = false
	a.page.Form().SetFormValue("name", "")
	a.page.Form().SetFormValue("pin", "")
	a.show(screenList)
}

func (a *App) signOut() {
	a.view.Operator = ""
	a.view.Error = ""
	a.view.Notice = ""
	a.view.Name = ""
	a.view.PIN = ""
	a.view.Ready = false
	a.view.Current = Ticket{}
	a.view.Note = ""
	a.page.Form().SetFormValue("name", "")
	a.page.Form().SetFormValue("pin", "")
	a.page.Form().SetFormValue("note", "")
	a.show(screenLogin)
}

func (a *App) show(screen string) {
	a.view.Screen = screen
	if screen != screenLogin {
		a.view.Error = ""
	}
}
