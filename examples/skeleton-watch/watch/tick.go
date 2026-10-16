package watch

import (
	"context"
	"fmt"
	"time"
)

// Tick advances the watch train, runs escapement balance, and repaints.
func (a *App) Tick(ctx context.Context) error {
	a.page.UseFrameDirty()
	if err := ctx.Err(); err != nil {
		return err
	}

	now := a.now()
	dt := a.elapsedSinceLast(now)
	a.reserve.Drain(dt)
	stopped := a.reserve.Depleted()

	h, m, s := timeAngles(now)
	if stopped {
		h, m, s = a.view.HourAngle, a.view.MinuteAngle, a.view.SecondAngle
	}
	cage, bal := tourbillonAngles(now, stopped)
	cSec, cMin, cHr := chronoAngles(a.chrono.DisplayDuration(now))
	center, third, fourth, barrel := gearAngles(now, stopped)

	a.view.HourAngle = h
	a.view.MinuteAngle = m
	a.view.SecondAngle = s
	a.view.TourbillonAngle = cage
	a.view.BalanceAngle = bal
	a.view.ChronoSecAngle = cSec
	a.view.ChronoMinAngle = cMin
	a.view.ChronoHourAngle = cHr
	a.view.ReserveAngle = a.reserve.Angle()
	a.view.CenterWheelAngle = center
	a.view.ThirdWheelAngle = third
	a.view.FourthWheelAngle = fourth
	a.view.BarrelAngle = barrel

	a.view.PowerHours = a.reserve.Hours()
	a.view.PowerPercent = a.reserve.Percent()
	a.view.ClockText = now.Format("15:04:05")
	a.view.ChronoText = a.chrono.Format(now)

	if stopped {
		a.view.Status = "STOPPED (Depleted)"
	} else {
		a.view.Status = "RUNNING (28,800 vph)"
	}

	a.paint()
	return nil
}

func (a *App) elapsedSinceLast(now time.Time) float64 {
	if a.lastTick.IsZero() {
		a.lastTick = now
		return 0
	}
	dt := now.Sub(a.lastTick).Seconds()
	a.lastTick = now
	if dt < 0 {
		return 0
	}
	if dt > 0.1 {
		return 0.1
	}
	return dt
}

// FormatReserve returns the formatted string for power reserve readout.
func FormatReserve(hours float64, percent int) string {
	return fmt.Sprintf("%d%% (%.1fh)", percent, hours)
}
