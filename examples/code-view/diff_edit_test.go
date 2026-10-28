package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func openWorkingDiff(t *testing.T, app *App, ctx context.Context, path string) {
	t.Helper()

	app.reloadChanges(ctx, false)
	app.setMode(ctx, "changes")

	if app.diffPath != path {
		app.selectDiff(path)
	}

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if !app.View().Diff.CanEdit {
		t.Fatalf("diff %q is not editable", path)
	}
}

func TestDiffEditLineKeepsColumn(t *testing.T) {
	dir, _ := historyRepo(t)

	lines := ""
	for i := 1; i <= 100; i++ {
		lines += ".row { color: red; }\n"
	}
	writeFile(t, dir, "brfs.css", lines)
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "add css")

	changed := ""
	for i := 1; i <= 100; i++ {
		if i == 82 {
			changed += ".row { color: blue; }\n"
		} else {
			changed += ".row { color: red; }\n"
		}
	}
	writeFile(t, dir, "brfs.css", changed)

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	openWorkingDiff(t, app, ctx, "brfs.css")

	btn, ok := boxByID(app.Page().Boxes(), "diff-edit-line-82")
	if !ok {
		t.Fatal("no diff-edit-line-82 button")
	}

	codeBody, ok := boxByID(app.Page().Boxes(), "code-body")
	if !ok {
		t.Fatal("no code-body")
	}

	const wantCol = 11
	x := codeBody.X + diffUnifiedText + float64(wantCol)*diffCellW
	y := btn.Y + btn.H/2

	if _, err := app.onPress(ctx, btn, x, y); err != nil {
		t.Fatal(err)
	}

	if !app.editing || app.editPath() != "brfs.css" {
		t.Fatalf("press editing=%v path=%q", app.editing, app.editPath())
	}

	line, col := offsetLineCol(app.edit.Text(), app.editCaret)
	if line != 81 || col != wantCol {
		t.Fatalf("press caret = %d,%d want 81,%d", line, col, wantCol)
	}

	if err := app.onClick(ctx, ownframe.Box{Action: "diff-edit-line:82"}); err != nil {
		t.Fatal(err)
	}

	line, col = offsetLineCol(app.edit.Text(), app.editCaret)
	if line != 81 || col != wantCol {
		t.Fatalf("click caret = %d,%d want 81,%d", line, col, wantCol)
	}
}

func TestDiffEditHeaderOpensFirstChange(t *testing.T) {
	dir, _ := historyRepo(t)

	writeFile(t, dir, "notes.txt", "one\ntwo\nthree\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "add notes")
	writeFile(t, dir, "notes.txt", "one\nTWO\nthree\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	openWorkingDiff(t, app, ctx, "notes.txt")

	if _, ok := boxByID(app.Page().Boxes(), "diff-edit"); !ok {
		t.Fatal("no diff-edit button")
	}

	if err := app.onClick(ctx, ownframe.Box{Action: "diff-edit"}); err != nil {
		t.Fatal(err)
	}

	if app.mode != "explorer" || !app.editing || app.editPath() != "notes.txt" {
		t.Fatalf("mode=%q editing=%v path=%q", app.mode, app.editing, app.editPath())
	}
}

func TestCtrlSSavesDirtyWithoutEditBuffer(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	app.openFile("main.go")
	app.toggleEdit()

	if err := app.Page().Edit().Type(ctx, "x"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "escape"); err != nil {
		t.Fatal(err)
	}

	if app.editing || !app.dirty["main.go"] {
		t.Fatalf("editing=%v dirty=%v", app.editing, app.dirty["main.go"])
	}

	if err := app.onKeyDown(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "s"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyUp(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if app.dirty["main.go"] {
		t.Fatal("Ctrl+S did not save the dirty file")
	}
}

func TestDiffEditReadOnlySources(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	app.setMode(ctx, "changes")
	app.diffSource = "commit"
	app.commitOpen = "deadbeef"
	app.editDiffFile(-1)

	if app.editing {
		t.Fatal("commit diff should not enter edit mode")
	}

	app.diffSource = "branch"
	app.editDiffFile(-1)

	if app.editing {
		t.Fatal("branch diff should not enter edit mode")
	}
}

func TestEditLineClickKeepsColumn(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n\nfunc main() { println(1) }\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	app.openFile("main.go")
	app.beginEditAt(2, 11)

	if err := app.onClick(ctx, ownframe.Box{Action: "edit-begin-line:2"}); err != nil {
		t.Fatal(err)
	}

	line, col := offsetLineCol(app.edit.Text(), app.editCaret)
	if line != 2 || col != 11 {
		t.Fatalf("click caret = %d,%d want 2,11", line, col)
	}
}

func TestDiffRedirectCentersTarget(t *testing.T) {
	dir := t.TempDir()
	lines := ""
	for i := 1; i <= 400; i++ {
		lines += "line content number\n"
	}
	writeFile(t, dir, "big.css", lines)

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	app.openFile("big.css")
	app.toggleEdit()
	app.beginEditAt(320, 11)
	app.centerCaretInView()

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	rows := app.View().Rows
	if len(rows) == 0 {
		t.Fatal("no rows")
	}

	index := -1
	for i, row := range rows {
		if row.No == 321 {
			index = i
			break
		}
	}

	if index < 0 {
		t.Fatalf("target 321 not visible in %d..%d", rows[0].No, rows[len(rows)-1].No)
	}

	if index == 0 || index == len(rows)-1 {
		t.Fatalf("target at edge index=%d len=%d", index, len(rows))
	}

	margin := len(rows) / 4
	if index < margin || index >= len(rows)-margin {
		t.Fatalf("target not centered: index=%d len=%d", index, len(rows))
	}
}

func TestDiffDelRowRedirectsToNearestNewLine(t *testing.T) {
	dir, _ := historyRepo(t)

	mk := func(n int) string {
		var b strings.Builder
		b.WriteString("<html><body>\n")
		for i := 1; i <= n; i++ {
			fmt.Fprintf(&b, "<div id=\"row-%d\">content %d</div>\n", i, i)
		}
		b.WriteString("</body></html>\n")
		return b.String()
	}
	mkDel := func(n int) string {
		var b strings.Builder
		b.WriteString("<html><body>\n")
		for i := 1; i <= n; i++ {
			if i >= 40 && i <= 44 {
				continue
			}
			fmt.Fprintf(&b, "<div id=\"row-%d\">content %d</div>\n", i, i)
		}
		b.WriteString("</body></html>\n")
		return b.String()
	}

	writeFile(t, dir, "editerror.html", mk(60))
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "add html")
	writeFile(t, dir, "editerror.html", mkDel(60))

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	openWorkingDiff(t, app, ctx, "editerror.html")

	btn, ok := boxByID(app.Page().Boxes(), "diff-edit-line-del-42")
	if !ok {
		t.Fatal("deleted old line 42 has no edit button")
	}

	codeBody, ok := boxByID(app.Page().Boxes(), "code-body")
	if !ok {
		t.Fatal("no code-body")
	}

	const wantCol = 11
	x := codeBody.X + diffUnifiedText + float64(wantCol)*diffCellW
	y := btn.Y + btn.H/2

	if _, err := app.onPress(ctx, btn, x, y); err != nil {
		t.Fatal(err)
	}

	if !app.editing || app.editPath() != "editerror.html" {
		t.Fatalf("press editing=%v path=%q", app.editing, app.editPath())
	}

	line, col := offsetLineCol(app.edit.Text(), app.editCaret)
	if line != 40 || col != wantCol {
		t.Fatalf("press caret = %d,%d want 40,%d", line, col, wantCol)
	}

	rows := app.View().Rows
	index := -1
	for i, row := range rows {
		if row.No == 41 {
			index = i
			break
		}
	}

	if index < 0 {
		t.Fatalf("target 41 not visible in %d..%d", rows[0].No, rows[len(rows)-1].No)
	}

	if index == 0 || index == len(rows)-1 {
		t.Fatalf("target at edge index=%d len=%d", index, len(rows))
	}
}

func TestDiffWrapTailRedirectKeepsColumn(t *testing.T) {
	dir, _ := historyRepo(t)

	longOld := "prefix " + strings.Repeat("a", 200)
	longNew := "prefix " + strings.Repeat("b", 200)
	writeFile(t, dir, "page.html", "<html>\n"+longOld+"\n</html>\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "add html")
	writeFile(t, dir, "page.html", "<html>\n"+longNew+"\n</html>\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	openWorkingDiff(t, app, ctx, "page.html")

	tailID := ""
	for _, row := range app.View().Diff.Rows {
		if strings.Contains(row.EditID, "-c") {
			tailID = row.EditID
			break
		}
	}

	if tailID == "" {
		t.Fatal("long changed line did not wrap into a clickable tail chunk")
	}

	btn, ok := boxByID(app.Page().Boxes(), tailID)
	if !ok {
		t.Fatalf("no box %s", tailID)
	}

	codeBody, ok := boxByID(app.Page().Boxes(), "code-body")
	if !ok {
		t.Fatal("no code-body")
	}

	const tailCol = 5
	x := codeBody.X + diffUnifiedText + float64(tailCol)*diffCellW
	y := btn.Y + btn.H/2

	if _, err := app.onPress(ctx, btn, x, y); err != nil {
		t.Fatal(err)
	}

	if !app.editing || app.editPath() != "page.html" {
		t.Fatalf("press editing=%v path=%q", app.editing, app.editPath())
	}

	line, col := offsetLineCol(app.edit.Text(), app.editCaret)
	if line != 1 {
		t.Fatalf("press caret line=%d want 1", line)
	}

	if col <= tailCol {
		t.Fatalf("tail click mapped to col %d, want past the first chunk", col)
	}
}
