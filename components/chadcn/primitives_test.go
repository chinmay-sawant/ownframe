package chadcn_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func renderPrimitive(t *testing.T, root *ui.Element) string {
	t.Helper()
	html, err := ui.Render(root)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return html
}

func wantPrimitiveHTML(t *testing.T, html string, parts ...string) {
	t.Helper()
	for _, part := range parts {
		if !strings.Contains(html, part) {
			t.Fatalf("rendered HTML missing %q:\n%s", part, html)
		}
	}
}

func TestCardParts(t *testing.T) {
	t.Parallel()
	html := renderPrimitive(t, chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Login to your account"),
			chadcn.CardDescription("Enter your email below."),
		),
		chadcn.CardContent(ui.Text("Body")),
		chadcn.CardFooter(ui.Text("Footer")),
	))
	wantPrimitiveHTML(t, html,
		`<section class="of-card">`,
		`<div class="of-card-header">`,
		`<h3 class="of-card-title">Login to your account</h3>`,
		`<p class="of-card-description">Enter your email below.</p>`,
		`<div class="of-card-content">`,
		`<div class="of-card-footer">`,
	)
}

func TestBadgeVariants(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, label, want string
		element           *ui.Element
	}{
		{"default", "Badge", "of-badge--default", chadcn.Badge("Badge")},
		{"secondary", "Secondary", "of-badge--secondary", chadcn.Badge("Secondary", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary})},
		{"outline", "Outline", "of-badge--outline", chadcn.Badge("Outline", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline})},
		{"danger", "Danger", "of-badge--danger", chadcn.Badge("Danger", chadcn.BadgeOptions{Variant: chadcn.BadgeDanger})},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			html := renderPrimitive(t, tc.element)
			wantPrimitiveHTML(t, html, `class="of-badge `+tc.want+`"`, `>`+tc.label+`</span>`)
		})
	}
}

func TestProgressSemantics(t *testing.T) {
	t.Parallel()
	const overMaximum = 150
	const belowMinimum = -5
	html := renderPrimitive(t, chadcn.Progress("Upload", overMaximum))
	wantPrimitiveHTML(t, html,
		`class="of-progress"`,
		`role="progressbar"`,
		`aria-label="Upload"`,
		`aria-valuemin="0"`,
		`aria-valuemax="100"`,
		`aria-valuenow="100"`,
		`class="of-progress-fill"`,
		`width:100%`,
	)
	low := renderPrimitive(t, chadcn.Progress("Upload", belowMinimum))
	wantPrimitiveHTML(t, low, `aria-valuenow="0"`, `width:0%`)
}

func TestSeparatorOrientation(t *testing.T) {
	t.Parallel()
	horizontal := renderPrimitive(t, chadcn.Separator(chadcn.SeparatorOptions{}))
	wantPrimitiveHTML(t, horizontal,
		`class="of-separator of-separator--horizontal"`,
		`role="separator"`,
		`aria-orientation="horizontal"`,
	)
	vertical := renderPrimitive(t, chadcn.Separator(chadcn.SeparatorOptions{Orientation: chadcn.SeparatorVertical}))
	wantPrimitiveHTML(t, vertical,
		`class="of-separator of-separator--vertical"`,
		`aria-orientation="vertical"`,
	)
}

func TestAspectRatioRatioBox(t *testing.T) {
	t.Parallel()
	defaultRatio := renderPrimitive(t, chadcn.AspectRatio(chadcn.AspectRatioOptions{}, ui.Text("Preview")))
	wantPrimitiveHTML(t, defaultRatio,
		`class="of-aspect-ratio"`,
		`padding-top:56.25%`,
		`class="of-aspect-ratio-content"`,
		`Preview`,
	)
	fallback := renderPrimitive(t, chadcn.AspectRatio(chadcn.AspectRatioOptions{Ratio: -1}))
	wantPrimitiveHTML(t, fallback, `padding-top:56.25%`)
	wide := renderPrimitive(t, chadcn.AspectRatio(chadcn.AspectRatioOptions{Ratio: 4.0 / 3.0}))
	wantPrimitiveHTML(t, wide, `padding-top:75%`)
}

func TestAvatarImageInitialsAndStatus(t *testing.T) {
	t.Parallel()
	image := renderPrimitive(t, chadcn.Avatar(chadcn.AvatarOptions{
		Source: "alice.png", Alt: "Alice", Size: chadcn.AvatarLarge, Status: chadcn.AvatarStatusOnline,
	}))
	wantPrimitiveHTML(t, image,
		`class="of-avatar of-avatar--large"`,
		`<img alt="Alice" class="of-avatar-image" src="alice.png">`,
		`class="of-avatar-status of-avatar-status--online"`,
		`Online`,
	)

	initials := renderPrimitive(t, chadcn.Avatar(chadcn.AvatarOptions{
		Initials: "AS", Status: chadcn.AvatarStatusAway,
	}))
	wantPrimitiveHTML(t, initials,
		`class="of-avatar of-avatar--medium"`,
		`role="img"`,
		`aria-label="AS"`,
		`aria-hidden="true"`,
		`class="of-avatar-fallback"`,
		`>AS</span>`,
		`of-avatar-status--away`,
		`Away`,
	)

	fallback := renderPrimitive(t, chadcn.Avatar(chadcn.AvatarOptions{}))
	wantPrimitiveHTML(t, fallback, `aria-label="?"`, `>?</span>`)
}

func TestKbdLabel(t *testing.T) {
	t.Parallel()
	html := renderPrimitive(t, chadcn.Kbd("Ctrl"))
	wantPrimitiveHTML(t, html, `<span class="of-kbd">Ctrl</span>`)
}

func TestSkeletonGeometry(t *testing.T) {
	t.Parallel()
	plain := renderPrimitive(t, chadcn.Skeleton(chadcn.SkeletonOptions{}))
	wantPrimitiveHTML(t, plain, `class="of-skeleton"`)
	if strings.Contains(plain, "width:") || strings.Contains(plain, "height:") {
		t.Fatalf("default skeleton should use CSS geometry: %s", plain)
	}
	custom := renderPrimitive(t, chadcn.Skeleton(chadcn.SkeletonOptions{Width: 64, Height: 64, Rounded: true}))
	wantPrimitiveHTML(t, custom, `class="of-skeleton of-skeleton--rounded"`, `width:64px`, `height:64px`)
}

func TestSpinnerStatus(t *testing.T) {
	t.Parallel()
	plain := renderPrimitive(t, chadcn.Spinner(chadcn.SpinnerOptions{}))
	wantPrimitiveHTML(t, plain,
		`class="of-spinner of-spinner--medium"`,
		`role="status"`,
		`aria-label="Loading"`,
	)
	large := renderPrimitive(t, chadcn.Spinner(chadcn.SpinnerOptions{Size: chadcn.SpinnerLarge, Label: "Saving"}))
	wantPrimitiveHTML(t, large, `class="of-spinner of-spinner--large"`, `aria-label="Saving"`)
}

func TestEmptySlots(t *testing.T) {
	t.Parallel()
	html := renderPrimitive(t, chadcn.Empty(chadcn.EmptyOptions{
		Icon:        ui.Text("[]"),
		Title:       "No Projects Yet",
		Description: "You haven't created any projects yet.",
		Actions:     []*ui.Element{ui.Text("Create Project"), ui.Text("Import Project")},
		Footer:      ui.Text("Learn More"),
	}))
	wantPrimitiveHTML(t, html,
		`class="of-empty"`,
		`class="of-empty-icon"`,
		`aria-hidden="true"`,
		`class="of-empty-content"`,
		`class="of-empty-title">No Projects Yet</h3>`,
		`class="of-empty-description">You haven`,
		`class="of-empty-actions"`,
		`Create Project`,
		`Import Project`,
		`class="of-empty-footer"`,
		`Learn More`,
	)
}

func TestItemSlots(t *testing.T) {
	t.Parallel()
	basic := renderPrimitive(t, chadcn.Item(chadcn.ItemOptions{
		Title:       "Basic Item",
		Description: "A simple item with title and description.",
		Actions:     ui.Text("Action"),
	}))
	wantPrimitiveHTML(t, basic,
		`class="of-item"`,
		`class="of-item-content"`,
		`class="of-item-title">Basic Item</h3>`,
		`class="of-item-description">A simple item with title and description.</p>`,
		`class="of-item-actions"`,
		`Action`,
	)
	slotted := renderPrimitive(t, chadcn.Item(chadcn.ItemOptions{
		Media:   ui.Text("i"),
		Content: ui.Text("Verified"),
	}))
	wantPrimitiveHTML(t, slotted,
		`class="of-item-media"`,
		`class="of-item-content"`,
		`Verified`,
	)
}

func TestLabelFor(t *testing.T) {
	t.Parallel()
	bound := renderPrimitive(t, chadcn.Label("Email", chadcn.LabelOptions{For: "email"}))
	wantPrimitiveHTML(t, bound, `<label class="of-label" for="email">Email</label>`)
	loose := renderPrimitive(t, chadcn.Label("Email"))
	if strings.Contains(loose, "for=") {
		t.Fatalf("unbound label wrote a for attribute: %s", loose)
	}
}

func TestTableParts(t *testing.T) {
	t.Parallel()
	html := renderPrimitive(t, chadcn.Table(
		chadcn.TableCaption("A list of your recent invoices."),
		chadcn.TableHeader(chadcn.TableRow(
			chadcn.TableHead("Invoice"),
			chadcn.TableHead("Amount"),
		)),
		chadcn.TableBody(chadcn.TableRow(
			chadcn.TableCell("INV001"),
			chadcn.TableCell("$250.00"),
		)),
		chadcn.TableFooter(chadcn.TableRow(
			chadcn.TableCell("Total"),
			chadcn.TableCell("$250.00"),
		)),
	))
	wantPrimitiveHTML(t, html,
		`<table class="of-table">`,
		`<caption class="of-table-caption">A list of your recent invoices.</caption>`,
		`<thead class="of-table-header">`,
		`<tbody class="of-table-body">`,
		`<tfoot class="of-table-footer">`,
		`<tr class="of-table-row">`,
		`class="of-table-head" scope="col"`,
		`class="of-table-cell"`,
		`INV001`,
		`Total`,
	)
}

func TestTypographyHelpers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		element *ui.Element
		want    string
	}{
		{"h1", chadcn.H1("Title"), `<h1 class="of-typography of-typography--h1">Title</h1>`},
		{"h2", chadcn.H2("Section"), `<h2 class="of-typography of-typography--h2">Section</h2>`},
		{"h3", chadcn.H3("Subsection"), `<h3 class="of-typography of-typography--h3">Subsection</h3>`},
		{"h4", chadcn.H4("Minor"), `<h4 class="of-typography of-typography--h4">Minor</h4>`},
		{"p", chadcn.P("Body"), `<p class="of-typography of-typography--p">Body</p>`},
		{"lead", chadcn.Lead("Intro"), `<p class="of-typography of-typography--lead">Intro</p>`},
		{"muted", chadcn.Muted("Note"), `<p class="of-typography of-typography--muted">Note</p>`},
		{"code", chadcn.InlineCode("go test"), `<code class="of-typography-inline-code">go test</code>`},
		{"quote", chadcn.Blockquote("Quoted"), `<blockquote class="of-typography-blockquote">Quoted</blockquote>`},
		{"list", chadcn.List("One", "Two"), `<ul class="of-typography-list">`},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wantPrimitiveHTML(t, renderPrimitive(t, tc.element), tc.want)
		})
	}
	list := renderPrimitive(t, chadcn.List("One", "Two"))
	wantPrimitiveHTML(t, list, `<li class="of-typography-list-item">One</li>`, `<li class="of-typography-list-item">Two</li>`)
}

func TestDirectionAttribute(t *testing.T) {
	t.Parallel()
	ltr := renderPrimitive(t, chadcn.Direction(chadcn.DirectionOptions{}, ui.Text("Left")))
	wantPrimitiveHTML(t, ltr, `class="of-direction"`, `dir="ltr"`, `Left`)
	rtl := renderPrimitive(t, chadcn.Direction(chadcn.DirectionOptions{Mode: chadcn.DirectionRTL}, ui.Text("Right")))
	wantPrimitiveHTML(t, rtl, `class="of-direction"`, `dir="rtl"`, `Right`)
}
