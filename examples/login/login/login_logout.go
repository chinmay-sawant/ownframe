package login

import osclip "github.com/chinmay-sawant/ownframe/internal/clipboard"

// Logout clears the fields, the undo and redo stacks, and the clipboard.
// Demo only; secret/secret stays the test login. DO NOT COPY.
func (a *App) Logout() {
	a.page.SetFormValue("email", "")
	a.page.SetFormValue("password", "")
	a.undo = nil
	a.redo = nil
	a.view = View{}
	a.page.SetData(a.view)
	osclip.Clear()
}
