package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// demoWindowStep is how long the live demo holds each scripted step. It is
// slow enough to watch, and longer than every interest delay in the fixture.
const demoWindowStep = 250 * time.Millisecond

// runDemoWindow opens the real OwnFrame window and plays the same script as
// -demo inside the window's own per-frame callback. The script sets the page
// scroll offset and feeds a scripted cursor, click and key to the same
// session.frame the real input uses. After each step it saves the page picture
// as a frame, then it closes the window and stitches the frames into an mp4.
// It never touches the operating system cursor or any other window.
func runDemoWindow(pagePath string, width, height int, dir string) error {
	html, err := os.ReadFile(pagePath)
	if err != nil {
		return err
	}

	if dir, err = demoDir(dir); err != nil {
		return err
	}

	fmt.Println("demo output:", dir)
	fmt.Println("opening one window; it closes itself when the script ends")

	ctx := context.Background()

	s, err := newSession(ctx, string(html), width, height, nil)
	if err != nil {
		return err
	}

	steps := demoScript(s)

	var (
		cur     demoStep
		n       = -1
		next    time.Time
		prev    = map[string]string{}
		done    bool
		clicked bool
		hit     string
	)

	// finish logs step n and saves its frame. s.last is the result of the
	// latest tick, so it holds the state the step settled on.
	finish := func() {
		state := s.host.describe(s.last.Frame, s.last.Frame.InterestOn)
		l := StepLog{N: n, Label: steps[n].Label, Scroll: int(s.viewport(steps[n].Scroll).ScrollY),
			Changes: diffState(prev, state), Clicked: clicked, Hit: hit}
		prev = state
		clicked, hit = false, ""

		fmt.Println(l.String())
	}

	s.attach(inputSource{
		cursor: func() (float64, float64) { return cur.CursorX, cur.CursorY },
		click: func() bool {
			c := cur.Click
			cur.Click = false // a click is one frame

			return c
		},
		keys: func() []string {
			k := cur.Keys
			cur.Keys = nil // a key press is one frame

			return k
		},
	})

	s.beforeFrame = func() {
		if done {
			return
		}

		if s.last.Clicked {
			clicked, hit = true, s.last.Hit
		}

		now := time.Now()
		if n >= 0 && now.Before(next) {
			s.page.WantFrameAt(next)

			return
		}

		if n >= 0 {
			finish()
		}

		n++
		if n >= len(steps) {
			done = true
			s.page.Window().Close()

			return
		}

		cur = steps[n]
		s.page.Scroll().SetScrollOffset(0, int(cur.Scroll))
		next = now.Add(demoWindowStep)
		s.page.WantFrameAt(next)
	}

	if err = ownframe.RunWithOptions(ctx, s.page, ownframe.WindowOptions{NoAudio: true}); err != nil {
		return err
	}

	// Saving a full-page picture per step is slow, so it waits until the window
	// is gone. These frames come from the same script through the same
	// session.frame, rendered with a manual clock.
	fmt.Println("window closed; rendering frames for the video")

	logs, err := runDemo(ctx, string(html), width, height, dir, true, nil)
	if err != nil {
		return err
	}

	fmt.Printf("%d frames written to %s\n", len(logs), dir)

	mp4, why := stitch(dir)
	if mp4 == "" {
		fmt.Println(why)

		return nil
	}

	fmt.Println("video:", mp4)

	return nil
}

// demoDir returns dir, creating it, or a fresh temp directory when empty.
func demoDir(dir string) (string, error) {
	if dir == "" {
		return os.MkdirTemp("", "blinkless-scroll-demo-")
	}

	return dir, os.MkdirAll(dir, 0o755)
}
