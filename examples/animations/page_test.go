package main

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestNewApp(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}
	if app.Page() == nil {
		t.Fatal("expected non-nil page")
	}

	w, h := app.Page().Size()
	if w != defaultWidth || h != defaultHeight {
		t.Fatalf("expected size %dx%d, got %dx%d", defaultWidth, defaultHeight, w, h)
	}

	if app.view.Total != maxQuestions {
		t.Fatalf("expected total %d, got %d", maxQuestions, app.view.Total)
	}
	if app.view.Score != 0 {
		t.Fatalf("expected initial score 0, got %d", app.view.Score)
	}
	if !app.view.IsPlaying {
		t.Fatal("expected initial play state to be true")
	}
	if len(app.chips) == 0 {
		t.Fatal("expected registered chips for animation")
	}
}

func TestCSSAnimationBinding(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}
	d := app.Page().Probe().Display()
	if d == nil {
		t.Fatal("expected non-nil display")
	}

	var boundChip *chipSample
	for _, chip := range app.chips {
		if chip.id == "chip-q1-0" {
			boundChip = chip
			break
		}
	}
	if boundChip == nil {
		t.Fatal("expected chip-q1-0 in registered chips")
	}
	if boundChip.player == nil {
		t.Fatal("expected chip-q1-0 to have bound CSS animation player")
	}
	if len(boundChip.indices) == 0 {
		t.Fatal("expected chip-q1-0 to have captured operation indices")
	}
}

func TestHandleOptionClick(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	ctx := context.Background()

	// Question 1 correct answer is index 1 ("opt-q1-1", steps)
	err = app.handleClick(ctx, ownframe.Box{ID: "opt-q1-1"})
	if err != nil {
		t.Fatalf("click option failed: %v", err)
	}

	if app.view.Score != 1 {
		t.Fatalf("expected score 1 after correct answer, got %d", app.view.Score)
	}
	if app.view.AnsweredCount != 1 {
		t.Fatalf("expected answered count 1, got %d", app.view.AnsweredCount)
	}
	if !app.view.Questions[0].IsCorrect {
		t.Fatal("expected Question 1 to be marked correct")
	}

	// Repeated clicks on answered question should be ignored
	err = app.handleClick(ctx, ownframe.Box{ID: "opt-q1-0"})
	if err != nil {
		t.Fatalf("subsequent click failed: %v", err)
	}
	if app.view.Score != 1 {
		t.Fatalf("expected score to remain 1, got %d", app.view.Score)
	}

	// Question 2 wrong answer is index 0 ("opt-q2-0", correct is index 2)
	err = app.handleClick(ctx, ownframe.Box{ID: "opt-q2-0"})
	if err != nil {
		t.Fatalf("click wrong option failed: %v", err)
	}
	if app.view.Score != 1 {
		t.Fatalf("expected score to remain 1 after wrong answer, got %d", app.view.Score)
	}
	if app.view.AnsweredCount != 2 {
		t.Fatalf("expected answered count 2, got %d", app.view.AnsweredCount)
	}
	if app.view.Questions[1].IsCorrect {
		t.Fatal("expected Question 2 to be marked incorrect")
	}
}

func TestHandlePlayClick(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	ctx := context.Background()

	// Click play on track 1 of question 1
	err = app.handleClick(ctx, ownframe.Box{ID: "play-q1-1"})
	if err != nil {
		t.Fatalf("click play button failed: %v", err)
	}

	if !app.view.Questions[0].Tracks[1].IsActive {
		t.Fatal("expected Track 1 of Question 1 to be active")
	}
	if app.view.Questions[0].Tracks[0].IsActive {
		t.Fatal("expected Track 0 of Question 1 to be inactive")
	}
}

func TestTickAnimation(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	ctx := context.Background()

	// Initial tick advances CSS animation players
	err = app.Tick(ctx)
	if err != nil {
		t.Fatalf("tick failed: %v", err)
	}

	// Find the first bound chip with player
	var boundChip *chipSample
	for _, chip := range app.chips {
		if chip.player != nil && len(chip.indices) > 0 {
			boundChip = chip
			break
		}
	}

	if boundChip == nil {
		t.Fatal("expected at least one bound chip player")
	}

	d := app.Page().Probe().Display()
	startX := boundChip.opX(d)

	// Advance time artificially and step
	now := time.Now().Add(500 * time.Millisecond)
	boundChip.player.Step(now)

	newX := boundChip.opX(d)
	if newX == startX {
		t.Fatalf("expected chip operation X to change during animation step, got %f", newX)
	}
}

func TestTogglePlayback(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	ctx := context.Background()

	if !app.view.IsPlaying {
		t.Fatal("expected playback to start running")
	}

	// Toggle pause
	err = app.handleClick(ctx, ownframe.Box{ID: "btn-toggle-play"})
	if err != nil {
		t.Fatalf("click toggle play failed: %v", err)
	}
	if app.view.IsPlaying {
		t.Fatal("expected playback to be paused")
	}

	// Toggle resume
	err = app.handleClick(ctx, ownframe.Box{ID: "btn-toggle-play"})
	if err != nil {
		t.Fatalf("click toggle play failed: %v", err)
	}
	if !app.view.IsPlaying {
		t.Fatal("expected playback to be resumed")
	}
}

func TestResetQuiz(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	ctx := context.Background()

	// Answer question 1
	_ = app.handleClick(ctx, ownframe.Box{ID: "opt-q1-1"})
	if app.view.Score != 1 || app.view.AnsweredCount != 1 {
		t.Fatalf("expected answered question before reset")
	}

	// Click reset
	err = app.handleClick(ctx, ownframe.Box{ID: "btn-reset"})
	if err != nil {
		t.Fatalf("click reset failed: %v", err)
	}

	if app.view.Score != 0 {
		t.Fatalf("expected score 0 after reset, got %d", app.view.Score)
	}
	if app.view.AnsweredCount != 0 {
		t.Fatalf("expected answered count 0 after reset, got %d", app.view.AnsweredCount)
	}
	for i, q := range app.view.Questions {
		if q.Answered || q.SelectedIndex != -1 {
			t.Fatalf("expected question %d to be unselected after reset", i)
		}
	}
}

func TestCategoryFilter(t *testing.T) {
	app, err := newApp()
	if err != nil {
		t.Fatalf("newApp failed: %v", err)
	}

	ctx := context.Background()

	if app.view.ActiveFilter != "all" {
		t.Fatalf("expected default filter 'all', got %q", app.view.ActiveFilter)
	}

	err = app.handleClick(ctx, ownframe.Box{ID: "filter-timing"})
	if err != nil {
		t.Fatalf("click filter timing failed: %v", err)
	}
	if app.view.ActiveFilter != "timing" {
		t.Fatalf("expected active filter 'timing', got %q", app.view.ActiveFilter)
	}

	err = app.handleClick(ctx, ownframe.Box{ID: "filter-all"})
	if err != nil {
		t.Fatalf("click filter all failed: %v", err)
	}
	if app.view.ActiveFilter != "all" {
		t.Fatalf("expected active filter 'all', got %q", app.view.ActiveFilter)
	}
}
