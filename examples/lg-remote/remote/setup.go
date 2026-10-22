package remote

import "github.com/chinmay-sawant/ownframe"

func (a *App) installPage() {
	a.page.Scroll().SetAllowScroll(!a.phone)
	a.installIcons()
	a.installMedia()
	a.page.Handle(ownframe.Handlers{
		Click:     a.onClick,
		Press:     a.onPress,
		KeyDown:   a.onKey,
		DragStart: a.padStart,
		DragMove:  a.movePad,
	})
	a.page.SetTick(a.onTick)
	a.setData()

}
