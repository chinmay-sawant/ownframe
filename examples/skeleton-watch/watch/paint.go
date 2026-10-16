package watch

import "github.com/chinmay-sawant/ownframe"

type watchParts struct {
	handHour    *ownframe.DisplayOp
	handMinute  *ownframe.DisplayOp
	handSecond  *ownframe.DisplayOp
	chronoSec   *ownframe.DisplayOp
	chronoMin   *ownframe.DisplayOp
	tourbillon  *ownframe.DisplayOp
	balance     *ownframe.DisplayOp
	powerHand   *ownframe.DisplayOp
	gearBarrel  *ownframe.DisplayOp
	gearCenter  *ownframe.DisplayOp
	gearThird   *ownframe.DisplayOp
	gearFourth  *ownframe.DisplayOp
	textClock   *ownframe.DisplayOp
	textChrono  *ownframe.DisplayOp
	textReserve *ownframe.DisplayOp
	textStatus  *ownframe.DisplayOp
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
	rotateHand(p.handHour, a.view.HourAngle)
	rotateHand(p.handMinute, a.view.MinuteAngle)
	rotateHand(p.handSecond, a.view.SecondAngle)
	rotateHand(p.chronoSec, a.view.ChronoSecAngle)
	rotateHand(p.chronoMin, a.view.ChronoMinAngle)
	rotateHand(p.powerHand, a.view.ReserveAngle)

	rotateCenter(p.tourbillon, a.view.TourbillonAngle)
	rotateCenter(p.balance, a.view.BalanceAngle)
	rotateCenter(p.gearBarrel, a.view.BarrelAngle)
	rotateCenter(p.gearCenter, a.view.CenterWheelAngle)
	rotateCenter(p.gearThird, a.view.ThirdWheelAngle)
	rotateCenter(p.gearFourth, a.view.FourthWheelAngle)

	if p.textClock != nil {
		p.textClock.Text = a.view.ClockText
	}
	if p.textChrono != nil {
		p.textChrono.Text = a.view.ChronoText
	}
	if p.textReserve != nil {
		p.textReserve.Text = FormatReserve(a.view.PowerHours, a.view.PowerPercent)
	}
	if p.textStatus != nil {
		p.textStatus.Text = a.view.Status
	}
}
