package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func renderChatComposite(t *testing.T, root *ui.Element) string {
	t.Helper()
	markup, err := ui.Render(root)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return markup
}

func assertChatMarkup(t *testing.T, markup string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(markup, want) {
			t.Errorf("missing %q in\n%s", want, markup)
		}
	}
}

func TestChatChartFrame(t *testing.T) {
	t.Parallel()
	chart := chadcn.Chart(chadcn.ChartOptions{
		Title:       "Bar Chart - Interactive",
		Description: "Showing total visitors for the last 3 months",
		Series: []chadcn.ChartSeries{
			{Label: "Desktop", Values: []float64{100, 200, 300}},
			{Label: "Mobile", Values: []float64{50, 150, 250}, Color: "#ef4444"},
		},
		Labels: []string{"Apr 2", "Apr 6", "Apr 10"},
		Total:  1050,
	})
	markup := renderChatComposite(t, chart)
	assertChatMarkup(t, markup,
		`class="of-chart"`,
		`class="of-chart-title"`,
		`class="of-chart-plot"`,
		`class="of-chart-legend"`,
		`data-series="1"`,
		`data-series="2"`,
		`--of-chart-bar:#ef4444`,
		"Bar Chart - Interactive",
		"Showing total visitors",
		"Desktop",
		"Mobile",
		"Apr 2",
		"Apr 10",
		"1,050",
		"600",
		"450",
		"height:100%",
		`aria-label="Bar chart: Bar Chart - Interactive. Showing total visitors for the last 3 months"`,
	)
	const bars = 6
	if got := strings.Count(markup, `class="of-chart-bar"`); got != bars {
		t.Fatalf("rendered %d bars, want %d", got, bars)
	}
}

func TestChatChartWithoutSeries(t *testing.T) {
	t.Parallel()
	markup := renderChatComposite(t, chadcn.Chart(chadcn.ChartOptions{Title: "Empty"}))
	assertChatMarkup(t, markup, `class="of-chart"`, "Empty")
	if strings.Contains(markup, "of-chart-plot") {
		t.Fatal("an empty chart should not render a plot")
	}
}

func TestChatCodeBlockEscapesCode(t *testing.T) {
	t.Parallel()
	block := chadcn.CodeBlock(chadcn.CodeBlockOptions{
		Filename: "app/api/chat/route.ts",
		Language: "ts",
		Code:     `export const node = <div>{value}</div>`,
	})
	markup := renderChatComposite(t, block)
	assertChatMarkup(t, markup,
		`class="of-code-block"`,
		`class="of-code-block-header"`,
		`class="of-code-block-filename"`,
		`class="of-code-block-language"`,
		`class="of-code-block-copy"`,
		`<pre class="of-code-block-body"><code class="of-code-block-code">`,
		"app/api/chat/route.ts",
		"ts",
		"Copy",
		"&lt;div&gt;",
	)
}

func TestChatMessageRows(t *testing.T) {
	t.Parallel()
	sent := chadcn.Message(chadcn.MessageOptions{
		Name:  "Pedro Duarte",
		Time:  "4:55 PM",
		Align: chadcn.MessageEnd,
		Children: []*ui.Element{
			chadcn.Bubble(chadcn.BubbleOptions{Variant: chadcn.BubblePrimary}, ui.Text("Deploying to prod")),
		},
	})
	markup := renderChatComposite(t, sent)
	assertChatMarkup(t, markup,
		`class="of-message of-message--end"`,
		`aria-label="Pedro Duarte"`,
		`class="of-message-avatar"`,
		">PD</div>",
		`class="of-message-name"`,
		`class="of-message-time"`,
		"4:55 PM",
		`class="of-bubble of-bubble--primary"`,
		"Deploying to prod",
	)
	if !strings.Contains(markup, "<article") {
		t.Fatalf("message is not an article:\n%s", markup)
	}

	avatar := ui.Span().ID("avatar").Text("A")
	received := chadcn.Message(chadcn.MessageOptions{
		Name:   "Ana Silva",
		Align:  chadcn.MessageStart,
		Avatar: avatar,
		Children: []*ui.Element{
			chadcn.Bubble(chadcn.BubbleOptions{}, ui.Text("On my way")),
		},
	})
	markup = renderChatComposite(t, received)
	assertChatMarkup(t, markup,
		`class="of-message of-message--start"`,
		`<div class="of-message-avatar"><span id="avatar">A</span></div>`,
	)
}

func TestChatBubbleVariants(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		variant chadcn.BubbleVariant
		class   string
	}{
		{"default", chadcn.BubbleDefault, "of-bubble--default"},
		{"primary", chadcn.BubblePrimary, "of-bubble--primary"},
		{"muted", chadcn.BubbleMuted, "of-bubble--muted"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			markup := renderChatComposite(t, chadcn.Bubble(chadcn.BubbleOptions{Variant: tt.variant}, ui.Text("Hi")))
			assertChatMarkup(t, markup, `class="of-bubble `+tt.class+`"`, "Hi")
		})
	}
}

func TestChatAttachmentTile(t *testing.T) {
	t.Parallel()
	tile := chadcn.Attachment(chadcn.AttachmentOptions{
		Name:     "sales-dashboard.pdf",
		Meta:     "Uploading · 64%",
		Progress: 64,
		Action:   "attachment:remove",
	})
	markup := renderChatComposite(t, tile)
	assertChatMarkup(t, markup,
		`class="of-attachment of-attachment--action"`,
		`data-action="attachment:remove"`,
		`class="of-attachment-icon"`,
		">PDF</div>",
		`class="of-attachment-name"`,
		"sales-dashboard.pdf",
		"Uploading · 64%",
		`role="progressbar"`,
		`aria-valuenow="64"`,
		"width:64%",
		"×",
	)
}

func TestChatAttachmentIconOverride(t *testing.T) {
	t.Parallel()
	tile := chadcn.Attachment(chadcn.AttachmentOptions{
		Name: "notes.txt",
		Icon: ui.Span().Class("host-icon").Text("N"),
	})
	markup := renderChatComposite(t, tile)
	assertChatMarkup(t, markup, `class="host-icon"`, ">N</span>")
	if strings.Contains(markup, ">TXT<") {
		t.Fatalf("host icon should replace the extension label:\n%s", markup)
	}
}

func TestChatMarkerStates(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		state chadcn.MarkerState
		class string
	}{
		{"default", chadcn.MarkerDefault, "of-marker--default"},
		{"success", chadcn.MarkerSuccess, "of-marker--success"},
		{"warning", chadcn.MarkerWarning, "of-marker--warning"},
		{"danger", chadcn.MarkerDanger, "of-marker--danger"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			markup := renderChatComposite(t, chadcn.Marker(chadcn.MarkerOptions{
				Label:  "Switched to a new branch",
				Detail: "feature/chat",
				State:  tt.state,
			}))
			assertChatMarkup(t, markup,
				`class="of-marker `+tt.class+`"`,
				`class="of-marker-icon"`,
				`class="of-marker-label"`,
				`class="of-marker-detail"`,
				"Switched to a new branch",
				"feature/chat",
			)
		})
	}
}

func TestChatMessageScroller(t *testing.T) {
	t.Parallel()
	scroller := chadcn.MessageScroller(chadcn.MessageScrollerOptions{
		Height: ui.Px(320),
		Children: []*ui.Element{
			chadcn.Message(chadcn.MessageOptions{
				Name:     "Ana",
				Children: []*ui.Element{chadcn.Bubble(chadcn.BubbleOptions{}, ui.Text("Hello"))},
			}),
		},
	})
	markup := renderChatComposite(t, scroller)
	assertChatMarkup(t, markup,
		`class="of-message-scroller"`,
		`class="of-message-scroller-viewport"`,
		`class="of-message-scroller-fade"`,
		`role="log"`,
		`aria-label="Messages"`,
		"height:320px",
		"Hello",
	)
}

func TestChatMessageScrollerDefaultHeight(t *testing.T) {
	t.Parallel()
	markup := renderChatComposite(t, chadcn.MessageScroller(chadcn.MessageScrollerOptions{}))
	if strings.Contains(markup, "style=") {
		t.Fatalf("zero height should keep the stylesheet height:\n%s", markup)
	}
}

func TestChatQuestionnaire(t *testing.T) {
	t.Parallel()
	questionnaire := chadcn.Questionnaire(chadcn.QuestionnaireOptions{
		Title:       "What should the agent build next?",
		Description: "Choose a direction or describe another task.",
		Questions: []chadcn.Question{
			{
				Label:    "What should the agent build next?",
				Help:     "Pick one direction.",
				Selected: "b",
				Choices: []chadcn.ChoiceItem{
					{Value: "a", Label: "Tool call timeline", Detail: "Show what the agent ran."},
					{Value: "b", Label: "Approval checkpoints", Detail: "Ask before destructive actions."},
				},
			},
			{
				Label:   "How should results look?",
				Choices: []chadcn.ChoiceItem{{Value: "x", Label: "Compact"}},
			},
		},
		OnSelect: func(context.Context, string, string) error { return nil },
	})
	markup := renderChatComposite(t, questionnaire)
	assertChatMarkup(t, markup,
		`class="of-questionnaire"`,
		`class="of-questionnaire-progress"`,
		`class="of-questionnaire-count"`,
		`class="of-questionnaire-label"`,
		`class="of-questionnaire-help"`,
		`role="radiogroup"`,
		`role="radio"`,
		`aria-checked="true"`,
		`aria-checked="false"`,
		`of-questionnaire-choice--selected`,
		"1 of 2 answered",
		"Question 1 of 2",
		"Question 2 of 2",
		"Tool call timeline",
		"Approval checkpoints",
		"Ask before destructive actions.",
		">A</span>",
		">B</span>",
		"data-action=",
	)
	if got := strings.Count(markup, `role="radio"`); got != 3 {
		t.Fatalf("rendered %d radio choices, want 3", got)
	}
	if got := strings.Count(markup, "of-questionnaire-choice--selected"); got != 1 {
		t.Fatalf("rendered %d selected choices, want 1", got)
	}
}

func TestChatTraceRows(t *testing.T) {
	t.Parallel()
	trace := chadcn.Trace(chadcn.TraceOptions{
		Rows: []chadcn.TraceRow{
			{Label: "Ran pnpm --version", Detail: "120 ms", Start: 0, Width: 40, Status: "ok"},
			{Label: "Failed to run pnpm install", Detail: "2.4 s", Start: 40, Width: 60, Status: "failed"},
			{Label: "Used 2 sources", Start: 70, Width: 60, Status: "warning"},
		},
	})
	markup := renderChatComposite(t, trace)
	assertChatMarkup(t, markup,
		`class="of-trace"`,
		`role="list"`,
		`role="listitem"`,
		`class="of-trace-row of-trace-row--success"`,
		`class="of-trace-row of-trace-row--danger"`,
		`class="of-trace-row of-trace-row--warning"`,
		`class="of-trace-track"`,
		`class="of-trace-bar"`,
		"Ran pnpm --version",
		"Failed to run pnpm install",
		"120 ms",
		`class="of-trace-gap"`,
		`style="flex-grow:40;"`,
		`style="flex-grow:60;"`,
	)
	// The warning row starts at 70%, so its bar clamps to the remaining track.
	assertChatMarkup(t, markup, `style="flex-grow:70;"`, `style="flex-grow:30;"`)
}

func TestChatCompositesZeroValue(t *testing.T) {
	t.Parallel()
	trees := map[string]*ui.Element{
		"chart":            chadcn.Chart(chadcn.ChartOptions{}),
		"code-block":       chadcn.CodeBlock(chadcn.CodeBlockOptions{}),
		"message":          chadcn.Message(chadcn.MessageOptions{}),
		"bubble":           chadcn.Bubble(chadcn.BubbleOptions{}),
		"attachment":       chadcn.Attachment(chadcn.AttachmentOptions{}),
		"marker":           chadcn.Marker(chadcn.MarkerOptions{}),
		"message-scroller": chadcn.MessageScroller(chadcn.MessageScrollerOptions{}),
		"questionnaire":    chadcn.Questionnaire(chadcn.QuestionnaireOptions{}),
		"trace":            chadcn.Trace(chadcn.TraceOptions{}),
	}
	for name, tree := range trees {
		t.Run(name, func(t *testing.T) {
			if _, err := ui.Render(tree); err != nil {
				t.Fatalf("render: %v", err)
			}
		})
	}
}
