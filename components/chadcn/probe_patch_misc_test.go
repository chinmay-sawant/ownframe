package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func probeMiscRoot(el *ui.Element) *ui.Element {
	return ui.Div().ID("probe-root").Children(el)
}

// TestProbePatchMisc locks the incremental patch path for the misc group: a
// same-shape state edit must commit one generation without reparsing, and an
// identical rebuild must be a no-op.
//
// What stays FULL by design, and why:
//   - Toast closed->open mounts a surface from nil, so the trees cannot share
//     shape; description edits while open patch.
//   - Sidebar collapse narrows the nav box (192x18 to 48x18) and the engine
//     refuses a single-element relayout that changes size, so the class edit
//     falls back to a full load. No always-render variant avoids the width
//     change, so there is nothing to restructure.
//   - Attachment progress fill width edits change the fill box the same way
//     and fall back too; meta text edits patch.
//   - Trace add-row changes the child count and falls back; same-row detail
//     edits patch while the new text keeps the box (the detail node is
//     content-sized, so a reflow still falls back).
//   - CodeBlock code edits fall back because the code body must not take an
//     id (TestCodeBlockCopyButtonWiresOnCopy pins that).
func TestProbePatchMisc(t *testing.T) {
	t.Parallel()

	t.Run("toast-open-description-edit", func(t *testing.T) {
		t.Parallel()
		first := probeMiscRoot(chadcn.Toast(chadcn.ToastOptions{
			ID: "t-1", Open: true, Title: "Saved", Description: "All done",
		}))
		v, ctx := mountProbeView(t, first)
		next := probeMiscRoot(chadcn.Toast(chadcn.ToastOptions{
			ID: "t-1", Open: true, Title: "Saved", Description: "Done!",
		}))
		expectProbePatch(t, v, ctx, next)
	})

	t.Run("questionnaire-answer", func(t *testing.T) {
		t.Parallel()
		questions := func(selected string) []chadcn.Question {
			return []chadcn.Question{{
				Label:    "OS?",
				Selected: selected,
				Choices: []chadcn.ChoiceItem{
					{Value: "a", Label: "Linux"},
					{Value: "b", Label: "Mac"},
				},
			}}
		}
		v, ctx := mountProbeView(t, probeMiscRoot(chadcn.Questionnaire(
			chadcn.QuestionnaireOptions{ID: "q-1", Title: "Setup", Questions: questions("")})))
		expectProbePatch(t, v, ctx, probeMiscRoot(chadcn.Questionnaire(
			chadcn.QuestionnaireOptions{ID: "q-1", Title: "Setup", Questions: questions("a")})))
	})

	t.Run("trace-row-detail-edit", func(t *testing.T) {
		t.Parallel()
		rows := func(detail string) []chadcn.TraceRow {
			return []chadcn.TraceRow{
				{Label: "build", Detail: detail, Start: 0, Width: 40, Status: "ok"},
			}
		}
		v, ctx := mountProbeView(t, probeMiscRoot(chadcn.Trace(
			chadcn.TraceOptions{ID: "tr-1", Rows: rows("120 ms")})))
		expectProbePatch(t, v, ctx, probeMiscRoot(chadcn.Trace(
			chadcn.TraceOptions{ID: "tr-1", Rows: rows("121 ms")})))
	})

	t.Run("attachment-meta-edit", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeMiscRoot(chadcn.Attachment(
			chadcn.AttachmentOptions{ID: "att-1", Name: "shot.png", Meta: "2 MB", Progress: 10})))
		expectProbePatch(t, v, ctx, probeMiscRoot(chadcn.Attachment(
			chadcn.AttachmentOptions{ID: "att-1", Name: "shot.png", Meta: "12 MB", Progress: 10})))
	})

	t.Run("marker-detail-edit", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeMiscRoot(chadcn.Marker(
			chadcn.MarkerOptions{ID: "mk-1", Label: "Deploy", Detail: "queued"})))
		expectProbePatch(t, v, ctx, probeMiscRoot(chadcn.Marker(
			chadcn.MarkerOptions{ID: "mk-1", Label: "Deploy", Detail: "live"})))
	})

	t.Run("alert-description-edit", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeMiscRoot(chadcn.Alert(
			chadcn.AlertOptions{ID: "al-1", Title: "Heads up", Description: "v1"})))
		expectProbePatch(t, v, ctx, probeMiscRoot(chadcn.Alert(
			chadcn.AlertOptions{ID: "al-1", Title: "Heads up", Description: "v2"})))
	})

	t.Run("sidebar-identical-noop", func(t *testing.T) {
		t.Parallel()
		// Collapse itself stays a full render by design: the collapsed class
		// narrows the nav box (192x18 to 48x18) and the engine refuses a
		// single-element relayout that changes size, so PatchElements falls
		// back to a full load. The identical rebuild below locks that the
		// sidebar tree is at least stable when nothing changes.
		options := chadcn.SidebarOptions{ID: "sb-1", Open: true, Collapsed: false}
		v, ctx := mountProbeView(t, probeMiscRoot(chadcn.Sidebar(options)))
		expectProbeNoop(t, v, ctx, probeMiscRoot(chadcn.Sidebar(options)))
	})

	t.Run("message-text-edit", func(t *testing.T) {
		t.Parallel()
		message := func(text string) *ui.Element {
			return chadcn.Message(chadcn.MessageOptions{
				Name:     "Ada",
				Time:     "09:00",
				Children: []*ui.Element{ui.P().ID("msg-1").Text(text)},
			})
		}
		v, ctx := mountProbeView(t, probeMiscRoot(message("hello")))
		expectProbePatch(t, v, ctx, probeMiscRoot(message("hello edited")))
	})

	t.Run("breadcrumb-identical-noop", func(t *testing.T) {
		t.Parallel()
		items := []chadcn.BreadcrumbItem{{Label: "Home"}, {Label: "Orders"}, {Label: "#12", Current: true}}
		v, ctx := mountProbeView(t, probeMiscRoot(chadcn.Breadcrumb(items)))
		expectProbeNoop(t, v, ctx, probeMiscRoot(chadcn.Breadcrumb(items)))
	})

	t.Run("avatar-identical-noop", func(t *testing.T) {
		t.Parallel()
		options := chadcn.AvatarOptions{Initials: "AS"}
		v, ctx := mountProbeView(t, probeMiscRoot(chadcn.Avatar(options)))
		expectProbeNoop(t, v, ctx, probeMiscRoot(chadcn.Avatar(options)))
	})

	t.Run("codeblock-identical-noop", func(t *testing.T) {
		t.Parallel()
		options := chadcn.CodeBlockOptions{ID: "code-1", Filename: "main.go", Code: "a := 1"}
		v, ctx := mountProbeView(t, probeMiscRoot(chadcn.CodeBlock(options)))
		expectProbeNoop(t, v, ctx, probeMiscRoot(chadcn.CodeBlock(options)))
	})
}
