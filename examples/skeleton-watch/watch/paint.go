package watch

type watchParts struct {
	handHour   partAssembly
	handMinute partAssembly
	handSecond partAssembly
	chronoMin  partAssembly
	powerHand  partAssembly
	tourbillon partAssembly
}

// paint updates display list operations in-place for 60fps native replay.
func (a *App) paint() {
	d := a.page.Display()
	if d == nil {
		return
	}
	if a.page.Generation() != a.bound {
		a.bind(d)
	}

	p := &a.parts
	p.handHour.rotate(a.view.HourAngle)
	p.handMinute.rotate(a.view.MinuteAngle)
	p.handSecond.rotate(a.view.ChronoSecAngle)
	p.chronoMin.rotate(a.view.ChronoMinAngle)
	p.powerHand.rotate(a.view.ReserveAngle)
	p.tourbillon.rotate(a.view.TourbillonAngle)
}
