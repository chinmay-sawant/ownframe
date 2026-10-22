package flappy

// stepScene drifts the clouds and the ground marks left. The ready screen
// drifts too; a crash freezes the scene.
func (a *App) stepScene(dt float64) {
	if a.game.phase == over {
		return
	}

	drift := a.game.speed * dt

	for i := range a.clouds {
		a.clouds[i] -= drift * cloudDrift

		if a.clouds[i] < -90 {
			a.clouds[i] = sceneW + cloudResetX + float64(i)*cloudResetGap
		}
	}

	for i := range a.stripes {
		a.stripes[i] -= drift

		if a.stripes[i] < -50 {
			a.stripes[i] = sceneW + stripeResetX + float64(i)*stripeResetGap
		}
	}
}
