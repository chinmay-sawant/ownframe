package remote

import (
	"fmt"
	"math"
)

func (a *App) setSensitivity(value float64) {
	a.view.Sensitivity = clampInt(int(math.Round(value/sensitivityStep))*sensitivityStep, sensitivityMin, sensitivityMax)
	a.view.SensitivityLabel = fmt.Sprintf("%.2fx", float64(a.view.Sensitivity)/percentScale)
	a.view.SensitivityFill = (a.view.Sensitivity - sensitivityMin) * percentScale / sensitivityRange
}

func (a *App) sensitivityAt(x float64) {
	for _, b := range a.page.Boxes() {
		if b.ID == "sensitivity" && b.W > 0 {
			a.sliderPosition = math.Max(sensitivityMin, math.Min(sensitivityMax, sensitivityMin+(x-b.X)*sensitivityRange/b.W))
			a.setSensitivity(a.sliderPosition)
			return
		}
	}
}

func (a *App) moveSensitivity(dx float64) {
	for _, b := range a.page.Boxes() {
		if b.ID == "sensitivity" && b.W > 0 {
			a.sliderPosition = math.Max(sensitivityMin, math.Min(sensitivityMax, a.sliderPosition+dx*sensitivityRange/b.W))
			a.setSensitivity(a.sliderPosition)
			a.setData()
			return
		}
	}
}
