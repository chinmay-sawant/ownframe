package chadcn_test

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func TestChoiceValuesCountOnlyListedAnswers(t *testing.T) {
	t.Parallel()
	source, err := ui.Render(chadcn.Questionnaire(chadcn.QuestionnaireOptions{
		Title: "Pick",
		Questions: []chadcn.Question{
			{
				Label:    "Known",
				Selected: "yes",
				Choices: []chadcn.ChoiceItem{
					{Value: "yes", Label: "Yes"},
					{Value: "no", Label: "No"},
				},
			},
			{
				Label:    "Label is not a value",
				Selected: "Yes",
				Choices:  []chadcn.ChoiceItem{{Value: "yes", Label: "Yes"}},
			},
			{
				Label:    "Empty",
				Selected: "",
				Choices:  []chadcn.ChoiceItem{{Value: "yes", Label: "Yes"}},
			},
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, "1 of 3 answered") {
		t.Fatalf("progress counted a selection that is not a choice value:\n%s", source)
	}
	if strings.Contains(source, "2 of 3 answered") || strings.Contains(source, "3 of 3 answered") {
		t.Fatalf("progress counted an unmatched selection:\n%s", source)
	}
	if got := strings.Count(source, "of-questionnaire-choice--selected"); got != 1 {
		t.Fatalf("selected choices=%d, want 1:\n%s", got, source)
	}
}

func TestSliderValueAtStepsAndClamps(t *testing.T) {
	t.Parallel()
	opts := chadcn.SliderOptions{Min: 0, Max: 100, Step: 10, Value: 80, Disabled: true}
	cases := []struct {
		name  string
		x     float64
		width float64
		want  float64
	}{
		{name: "start", x: 0, width: 200, want: 0},
		{name: "middle", x: 100, width: 200, want: 50},
		{name: "snap", x: 30, width: 200, want: 20},
		{name: "past end", x: 250, width: 200, want: 100},
		{name: "before start", x: -10, width: 200, want: 0},
		{name: "empty track", x: 10, width: 0, want: 0},
		{name: "not a number", x: math.NaN(), width: 200, want: 0},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := chadcn.SliderValueAt(tt.x, tt.width, opts); got != tt.want {
				t.Fatalf("SliderValueAt(%v, %v) = %v, want %v", tt.x, tt.width, got, tt.want)
			}
		})
	}
}

func TestCalendarOutsideDaySkipsSelect(t *testing.T) {
	t.Parallel()
	month := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)
	source := renderSelectionHTML(t, chadcn.Calendar(chadcn.CalendarOptions{
		Month: month,
		OnSelect: func(context.Context, time.Time) error {
			return nil
		},
	}))
	var outside, insideAction int
	for tag := range strings.SplitSeq(source, "<button") {
		end := strings.Index(tag, ">")
		if end < 0 || !strings.Contains(tag[:end], "of-calendar-day") {
			continue
		}
		head := tag[:end]
		if strings.Contains(head, "of-calendar-day--outside") {
			outside++
			if strings.Contains(head, "data-action") {
				t.Fatalf("outside day is selectable:\n%s", head)
			}
			continue
		}
		if strings.Contains(head, "data-action") {
			insideAction++
		}
	}
	if outside == 0 {
		t.Fatal("November grid has no outside day")
	}
	if insideAction == 0 {
		t.Fatal("in-month days lost OnSelect")
	}
	rule := calendarFocusRule(t, chadcn.Stylesheet)
	if strings.Contains(rule, "outline: none") || !strings.Contains(rule, "outline: 2px solid") {
		t.Fatalf("calendar focus rule = %q", rule)
	}
}

func calendarFocusRule(t *testing.T, sheet string) string {
	t.Helper()
	const marker = ".of-calendar-day:focus-visible"
	idx := strings.Index(sheet, marker)
	if idx < 0 {
		t.Fatal("missing calendar day focus rule")
	}
	rest := sheet[idx:]
	end := strings.Index(rest, "}")
	if end < 0 {
		t.Fatal("calendar focus rule never closes")
	}
	return rest[:end]
}

func TestAreaChartNegativeSampleDrawsBand(t *testing.T) {
	t.Parallel()
	source := feedbackHTML(t, chadcn.AreaChart(chadcn.AreaChartOptions{
		Title: "Backlog",
		Series: []chadcn.ChartSeries{
			{Label: "Opened", Values: []float64{-8, -2}},
			{Label: "Resolved", Values: []float64{-3, -6}},
		},
	}))
	if got := strings.Count(source, `class="of-area-chart-band"`); got != 2 {
		t.Fatalf("negative columns drew %d bands, want 2:\n%s", got, source)
	}
}
