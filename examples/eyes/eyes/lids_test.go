package eyes

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestHappyBottomLid(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)
	a.st.mood = 1
	for range 90 {
		tk(t, a, c, time.Second/60)
	}
	if a.pm.lidLB == nil || a.pm.lidLB.H <= 0 {
		t.Fatal("happy has no bottom lid")
	}
	if a.pm.lidLT != nil && a.pm.lidLT.H > 0 {
		t.Fatal("happy should not drop the top lid")
	}
	pp := a.page.Display().PointsPerPixel
	if a.pm.left.W/pp < a.pm.baseL.w-1 {
		t.Fatal("happy narrowed the eye; lids shape it")
	}
}

func TestSleepyTopLid(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)
	a.st.mood = 2
	for range 90 {
		tk(t, a, c, time.Second/60)
	}
	if a.pm.lidLT == nil || a.pm.lidLT.H <= 0 {
		t.Fatal("sleepy has no top lid")
	}
}

func TestIdleParksLids(t *testing.T) {
	a, c := newT(t)
	for range 10 {
		tk(t, a, c, time.Second/60)
	}
	for _, lid := range []*ownframe.DisplayOp{a.pm.lidLT, a.pm.lidLB, a.pm.lidRT, a.pm.lidRB} {
		if lid != nil && lid.H > 0 {
			t.Fatal("idle left a lid on the eye")
		}
	}
}
