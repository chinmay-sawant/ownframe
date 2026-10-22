package dino

// stepSky drifts the clouds and the ground marks left. The ready screen
// drifts too; a crash freezes the scene.
func (a *App) stepSky(dt float64) {
	if a.game.phase == over {
		return
	}

	for i := range a.clouds {
		a.clouds[i].x -= a.game.speed * cloudDrift * dt
		if a.clouds[i].x < -80 {
			a.clouds[i].x = sceneW + cloudRespawn + float64(i)*cloudGap
		}
	}

	for i := range a.pebbles {
		a.pebbles[i] -= a.game.speed * dt
		if a.pebbles[i] < -16 {
			a.pebbles[i] = sceneW + pebbleRespawn + float64(i)*pebbleGap
		}
	}
}
