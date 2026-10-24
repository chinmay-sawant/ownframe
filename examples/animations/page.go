package main

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed animations.html
var animationsHTML string

const (
	defaultWidth  = 960
	defaultHeight = 740
	scoreUnit     = 1
	maxQuestions  = 6
	expectedParts = 3
)

// Option is a multiple-choice option for a quiz question.
type Option struct {
	Index       int
	Key         string
	Text        string
	TargetTrack int
}

// Track represents an animated track inside the live demonstration stage.
type Track struct {
	Index      int
	ChipID     string
	PlayBtnID  string
	ChipClass  string
	ChipLabel  string
	TrackLabel string
	Kind       string
	IsActive   bool
}

// Question models an interactive challenge testing CSS animation properties.
type Question struct {
	ID                 string
	Number             int
	Category           string
	CategoryLabel      string
	Title              string
	Prompt             string
	CardID             string
	Tracks             []Track
	Options            []Option
	SelectedIndex      int
	CorrectIndex       int
	CorrectAnswerLabel string
	Answered           bool
	IsCorrect          bool
	Explanation        string
	CSSDeclaration     string
}

// CatalogEntry documents an animation property from compatibility-matrix.md.
type CatalogEntry struct {
	Name     string
	Type     string
	Accepted string
	Behavior string
}

// ViewState is the complete template data model for the application.
type ViewState struct {
	Score         int
	Total         int
	AnsweredCount int
	IsPlaying     bool
	ActiveFilter  string
	Questions     []Question
	Catalog       []CatalogEntry
}

// App manages the CSS animation showcase and quiz. The automatic page runtime
// drives every chip from its declared @keyframes; the app only holds view state
// and click handlers.
type App struct {
	page *ownframe.Page
	view ViewState
}

// newApp constructs and initializes the animations showcase application.
func newApp() (*App, error) {
	app := &App{
		view: ViewState{
			Score:         0,
			Total:         maxQuestions,
			AnsweredCount: 0,
			IsPlaying:     true,
			ActiveFilter:  "all",
			Questions:     initialQuestions(),
			Catalog:       initialCatalog(),
		},
	}

	p, err := ownframe.New(ownframe.Config{
		Title:    "CSS Animation Knowledge Lab",
		HTML:     animationsHTML,
		Width:    defaultWidth,
		Height:   defaultHeight,
		DevTools: false,
	})
	if err != nil {
		return nil, fmt.Errorf("create animations page: %w", err)
	}

	app.page = p
	p.SetData(&app.view)
	if err := p.Redraw(context.Background()); err != nil {
		return nil, fmt.Errorf("initial redraw: %w", err)
	}

	p.Handle(ownframe.Handlers{
		Click: app.handleClick,
	})

	return app, nil
}

// Page returns the underlying ownframe.Page.
func (a *App) Page() *ownframe.Page {
	return a.page
}

// handleClick routes user interactions across buttons, filters, options, and tracks.
func (a *App) handleClick(ctx context.Context, box ownframe.Box) error {
	id := box.ID
	if id == "" {
		return nil
	}

	switch {
	case id == "btn-reset":
		a.resetQuiz()
	case id == "btn-toggle-play":
		a.view.IsPlaying = !a.view.IsPlaying
	case strings.HasPrefix(id, "filter-"):
		category := strings.TrimPrefix(id, "filter-")
		a.view.ActiveFilter = category
	case strings.HasPrefix(id, "opt-"):
		a.handleOptionClick(id)
	case strings.HasPrefix(id, "play-"):
		a.handlePlayClick(id)
	default:
		return nil
	}

	a.page.SetData(&a.view)
	return a.page.Redraw(ctx)
}

// resetQuiz clears all user answers and restores the initial view. The chips
// keep their declared animation; reset is quiz state, not a motion restart.
func (a *App) resetQuiz() {
	a.view.Score = 0
	a.view.AnsweredCount = 0
	for i := range a.view.Questions {
		q := &a.view.Questions[i]
		q.Answered = false
		q.IsCorrect = false
		q.SelectedIndex = -1
		for j := range q.Tracks {
			q.Tracks[j].IsActive = false
		}
	}
}

// handleOptionClick handles selection of an option for a question: opt-<qid>-<idx>.
func (a *App) handleOptionClick(id string) {
	parts := strings.Split(id, "-")
	if len(parts) != expectedParts {
		return
	}

	qid := parts[1]
	optIndex, err := strconv.Atoi(parts[2])
	if err != nil {
		return
	}

	for i := range a.view.Questions {
		q := &a.view.Questions[i]
		if q.ID != qid {
			continue
		}

		if q.Answered {
			return
		}

		q.SelectedIndex = optIndex
		q.Answered = true
		a.view.AnsweredCount++

		if optIndex == q.CorrectIndex {
			q.IsCorrect = true
			a.view.Score += scoreUnit
		} else {
			q.IsCorrect = false
		}

		if optIndex >= 0 && optIndex < len(q.Options) {
			targetTrack := q.Options[optIndex].TargetTrack
			a.activateTrack(qid, targetTrack)
		}
		break
	}
}

// handlePlayClick handles clicking a track's Play button: play-<qid>-<idx>.
func (a *App) handlePlayClick(id string) {
	parts := strings.Split(id, "-")
	if len(parts) != expectedParts {
		return
	}

	qid := parts[1]
	trackIndex, err := strconv.Atoi(parts[2])
	if err != nil {
		return
	}

	a.activateTrack(qid, trackIndex)
}

// activateTrack marks one track active, which the template expresses as the
// .active-track class the stylesheet already declares. The automatic runtime
// keeps each chip's phase across the redraw, so this does not restart the
// keyframes; a real replay needs animation-trigger, owned by animation phase 06
// (plans/v0.0.3/animations-and-transitions/animation/phase-06-timelines-ranges-triggers.md).
func (a *App) activateTrack(qid string, trackIndex int) {
	for i := range a.view.Questions {
		q := &a.view.Questions[i]
		if q.ID != qid {
			continue
		}
		for j := range q.Tracks {
			q.Tracks[j].IsActive = (j == trackIndex)
		}
		break
	}
}

// initialQuestions sets up 6 challenge categories covering compatibility matrix properties.
//
//nolint:mnd // the catalog data names its track and option indices explicitly
func initialQuestions() []Question {
	return []Question{
		{
			ID:                 "q1",
			Number:             1,
			Category:           "timing",
			CategoryLabel:      "Timing Functions",
			Title:              "Discrete Stepped Progression",
			Prompt:             "Which animation-timing-function value divides an animation into equal discrete jumps without smooth intermediate interpolation?",
			CardID:             "card-q1",
			SelectedIndex:      -1,
			CorrectIndex:       1,
			CorrectAnswerLabel: "B: steps(5, jump-end)",
			Explanation:        "steps(n, <jump-term>) calculates output as n discrete plateau values rather than continuous Bézier splines, perfect for sprite-sheet frames or ticking indicators.",
			CSSDeclaration:     "animation-timing-function: steps(5, jump-end);\n/* Accepts: linear | ease | ease-in | ease-out | ease-in-out | cubic-bezier(...) | steps(...) */",
			Tracks: []Track{
				{Index: 0, ChipID: "chip-q1-0", PlayBtnID: "play-q1-0", Kind: "ease", ChipClass: "chip-ease", ChipLabel: "EASE", TrackLabel: "ease-in-out (smooth sinusoidal curve)"},
				{Index: 1, ChipID: "chip-q1-1", PlayBtnID: "play-q1-1", Kind: "linear", ChipClass: "chip-linear", ChipLabel: "LINEAR", TrackLabel: "linear (constant rate velocity)"},
				{Index: 2, ChipID: "chip-q1-2", PlayBtnID: "play-q1-2", Kind: "bezier", ChipClass: "chip-bezier", ChipLabel: "BEZIER", TrackLabel: "cubic-bezier(0.68, -0.55, 0.27, 1.55) (overshoot bounce)"},
				{Index: 3, ChipID: "chip-q1-3", PlayBtnID: "play-q1-3", Kind: "steps", ChipClass: "chip-steps", ChipLabel: "STEPS", TrackLabel: "steps(5, jump-end) (discrete staircase jumps)"},
			},
			Options: []Option{
				{Index: 0, Key: "A", Text: "cubic-bezier(0.25, 0.1, 0.25, 1.0)", TargetTrack: 2},
				{Index: 1, Key: "B", Text: "steps(5, jump-end)", TargetTrack: 3},
				{Index: 2, Key: "C", Text: "linear", TargetTrack: 1},
			},
		},
		{
			ID:                 "q2",
			Number:             2,
			Category:           "direction",
			CategoryLabel:      "Direction & Loops",
			Title:              "Alternating Playback Cycles",
			Prompt:             "Which animation-direction value plays odd-numbered iterations in normal forward order and even-numbered iterations backward?",
			CardID:             "card-q2",
			SelectedIndex:      -1,
			CorrectIndex:       2,
			CorrectAnswerLabel: "C: alternate",
			Explanation:        "animation-direction: alternate reverses direction on every even cycle, creating seamless ping-pong loops without harsh boundary snaps.",
			CSSDeclaration:     "animation-direction: alternate;\nanimation-iteration-count: infinite;\n/* Accepts: normal | reverse | alternate | alternate-reverse */",
			Tracks: []Track{
				{Index: 0, ChipID: "chip-q2-0", PlayBtnID: "play-q2-0", Kind: "dir-normal", ChipClass: "chip-dir-normal", ChipLabel: "NORMAL", TrackLabel: "normal (resets to 0% at each cycle)"},
				{Index: 1, ChipID: "chip-q2-1", PlayBtnID: "play-q2-1", Kind: "dir-reverse", ChipClass: "chip-dir-reverse", ChipLabel: "REVERSE", TrackLabel: "reverse (plays 100% to 0%)"},
				{Index: 2, ChipID: "chip-q2-2", PlayBtnID: "play-q2-2", Kind: "dir-alternate", ChipClass: "chip-dir-alternate", ChipLabel: "ALT", TrackLabel: "alternate (ping-pongs forward and back)"},
			},
			Options: []Option{
				{Index: 0, Key: "A", Text: "normal", TargetTrack: 0},
				{Index: 1, Key: "B", Text: "reverse", TargetTrack: 1},
				{Index: 2, Key: "C", Text: "alternate", TargetTrack: 2},
			},
		},
		{
			ID:                 "q3",
			Number:             3,
			Category:           "delays",
			CategoryLabel:      "Delays",
			Title:              "Start Delays & Negative Offsets",
			Prompt:             "What visual effect occurs when animation-delay is set to a negative value such as -1.0s on a 2.0s duration animation?",
			CardID:             "card-q3",
			SelectedIndex:      -1,
			CorrectIndex:       0,
			CorrectAnswerLabel: "A: The animation begins immediately at the 50% midpoint",
			Explanation:        "A negative start delay instructs the engine to begin playback immediately as if the animation had already been running for the specified time.",
			CSSDeclaration:     "animation-delay: -1.0s;\nanimation-delay-start: -1.0s;\n/* Accepts signed <time> start with optional non-negative end <time> */",
			Tracks: []Track{
				{Index: 0, ChipID: "chip-q3-0", PlayBtnID: "play-q3-0", Kind: "delay-zero", ChipClass: "chip-delay-zero", ChipLabel: "0.0s", TrackLabel: "delay: 0s (starts at initial frame)"},
				{Index: 1, ChipID: "chip-q3-1", PlayBtnID: "play-q3-1", Kind: "delay-pos", ChipClass: "chip-delay-pos", ChipLabel: "+0.8s", TrackLabel: "delay: 0.8s (waits before starting)"},
				{Index: 2, ChipID: "chip-q3-2", PlayBtnID: "play-q3-2", Kind: "delay-neg", ChipClass: "chip-delay-neg", ChipLabel: "-1.0s", TrackLabel: "delay: -1.0s (starts midway instantly)"},
			},
			Options: []Option{
				{Index: 0, Key: "A", Text: "The animation begins immediately at the 50% midpoint", TargetTrack: 2},
				{Index: 1, Key: "B", Text: "The animation pauses for 1.0s before proceeding", TargetTrack: 1},
				{Index: 2, Key: "C", Text: "The playback speed doubles", TargetTrack: 0},
			},
		},
		{
			ID:                 "q4",
			Number:             4,
			Category:           "fill",
			CategoryLabel:      "Fill Modes",
			Title:              "Post-Completion Style Persistence",
			Prompt:             "Which animation-fill-mode ensures an element retains the computed styles of its final keyframe (100%) after the animation finishes?",
			CardID:             "card-q4",
			SelectedIndex:      -1,
			CorrectIndex:       1,
			CorrectAnswerLabel: "B: forwards",
			Explanation:        "animation-fill-mode: forwards holds the values defined in the last active keyframe after playback ends. 'both' applies forwards at the end and backwards during delays.",
			CSSDeclaration:     "animation-fill-mode: forwards;\n/* Accepts: none | forwards | backwards | both */",
			Tracks: []Track{
				{Index: 0, ChipID: "chip-q4-0", PlayBtnID: "play-q4-0", Kind: "fill-none", ChipClass: "chip-fill-none", ChipLabel: "NONE", TrackLabel: "none (snaps back to original styles on finish)"},
				{Index: 1, ChipID: "chip-q4-1", PlayBtnID: "play-q4-1", Kind: "fill-forwards", ChipClass: "chip-fill-forwards", ChipLabel: "FORW", TrackLabel: "forwards (freezes at final 100% position)"},
				{Index: 2, ChipID: "chip-q4-2", PlayBtnID: "play-q4-2", Kind: "fill-both", ChipClass: "chip-fill-both", ChipLabel: "BOTH", TrackLabel: "both (applies 0% in delay, holds 100% on finish)"},
			},
			Options: []Option{
				{Index: 0, Key: "A", Text: "none", TargetTrack: 0},
				{Index: 1, Key: "B", Text: "forwards", TargetTrack: 1},
				{Index: 2, Key: "C", Text: "backwards", TargetTrack: 2},
			},
		},
		{
			ID:                 "q5",
			Number:             5,
			Category:           "webkit",
			CategoryLabel:      "WebKit Prefixes",
			Title:              "WebKit Vendor Alias Compatibility",
			Prompt:             "How does the OwnFrame layout pipeline and Blinkless engine treat vendor -webkit-animation declarations?",
			CardID:             "card-q5",
			SelectedIndex:      -1,
			CorrectIndex:       1,
			CorrectAnswerLabel: "B: Mapped 1:1 and normalized to standard animation properties",
			Explanation:        "In Blinkless and OwnFrame, all 9 -webkit-animation properties are parsed, aliased to their standard property equivalents, and sampled consistently by the display pipeline.",
			CSSDeclaration:     "-webkit-animation: kf-webkit-hop 1.4s ease-in-out infinite alternate both;\n-webkit-animation-play-state: running;\n/* Group D Aliases: normalize to standard animation counterparts */",
			Tracks: []Track{
				{Index: 0, ChipID: "chip-q5-0", PlayBtnID: "play-q5-0", Kind: "ease", ChipClass: "chip-ease", ChipLabel: "STD", TrackLabel: "Standard: animation: kf-glide 2s ..."},
				{Index: 1, ChipID: "chip-q5-1", PlayBtnID: "play-q5-1", Kind: "webkit", ChipClass: "chip-webkit-sample", ChipLabel: "WEBKIT", TrackLabel: "WebKit: -webkit-animation: kf-webkit-hop 1.4s ..."},
			},
			Options: []Option{
				{Index: 0, Key: "A", Text: "Silently ignored as obsolete syntax", TargetTrack: 0},
				{Index: 1, Key: "B", Text: "Mapped 1:1 and normalized to standard animation properties", TargetTrack: 1},
				{Index: 2, Key: "C", Text: "Only executed on WebKit-based macOS browsers", TargetTrack: 1},
			},
		},
		{
			ID:                 "q6",
			Number:             6,
			Category:           "advanced",
			CategoryLabel:      "Advanced & Composite",
			Title:              "Animation Composition & Timeline Binding",
			Prompt:             "Which CSS property controls whether keyframe values completely replace base styles, add onto them, or accumulate values?",
			CardID:             "card-q6",
			SelectedIndex:      -1,
			CorrectIndex:       0,
			CorrectAnswerLabel: "A: animation-composition",
			Explanation:        "animation-composition defines the composite operation: 'replace' overrides earlier values, while 'add' or 'accumulate' combines keyframe transformations with underlying layout styles.",
			CSSDeclaration:     "animation-composition: replace;\nanimation-timeline: auto;\nanimation-range: entry 10% exit 90%;\nanimation-trigger: none;\n/* Accepts: replace | add | accumulate */",
			Tracks: []Track{
				{Index: 0, ChipID: "chip-q6-0", PlayBtnID: "play-q6-0", Kind: "composite", ChipClass: "chip-composite", ChipLabel: "COMP", TrackLabel: "animation-composition: replace; range: entry 10% exit 90%"},
			},
			Options: []Option{
				{Index: 0, Key: "A", Text: "animation-composition", TargetTrack: 0},
				{Index: 1, Key: "B", Text: "animation-range", TargetTrack: 0},
				{Index: 2, Key: "C", Text: "mix-blend-mode", TargetTrack: 0},
			},
		},
	}
}

// initialCatalog returns the 27 animation properties from documentation/compatibility-matrix.md.
func initialCatalog() []CatalogEntry {
	return []CatalogEntry{
		{Name: "animation", Type: "Shorthand", Accepted: "keyframe names with duration, easing, delay, iterations, direction, fill, state, timeline", Behavior: "Expands into longhands; stores records on layout.Display"},
		{Name: "animation-name", Type: "Longhand", Accepted: "none | custom-ident | quoted keyframe name", Behavior: "Matches authored @keyframes rule in the CSS cascade"},
		{Name: "animation-duration", Type: "Longhand", Accepted: "auto | non-negative <time>", Behavior: "Controls total single-cycle duration"},
		{Name: "animation-timing-function", Type: "Longhand", Accepted: "linear | ease | ease-in | ease-out | ease-in-out | cubic-bezier(...) | steps(...)", Behavior: "Easing curve applied between keyframe stops"},
		{Name: "animation-delay", Type: "Longhand", Accepted: "signed start <time> with optional non-negative end <time>", Behavior: "Shorthand for delay-start and delay-end"},
		{Name: "animation-delay-start", Type: "Longhand", Accepted: "signed <time>", Behavior: "Onset offset; negative values start playback mid-cycle"},
		{Name: "animation-delay-end", Type: "Longhand", Accepted: "non-negative <time>", Behavior: "Terminal hold delay before iteration resets"},
		{Name: "animation-iteration-count", Type: "Longhand", Accepted: "infinite | non-negative <number>", Behavior: "Total repetition count (e.g. 1, 2.5, infinite)"},
		{Name: "animation-direction", Type: "Longhand", Accepted: "normal | reverse | alternate | alternate-reverse", Behavior: "Iteration playback trajectory and inversion"},
		{Name: "animation-fill-mode", Type: "Longhand", Accepted: "none | forwards | backwards | both", Behavior: "Style application during pre-start delay and post-finish"},
		{Name: "animation-play-state", Type: "Longhand", Accepted: "running | paused", Behavior: "Freezes or advances the animation clock"},
		{Name: "animation-composition", Type: "Longhand", Accepted: "replace | add | accumulate", Behavior: "Controls combination with underlying base property styles"},
		{Name: "animation-timeline", Type: "Longhand", Accepted: "auto | none | <dashed-ident> | scroll() | view()", Behavior: "Timeline driver progress attachment"},
		{Name: "animation-range", Type: "Longhand", Accepted: "start timeline edge with optional end edge", Behavior: "Attaches animation progress to timeline edges"},
		{Name: "animation-range-start", Type: "Longhand", Accepted: "normal | timeline-range-name with <percentage>", Behavior: "Timeline progress onset threshold"},
		{Name: "animation-range-end", Type: "Longhand", Accepted: "normal | timeline-range-name with <percentage>", Behavior: "Timeline progress completion threshold"},
		{Name: "animation-range-center", Type: "Longhand", Accepted: "normal | source | target | <length-percentage>", Behavior: "Range calculation anchor reference"},
		{Name: "animation-trigger", Type: "Longhand", Accepted: "none | <dashed-ident> [play | pause | reset | replay]", Behavior: "Binds trigger actions to animation player"},
		{Name: "-webkit-animation", Type: "Vendor Alias", Accepted: "alias of standard animation shorthand", Behavior: "Normalizes 1:1 to animation in Blinkless pipeline"},
		{Name: "-webkit-animation-name", Type: "Vendor Alias", Accepted: "alias of animation-name", Behavior: "Normalizes 1:1 to animation-name"},
		{Name: "-webkit-animation-duration", Type: "Vendor Alias", Accepted: "alias of animation-duration", Behavior: "Normalizes 1:1 to animation-duration"},
		{Name: "-webkit-animation-timing-function", Type: "Vendor Alias", Accepted: "alias of animation-timing-function", Behavior: "Normalizes 1:1 to animation-timing-function"},
		{Name: "-webkit-animation-delay", Type: "Vendor Alias", Accepted: "alias of animation-delay", Behavior: "Normalizes 1:1 to animation-delay"},
		{Name: "-webkit-animation-iteration-count", Type: "Vendor Alias", Accepted: "alias of animation-iteration-count", Behavior: "Normalizes 1:1 to animation-iteration-count"},
		{Name: "-webkit-animation-direction", Type: "Vendor Alias", Accepted: "alias of animation-direction", Behavior: "Normalizes 1:1 to animation-direction"},
		{Name: "-webkit-animation-fill-mode", Type: "Vendor Alias", Accepted: "alias of animation-fill-mode", Behavior: "Normalizes 1:1 to animation-fill-mode"},
		{Name: "-webkit-animation-play-state", Type: "Vendor Alias", Accepted: "alias of animation-play-state", Behavior: "Normalizes 1:1 to animation-play-state"},
	}
}
