package inputlab

// snapshotClip reads both clipboard fields.
func (a *App) snapshotClip() csnap {
	return csnap{
		Left:  a.page.Form().FormValue("c-left"),
		Right: a.page.Form().FormValue("c-right"),
	}
}

// pushClip saves the clipboard pair with dedup.
func (a *App) pushClip() {
	a.lastEdit = "c"
	next := a.snapshotClip()
	if n := len(a.cUndo); n == 0 || a.cUndo[n-1] != next {
		a.cUndo = append(a.cUndo, next)
		if len(a.cUndo) > undoLimit {
			a.cUndo = a.cUndo[len(a.cUndo)-undoLimit:]
		}
	}
	a.cRedo = nil
}

func (a *App) restoreClip(s csnap) {
	a.page.Form().SetFormValue("c-left", s.Left)
	a.page.Form().SetFormValue("c-right", s.Right)
}

// snapshotLogin reads both sign-in fields.
func (a *App) snapshotLogin() lsnap {
	return lsnap{
		Email: a.page.Form().FormValue("si-email"),
		Pass:  a.page.Form().FormValue("si-password"),
	}
}

// pushLogin saves the sign-in pair with dedup.
func (a *App) pushLogin() {
	a.lastEdit = "si"
	next := a.snapshotLogin()
	if n := len(a.lUndo); n == 0 || a.lUndo[n-1] != next {
		a.lUndo = append(a.lUndo, next)
		if len(a.lUndo) > undoLimit {
			a.lUndo = a.lUndo[len(a.lUndo)-undoLimit:]
		}
	}
	a.lRedo = nil
}

func (a *App) restoreLogin(s lsnap) {
	a.page.Form().SetFormValue("si-email", s.Email)
	a.page.Form().SetFormValue("si-password", s.Pass)
}

func (a *App) undoLogin() error {
	if len(a.lUndo) == 0 {
		return nil
	}
	a.lRedo = append(a.lRedo, a.snapshotLogin())
	a.restoreLogin(a.lUndo[len(a.lUndo)-1])
	a.lUndo = a.lUndo[:len(a.lUndo)-1]
	a.syncReady()
	return nil
}

func (a *App) redoLogin() error {
	if len(a.lRedo) == 0 {
		return nil
	}
	a.lUndo = append(a.lUndo, a.snapshotLogin())
	a.restoreLogin(a.lRedo[len(a.lRedo)-1])
	a.lRedo = a.lRedo[:len(a.lRedo)-1]
	a.syncReady()
	return nil
}
