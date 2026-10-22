package remote

import (
	"context"
	"fmt"
	"testing"
)

func TestPhoneLayoutsKeepTouchTargets(t *testing.T) {
	for _, size := range [][2]int{{240, 320}, {320, 480}, {360, 640}, {390, 844}, {420, 800}, {800, 480}} {
		for _, font := range []int{16, 32, 48} {
			for _, light := range []bool{false, true} {
				for _, panel := range []string{"remote", "pad", "nums"} {
					t.Run(fmt.Sprintf("%dx%d/%d/%v/%s", size[0], size[1], font, light, panel), func(t *testing.T) {
						a := newTest(t, WithPhone(true))
						p := a.Page()
						p.SetSize(size[0], size[1])
						if light {
							a.toggleTheme()
						}
						a.view.FontSize = font
						a.view.Title = "Living room LG television"
						a.view.Status = "Connecting. Accept the pairing prompt on the television to continue."
						a.view.Panel = panel
						a.setData()
						if err := p.Redraw(context.Background()); err != nil {
							t.Fatal(err)
						}
						if !p.ViewLocked() || p.AllowScroll() {
							t.Fatal("phone allows scrolling or is not fitted")
						}
						for _, b := range p.Boxes() {
							if b.Tag != "button" {
								continue
							}
							if b.W < 47.9 || b.H < 47.9 {
								t.Errorf("%s target %.1fx%.1f", b.ID, b.W, b.H)
							}
							if b.X < -0.1 || b.X+b.W > float64(size[0])+.1 {
								t.Errorf("%s overflows: x %.1f w %.1f", b.ID, b.X, b.W)
							}
						}
					})
				}
			}
		}
	}
}
