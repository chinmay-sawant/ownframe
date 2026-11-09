package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func fix2StaticRoot(el *ui.Element) *ui.Element {
	return ui.Div().ID("probe-root").Children(el)
}

func fix2StaticCard() *ui.Element {
	return chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Login to your account"),
			chadcn.CardDescription("Enter your email below."),
		),
		chadcn.CardContent(ui.Text("Body")),
		chadcn.CardFooter(ui.Text("Footer")),
	)
}

func fix2StaticTable() *ui.Element {
	return chadcn.Table(
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
	)
}

func fix2StaticTypography() *ui.Element {
	return ui.Div().Children(
		chadcn.H1("Title"),
		chadcn.H2("Section"),
		chadcn.H3("Subsection"),
		chadcn.H4("Minor"),
		chadcn.P("Body"),
		chadcn.Lead("Intro"),
		chadcn.Muted("Note"),
		chadcn.InlineCode("go test"),
		chadcn.Blockquote("Quoted"),
		chadcn.List("One", "Two"),
	)
}

// TestProbePatchFix2Static locks that identical rebuilds of static chadcn
// components are Update no-ops (generation+0, no reparse).
func TestProbePatchFix2Static(t *testing.T) {
	t.Parallel()

	t.Run("empty-identical-noop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return fix2StaticRoot(chadcn.Empty(chadcn.EmptyOptions{
				Icon:        ui.Text("[]"),
				Title:       "No Projects Yet",
				Description: "You haven't created any projects yet.",
				Actions:     []*ui.Element{ui.Text("Create Project"), ui.Text("Import Project")},
				Footer:      ui.Text("Learn More"),
			}))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("item-identical-noop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return fix2StaticRoot(chadcn.Item(chadcn.ItemOptions{
				Title:       "Basic Item",
				Description: "A simple item with title and description.",
				Actions:     ui.Text("Action"),
			}))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("card-family-identical-noop", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, fix2StaticRoot(fix2StaticCard()))
		expectProbeNoop(t, v, ctx, fix2StaticRoot(fix2StaticCard()))
	})

	t.Run("table-family-identical-noop", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, fix2StaticRoot(fix2StaticTable()))
		expectProbeNoop(t, v, ctx, fix2StaticRoot(fix2StaticTable()))
	})

	t.Run("avatar-identical-noop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return fix2StaticRoot(chadcn.Avatar(chadcn.AvatarOptions{
				Source: "alice.png", Alt: "Alice", Size: chadcn.AvatarLarge, Status: chadcn.AvatarStatusOnline,
			}))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("kbd-identical-noop", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, fix2StaticRoot(chadcn.Kbd("Ctrl")))
		expectProbeNoop(t, v, ctx, fix2StaticRoot(chadcn.Kbd("Ctrl")))
	})

	t.Run("separator-identical-noop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return fix2StaticRoot(chadcn.Separator(chadcn.SeparatorOptions{}))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("skeleton-identical-noop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return fix2StaticRoot(chadcn.Skeleton(chadcn.SkeletonOptions{Width: 64, Height: 64, Rounded: true}))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("spinner-identical-noop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return fix2StaticRoot(chadcn.Spinner(chadcn.SpinnerOptions{Size: chadcn.SpinnerLarge, Label: "Saving"}))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("typography-identical-noop", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, fix2StaticRoot(fix2StaticTypography()))
		expectProbeNoop(t, v, ctx, fix2StaticRoot(fix2StaticTypography()))
	})

	t.Run("direction-identical-noop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return fix2StaticRoot(chadcn.Direction(
				chadcn.DirectionOptions{Mode: chadcn.DirectionRTL}, ui.Text("Right")))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("aspect-ratio-identical-noop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return fix2StaticRoot(chadcn.AspectRatio(
				chadcn.AspectRatioOptions{Ratio: 4.0 / 3.0}, ui.Text("Preview")))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("screen-panel-identical-noop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return fix2StaticRoot(chadcn.Screen(chadcn.Panel(ui.Text("Hello"))))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("notice-identical-noop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return fix2StaticRoot(chadcn.Notice("Heads up", "All systems go."))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("button-group-identical-noop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return fix2StaticRoot(chadcn.ButtonGroup(chadcn.ButtonGroupOptions{Label: "View"},
				chadcn.Button("List", chadcn.ButtonOptions{}),
				chadcn.Button("Grid", chadcn.ButtonOptions{}),
			))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("label-identical-noop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return fix2StaticRoot(chadcn.Label("Email", chadcn.LabelOptions{For: "email"}))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})
}
