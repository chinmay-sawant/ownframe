package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
)

func (r *rig) openInlineEditor(row, col int) {
	r.t.Helper()
	id := fmt.Sprintf("cell-%d-%d", row, col)
	r.a.lastClickID = ""
	r.tap(id)
	r.now = r.now.Add(60 * time.Millisecond)
	r.tap(id)
	if r.a.activeTabPtr().editing == nil {
		r.t.Fatalf("double-click on %s did not open inline editor", id)
	}
}

func TestInlineEditorLayoutNoWrapOrOverflow(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openTableByName("users")
	countryCol := r.colIndex("country")
	r.openInlineEditor(0, countryCol)

	cellID := fmt.Sprintf("cell-0-%d", countryCol)
	cellBox := r.box(cellID)
	inputBox := r.box("cell-editor")

	// Input must stay inside the cell bounds and never exceed cell height.
	if inputBox.X < cellBox.X || inputBox.X+inputBox.W > cellBox.X+cellBox.W {
		t.Fatalf("input horizontal overflow: input=%+v cell=%+v", inputBox, cellBox)
	}
	if inputBox.H > cellBox.H {
		t.Fatalf("input taller than cell: input.H=%v cell.H=%v", inputBox.H, cellBox.H)
	}
	// NOT NULL column (users.country) must not show the Set NULL button.
	if r.has("cell-set-null") {
		t.Fatal("NOT NULL column should not render cell-set-null button")
	}

	// Nullable column (audit_events.user_id) shows compact NULL button without overflow.
	r.openTableByName("audit_events")
	uidCol := r.colIndex("user_id")
	r.openInlineEditor(0, uidCol)

	uidCell := r.box(fmt.Sprintf("cell-0-%d", uidCol))
	uidInput := r.box("cell-editor")
	nullBtn := r.box("cell-set-null")

	if uidInput.W <= 0 || nullBtn.W <= 0 {
		t.Fatalf("nullable inline editor missing geometry: input=%+v null=%+v", uidInput, nullBtn)
	}
	if uidInput.X+uidInput.W > nullBtn.X+1 {
		t.Fatalf("input overlaps NULL button: input=%+v null=%+v", uidInput, nullBtn)
	}
	if nullBtn.X+nullBtn.W > uidCell.X+uidCell.W {
		t.Fatalf("NULL button overflows cell: null=%+v cell=%+v", nullBtn, uidCell)
	}
	if nullBtn.H > uidCell.H {
		t.Fatalf("NULL button taller than cell: null.H=%v cell.H=%v", nullBtn.H, uidCell.H)
	}
}

func TestInlineEditValidationBoolean(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openTableByName("users")
	tab := r.a.activeTabPtr()
	activeCol := r.colIndex("active")

	// Double-clicking a BOOLEAN column opens a true/false dropdown (no free-text input).
	r.openInlineEditor(0, activeCol)
	if r.has("cell-editor") {
		t.Fatal("BOOLEAN column must open a dropdown, not a free-text input")
	}
	if !r.has("cell-bool-menu") || !r.has("cell-bool-true") || !r.has("cell-bool-false") {
		t.Fatal("BOOLEAN dropdown must show true and false options")
	}
	if r.has("cell-bool-null") {
		t.Fatal("NOT NULL BOOLEAN column must not show NULL option")
	}
	r.tap("cell-bool-false")
	if tab.editing != nil {
		t.Fatal("picking false should close boolean dropdown")
	}
	v, staged := r.a.stagedValue(tab, 0, activeCol)
	if !staged || v.Kind != driver.KindBool || v.B != false {
		t.Fatalf("picking false: staged=%v val=%+v, want Bool(false)", staged, v)
	}

	// Cell Inspector also allows selecting true/false directly.
	r.tap("tbl-inspector")
	if !r.has("insp-bool-true") || !r.has("insp-bool-false") {
		t.Fatal("Cell Inspector must show true/false controls for BOOLEAN cells")
	}
	r.tap("insp-bool-true")
	if tab.cs != nil && tab.cs.Pending() > 0 {
		t.Fatalf("reverting active back to true should clear pending changes, got %d", tab.cs.Pending())
	}

	// parseCell still strictly validates boolean text values.
	boolCol := driver.ColumnMeta{Name: "active", Type: "BOOLEAN"}
	for _, bad := range []string{"maybe", "2", "truee", "", "abc123"} {
		if _, err := parseCell(bad, boolCol, driver.Bool(true)); err == nil {
			t.Fatalf("parseCell should reject invalid boolean %q", bad)
		}
	}
	for _, tc := range []struct {
		in   string
		want bool
	}{{"false", false}, {"  TRUE ", true}, {"0", false}, {"1", true}, {"f", false}, {"t", true}, {"no", false}, {"YES", true}} {
		got, err := parseCell(tc.in, boolCol, driver.Null())
		if err != nil || got.Kind != driver.KindBool || got.B != tc.want {
			t.Fatalf("parseCell(%q) = (%+v, %v), want Bool(%v)", tc.in, got, err, tc.want)
		}
	}
}

func TestInlineEditValidationIntegerAndPreviouslyNull(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openTableByName("products")
	tab := r.a.activeTabPtr()
	priceCol := r.colIndex("price_cents")

	for _, bad := range []string{"12abc", "3.14", "abc", ""} {
		r.openInlineEditor(0, priceCol)
		r.a.page.Form().SetFormValue("cell-editor", bad)
		r.must(r.a.onSubmit(r.ctx))
		if tab.cs != nil {
			t.Fatalf("invalid integer %q should not be staged", bad)
		}
		if tab.editing == nil || !tab.editing.invalid {
			t.Fatalf("invalid integer %q should keep editor open", bad)
		}
		if !strings.Contains(strings.ToLower(r.a.toast), "invalid integer") {
			t.Fatalf("expected integer error toast for %q, got %q", bad, r.a.toast)
		}
		r.key("escape")
	}

	r.openInlineEditor(0, priceCol)
	r.a.page.Form().SetFormValue("cell-editor", " 4200 ")
	r.must(r.a.onSubmit(r.ctx))
	v, ok := r.a.stagedValue(tab, 0, priceCol)
	if !ok || v.Kind != driver.KindInt || v.I != 4200 {
		t.Fatalf("valid integer edit failed: ok=%v v=%+v", ok, v)
	}

	// Previously NULL INTEGER cell in audit_events (row 49 has user_id = NULL since id=50 % 50 == 0).
	r.openTableByName("audit_events")
	atab := r.a.activeTabPtr()
	uidCol := r.colIndex("user_id")
	nullRow := -1
	for i, row := range atab.grid.rows {
		if row[uidCol].IsNull() {
			nullRow = i
			break
		}
	}
	if nullRow < 0 {
		t.Fatal("expected a NULL user_id row in audit_events")
	}
	atab.grid.scrollY = float64(nullRow * gridRowHeight)
	r.must(r.a.redraw(r.ctx))
	r.openInlineEditor(nullRow, uidCol)
	r.a.page.Form().SetFormValue("cell-editor", "notanint")
	r.must(r.a.onSubmit(r.ctx))
	if atab.cs != nil {
		t.Fatal("invalid integer on previously NULL cell must not be staged")
	}
	if atab.editing == nil || !atab.editing.invalid {
		t.Fatal("editor should remain open on invalid edit of NULL INTEGER cell")
	}
	r.a.page.Form().SetFormValue("cell-editor", "7")
	r.must(r.a.onSubmit(r.ctx))
	v, ok = r.a.stagedValue(atab, nullRow, uidCol)
	if !ok || v.Kind != driver.KindInt || v.I != 7 {
		t.Fatalf("valid integer on previously NULL cell failed: ok=%v v=%+v", ok, v)
	}
}

func TestNotNullEnforcement(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openTableByName("users")
	tab := r.a.activeTabPtr()
	activeCol := r.colIndex("active")

	// Context menu must disable g-null on NOT NULL column.
	items := r.a.contextMenuFor(fmt.Sprintf("cell-0-%d", activeCol))
	for _, it := range items {
		if it.ID == "g-null" && it.Enabled {
			t.Fatal("g-null should be disabled for NOT NULL column users.active")
		}
	}

	// Direct g-null action must still be rejected by stageCell.
	r.must(r.a.menuAction(r.ctx, "g-null"))
	if tab.cs != nil {
		t.Fatal("setting NULL on NOT NULL column must not stage a change")
	}
	if !strings.Contains(r.a.toast, "cannot be NULL") {
		t.Fatalf("expected NOT NULL toast, got %q", r.a.toast)
	}

	// Nullable column (audit_events.user_id) allows Set NULL.
	r.openTableByName("audit_events")
	atab := r.a.activeTabPtr()
	uidCol := r.colIndex("user_id")
	items = r.a.contextMenuFor(fmt.Sprintf("cell-0-%d", uidCol))
	foundEnabled := false
	for _, it := range items {
		if it.ID == "g-null" && it.Enabled {
			foundEnabled = true
		}
	}
	if !foundEnabled {
		t.Fatal("g-null should be enabled for nullable column audit_events.user_id")
	}
	r.must(r.a.menuAction(r.ctx, "g-null"))
	v, ok := r.a.stagedValue(atab, 0, uidCol)
	if !ok || !v.IsNull() {
		t.Fatalf("nullable column Set NULL failed: ok=%v v=%+v", ok, v)
	}
}

func TestJSONValidationInlineAndInspector(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openTableByName("audit_events")
	tab := r.a.activeTabPtr()
	payloadCol := r.colIndex("payload")

	// Double-clicking a JSON cell opens the Cell Inspector instead of an overflowing inline input.
	id := fmt.Sprintf("cell-0-%d", payloadCol)
	r.a.lastClickID = ""
	r.tap(id)
	r.now = r.now.Add(60 * time.Millisecond)
	r.tap(id)
	if tab.editing != nil {
		t.Fatal("double-clicking a JSON cell must open the Cell Inspector, not an inline input")
	}
	if !tab.inspectorOpen {
		t.Fatal("double-clicking a JSON cell must open the Cell Inspector")
	}

	// Inspector textarea must be full height (not crushed to 28px) and never overlap Apply.
	editBox := r.box("insp-edit")
	applyBox := r.box("insp-apply")
	if editBox.H < 100 {
		t.Fatalf("insp-edit height = %v, want >= 100 (must not be crushed by .qr-input)", editBox.H)
	}
	if editBox.Y+editBox.H > applyBox.Y+0.5 {
		t.Fatalf("insp-edit (%v..%v) overlaps insp-apply (%v)", editBox.Y, editBox.Y+editBox.H, applyBox.Y)
	}

	// Inspector edit rejects invalid JSON.
	r.a.page.Form().SetFormValue("insp-edit", "{broken")
	r.tap("insp-apply")
	if tab.cs != nil {
		t.Fatal("invalid JSON via inspector must not be staged")
	}
	if !strings.Contains(strings.ToLower(r.a.toast), "invalid json") {
		t.Fatalf("expected JSON error toast, got %q", r.a.toast)
	}

	// Inspector edit accepts valid JSON, updates the inspector preview immediately, and notifies.
	r.a.page.Form().SetFormValue("insp-edit", `{"ok":true}`)
	r.tap("insp-apply")
	if tab.cs == nil || tab.cs.Pending() != 1 {
		t.Fatal("valid JSON via inspector should stage 1 change")
	}
	if !strings.Contains(r.texts(), `"ok"`) || !strings.Contains(r.a.toast, "Applied to cell") {
		t.Fatalf("Apply to cell should update preview and notify, toast=%q", r.a.toast)
	}
}

func TestDarkModeInputSelectionHighlight(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.openTableByName("users")
	countryCol := r.colIndex("country")
	r.openInlineEditor(0, countryCol)
	r.must(r.a.page.Edit().SelectAll(r.ctx))

	ed := r.box("cell-editor")
	// Sample inside the selected text background: must be dark-theme --qr-sel-bg
	// (#2f5a47, dark green) and never the bright pale-blue default (#d6e2ff).
	foundDarkSel := false
	for x := int(ed.X) + 4; x < int(ed.X+ed.W)-4; x += 2 {
		c := pixelAt(r, x, int(ed.Y+ed.H/2))
		if int(c.B) > int(c.G)+15 && int(c.R) > int(c.G)-20 {
			t.Fatalf("found pale-blue selection highlight %v in dark mode at x=%d", c, x)
		}
		if int(c.G) > int(c.R)+15 && c.G < 140 {
			foundDarkSel = true
		}
	}
	if !foundDarkSel {
		t.Fatal("expected dark-theme green selection highlight inside cell-editor")
	}
}

func TestParseCellTypes(t *testing.T) {
	// Float / Numeric
	floatCol := driver.ColumnMeta{Name: "amount", Type: "NUMERIC(10,2)"}
	if v, err := parseCell(" 3.14 ", floatCol, driver.Null()); err != nil || v.Kind != driver.KindFloat || v.F != 3.14 {
		t.Fatalf("float valid: v=%+v err=%v", v, err)
	}
	for _, bad := range []string{"NaN", "Inf", "-Inf", "12.3abc", ""} {
		if _, err := parseCell(bad, floatCol, driver.Null()); err == nil {
			t.Fatalf("float should reject %q", bad)
		}
	}

	// Date / Time / Timestamp
	dateCol := driver.ColumnMeta{Name: "d", Type: "DATE"}
	if _, err := parseCell("2025-05-01", dateCol, driver.Null()); err != nil {
		t.Fatalf("date valid: %v", err)
	}
	if _, err := parseCell("2025-13-45", dateCol, driver.Null()); err == nil {
		t.Fatal("date should reject 2025-13-45")
	}

	timeCol := driver.ColumnMeta{Name: "t", Type: "TIME"}
	if _, err := parseCell("14:30:00", timeCol, driver.Null()); err != nil {
		t.Fatalf("time valid: %v", err)
	}
	if _, err := parseCell("25:99:00", timeCol, driver.Null()); err == nil {
		t.Fatal("time should reject 25:99:00")
	}

	tsCol := driver.ColumnMeta{Name: "ts", Type: "TIMESTAMP"}
	if _, err := parseCell("2025-05-01 12:30:00", tsCol, driver.Null()); err != nil {
		t.Fatalf("timestamp valid: %v", err)
	}
	if _, err := parseCell("2025-05-01T12:30:00Z", tsCol, driver.Null()); err != nil {
		t.Fatalf("rfc3339 timestamp valid: %v", err)
	}
	if _, err := parseCell("yesterday", tsCol, driver.Null()); err == nil {
		t.Fatal("timestamp should reject 'yesterday'")
	}
}
