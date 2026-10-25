package main

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

const cardPaintTolerance = 1e-9

// Freeze each keyframe track at its start, midpoint, and end. Original timing
// is tested separately; this checks that every card's effect reaches paint.
func TestEveryCardKeyframeChangesPaint(t *testing.T) {
	poses := []map[string][]string{}
	for _, elapsed := range []int{0, 500, 1000} {
		theme := fmt.Sprintf(`.motion {
 animation-play-state: paused !important;
 animation-duration: 1000ms !important;
 animation-delay: -%dms !important;
 animation-direction: normal !important;
 animation-iteration-count: 1 !important;
 animation-fill-mode: both !important;
}`, elapsed)
		p, err := ownframe.New(ownframe.Config{
			HTML: animationsHTML, Theme: theme, Width: defaultWidth, Height: defaultHeight, RetainState: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		p.SetData(&ViewState{IsPlaying: true})
		if err := p.Redraw(context.Background()); err != nil {
			t.Fatal(err)
		}
		if err := p.Testing().Tick(context.Background()); err != nil {
			t.Fatal(err)
		}
		if motion := p.Probe().Motion(); motion.LastError != "" || motion.RestampError != "" || len(motion.Gaps) != 0 {
			t.Fatalf("pose %d did not sample cleanly: %+v", elapsed, motion)
		}
		pose := map[string][]string{}
		for _, id := range showcaseTargetIDs() {
			for _, op := range opsInBox(p.Probe().Display(), boxByID(t, p, id)) {
				pose[id] = append(pose[id], fmt.Sprintf("%v|%s|%g,%g,%g,%g|%g,%g,%g,%g|%g|%+v",
					op.Kind, op.Text, op.X, op.Y, op.W, op.H, op.R, op.G, op.B, op.Alpha, op.Opacity(), op.Transform()))
			}
		}
		poses = append(poses, pose)
	}
	for _, id := range showcaseTargetIDs() {
		t.Run(id, func(t *testing.T) {
			if len(poses[0][id]) == 0 {
				t.Fatal("card has no paint operations")
			}
			if reflect.DeepEqual(poses[0][id], poses[1][id]) && reflect.DeepEqual(poses[0][id], poses[2][id]) {
				t.Fatal("card's start, midpoint, and final keyframes paint identically")
			}
		})
	}
}

func TestLongSnippetsWrapWithinTheirBox(t *testing.T) {
	p := newTestPage(t)
	for _, id := range []string{"snippet-matrix3d", "snippet-webkit"} {
		box := boxByID(t, p, id)
		lines := map[float64]bool{}
		for _, op := range opsInBox(p.Probe().Display(), box) {
			if op.Kind != ownframe.DisplayOpText {
				continue
			}
			lines[op.Y] = true
			const pointsToPixels = 96.0 / 72.0
			if (op.X+op.W)*pointsToPixels > box.X+box.W+1 {
				t.Fatalf("%s text %q overflows its box", id, op.Text)
			}
		}
		if len(lines) < 2 {
			t.Fatalf("%s stayed on one line", id)
		}
	}
}

func TestSteppedCardHasFiveJumpsAcrossTheWholePass(t *testing.T) {
	p := newTestPage(t)
	box := boxByID(t, p, "show-steps")
	animation := box.Animations[0]
	if animation.TimingFunction != "steps(5, jump-end)" {
		t.Fatalf("step timing = %q", animation.TimingFunction)
	}
	for _, rule := range p.Probe().Display().Keyframes {
		if rule.Name == animation.Name {
			if len(rule.Keyframes) != 2 {
				t.Fatal("intermediate keyframes would repeat the five-step easing per interval")
			}
			return
		}
	}
	t.Fatal("step keyframes are missing")
}

func TestDepthCardsReturnSmoothly(t *testing.T) {
	p := newTestPage(t)
	for _, id := range []string{"show-rotate-x", "show-rotate-y", "show-translate-z", "show-matrix3d"} {
		animation := boxByID(t, p, id).Animations[0]
		if animation.Direction != "alternate" || !animation.IterationInfinite {
			t.Fatalf("%s does not return smoothly: %+v", id, animation)
		}
	}
	if strings.Contains(animationsHTML, "rotateX(110deg)") || strings.Contains(animationsHTML, "rotateY(110deg)") {
		t.Fatal("card labels still advertise a different angle from their keyframes")
	}
}

func pausedGallery(t *testing.T, theme string) *ownframe.Page {
	t.Helper()
	p, err := ownframe.New(ownframe.Config{
		HTML: animationsHTML, Theme: theme, Width: defaultWidth, Height: defaultHeight, RetainState: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData(&ViewState{IsPlaying: true})
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := p.Testing().Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDelayCardsPaintTheirAdvertisedInitialPose(t *testing.T) {
	p := pausedGallery(t, `.motion { animation-play-state: paused !important; }`)
	positive := requireFill(t, p, "show-delay-positive")
	negative := requireFill(t, p, "show-delay-negative")
	if math.Abs(positive.Transform().E) > cardPaintTolerance || math.Abs(positive.Opacity()-0.3) > cardPaintTolerance {
		t.Fatalf("positive delay does not hold its faded start: matrix=%+v opacity=%g", positive.Transform(), positive.Opacity())
	}
	const expectedMidpoint = 90 * 72.0 / 96.0
	if math.Abs(negative.Transform().E-expectedMidpoint) > cardPaintTolerance || math.Abs(negative.Opacity()-1) > cardPaintTolerance {
		t.Fatalf("negative delay did not start halfway: matrix=%+v opacity=%g", negative.Transform(), negative.Opacity())
	}
}

func TestFillCardsPaintBeforeAndAfterTheirRun(t *testing.T) {
	grey := [3]float64{71.0 / 255, 85.0 / 255, 105.0 / 255}
	purple := [3]float64{124.0 / 255, 58.0 / 255, 237.0 / 255}
	green := [3]float64{5.0 / 255, 150.0 / 255, 105.0 / 255}
	for _, phase := range []struct {
		name, delay string
		colours     [][3]float64
	}{
		{name: "before", delay: "100s", colours: [][3]float64{grey, grey, purple}},
		{name: "after", delay: "-100s", colours: [][3]float64{grey, green, green}},
	} {
		t.Run(phase.name, func(t *testing.T) {
			p := pausedGallery(t, `.motion { animation-play-state: paused !important; animation-delay: `+phase.delay+` !important; }`)
			for i, id := range []string{"show-fill-none", "show-fill-forwards", "show-fill-both"} {
				got := colourOf(requireFill(t, p, id))
				for channel := range got {
					if math.Abs(got[channel]-phase.colours[i][channel]) > cardPaintTolerance {
						t.Fatalf("%s %s paints %v, want %v", phase.name, id, got, phase.colours[i])
					}
				}
			}
		})
	}
}
