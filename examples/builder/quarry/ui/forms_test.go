package ui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
)

func TestValidateConnSQLite(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "x.db")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		d    connDraft
		want map[string]string
	}{
		{"empty", connDraft{Engine: driver.EngineSQLite}, map[string]string{"cd-name": "Name is required.", "cd-path": "Choose a database file."}},
		{"missing", connDraft{Engine: driver.EngineSQLite, Name: "a", Path: filepath.Join(dir, "no.db")}, map[string]string{"cd-path": "File not found"}},
		{"folder", connDraft{Engine: driver.EngineSQLite, Name: "a", Path: dir}, map[string]string{"cd-path": "folder"}},
		{"ok", connDraft{Engine: driver.EngineSQLite, Name: "a", Path: file}, nil},
	}
	for _, c := range cases {
		got := validateConn(c.d)
		if len(got) != len(c.want) {
			t.Errorf("%s: errs %v, want keys %v", c.name, got, c.want)
		}
		for k, frag := range c.want {
			if !strings.Contains(got[k], frag) {
				t.Errorf("%s: %s = %q, want %q", c.name, k, got[k], frag)
			}
		}
	}
}

func TestValidateConnPostgresPort(t *testing.T) {
	base := connDraft{Engine: driver.EnginePostgres, Name: "pg", Host: "h", Database: "d", Port: "5432"}
	if errs := validateConn(base); len(errs) != 0 {
		t.Fatalf("valid draft rejected: %v", errs)
	}
	for port, frag := range map[string]string{"": "required", "abc": "number", "0": "between 1 and 65535", "65536": "between 1 and 65535", "-4": "between"} {
		d := base
		d.Port = port
		if got := validateConn(d)["cd-port"]; !strings.Contains(got, frag) {
			t.Errorf("port %q: %q, want %q", port, got, frag)
		}
	}
	d := base
	d.Host, d.Database = " ", ""
	errs := validateConn(d)
	if errs["cd-host"] == "" || errs["cd-db"] == "" {
		t.Errorf("host and database required: %v", errs)
	}
	d.Mock = true
	if errs := validateConn(d); len(errs) != 0 {
		t.Errorf("mock needs no host: %v", errs)
	}
}

func TestParseDSN(t *testing.T) {
	cfg, err := parseDSN("postgres://bob:pw@db.example:6543/shop?sslmode=require")
	if err != nil || cfg.Host != "db.example" || cfg.Port != 6543 || cfg.Database != "shop" || cfg.User != "bob" || cfg.Password != "pw" || cfg.SSLMode != "require" {
		t.Fatalf("cfg = %+v, err %v", cfg, err)
	}
	for _, bad := range []string{"", "mysql://u@h/d", "postgres:///db", "not a url", "postgres://h:99999/db", "postgres://h:x/db"} {
		if _, err := parseDSN(bad); err == nil {
			t.Errorf("parseDSN(%q) accepted", bad)
		}
	}
}

func (r *rig) setField(id, v string) {
	r.t.Helper()
	r.must(r.a.page.Form().Focus(r.ctx, id)) // the draft syncs the focused field, as typing does
	r.a.page.Form().SetFormValue(id, v)
	r.must(r.a.onChange(r.ctx, r.box(id)))
}

func TestConnDialogInlineErrors(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.tap("rail-connections")
	if r.has("cd-path-err") {
		t.Fatal("no error should show before the user touched or submitted the form")
	}
	// Editing a field validates it at once.
	r.setField("cd-path", filepath.Join(t.TempDir(), "gone.db"))
	if !r.has("cd-path-err") || !strings.Contains(r.texts(), "File not found") {
		t.Fatal("missing file should be reported under the field on change")
	}
	// Submitting shows every problem, and nothing connects.
	r.setField("cd-name", "")
	r.tap("conn-connect")
	if r.a.modal != "conn" || !r.has("cd-name-err") || !r.has("conn-badge") {
		t.Fatal("invalid draft must stay open with errors and a Failed badge")
	}
	// Errors sit below their control.
	if in, er := r.box("cd-path"), r.box("cd-path-err"); er.Y < in.Y+in.H {
		t.Errorf("error (y %.0f) is not under its field (bottom %.0f)", er.Y, in.Y+in.H)
	}
	// PostgreSQL: bad port and missing host.
	r.tap("cd-engine-postgresql")
	r.setField("cd-port", "70000")
	r.setField("cd-host", "x")
	r.setField("cd-host", "")
	for _, id := range []string{"cd-port-err", "cd-host-err", "cd-db-err"} {
		if id == "cd-db-err" {
			r.tap("conn-save")
		}
		if !r.has(id) {
			t.Errorf("%s missing", id)
		}
	}
	if !strings.Contains(r.texts(), "between 1 and 65535") {
		t.Error("port range message missing")
	}
	r.png("conn-errors")
}

func TestConnDialogDSNFill(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.tap("rail-connections")
	r.tap("cd-engine-postgresql")
	r.setField("cd-dsn", "mysql://nope")
	r.tap("cd-dsn-fill")
	if !r.has("cd-dsn-err") {
		t.Fatal("bad URL should be reported under the URL field")
	}
	r.setField("cd-dsn", "postgres://ann:secret@pg.local:6000/app?sslmode=disable")
	r.tap("cd-dsn-fill")
	d := r.a.connDraft
	if d.Host != "pg.local" || d.Port != "6000" || d.Database != "app" || d.User != "ann" || d.Password != "secret" || d.SSL != "disable" || d.DSN != "" {
		t.Fatalf("draft after Fill = %+v", d)
	}
	if r.has("cd-dsn-err") {
		t.Fatal("error should clear after a valid fill")
	}
}

func TestConnDialogPasswordToggle(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.tap("rail-connections")
	r.tap("cd-engine-postgresql")
	r.a.page.Form().SetFormValue("cd-pass", "hunter2")
	passNode := func() (password bool) {
		tree, _ := r.a.Page().Access().AccessibilityTree()
		for _, n := range tree.Nodes {
			if n.ElementID == "cd-pass" {
				return n.Password
			}
		}
		t.Fatal("password node missing")
		return false
	}
	if !passNode() {
		t.Fatal("password starts hidden")
	}
	r.tap("cd-pass-toggle")
	if passNode() || r.a.connDraft.Password != "hunter2" || !strings.Contains(r.texts(), "Hide") {
		t.Fatal("toggle should reveal the password and keep its value")
	}
	r.tap("cd-pass-toggle")
	if !passNode() || !strings.Contains(r.texts(), "Show") {
		t.Fatal("second toggle hides the password again")
	}
}

func TestConnTestRunsAsyncWithBadge(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.tap("rail-connections")
	r.tap("cd-engine-postgresql")
	r.tap("cd-mock-on")
	r.setField("cd-name", "Mock")
	r.tap("conn-test")
	deadline := time.Now().Add(5 * time.Second)
	for r.a.connDraft.Testing {
		if time.Now().After(deadline) {
			t.Fatal("test connection never finished")
		}
		time.Sleep(5 * time.Millisecond)
		r.tick()
	}
	if !r.a.connDraft.ResultOK || !r.has("conn-badge") || !strings.Contains(r.texts(), "Connected in") {
		t.Fatalf("result = %q ok=%v", r.a.connDraft.Result, r.a.connDraft.ResultOK)
	}
	if r.has("conn-spinner") {
		t.Fatal("spinner should be gone after the result")
	}
	// Invalid drafts never start a test.
	r.tap("cd-mock-off")
	r.tap("conn-test")
	if r.a.connDraft.Testing || r.a.connDraft.ResultOK || !r.has("cd-host-err") {
		t.Fatal("invalid draft must show field errors instead of testing")
	}
}

func TestConnTestSpinnerAndDisabledButton(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.tap("rail-connections")
	r.a.connDraft.Testing = true
	r.must(r.a.redraw(r.ctx))
	if !r.has("conn-spinner") || !strings.Contains(r.texts(), "Testing connection") {
		t.Fatal("spinner and label expected while testing")
	}
	before := r.a.connTest.frame
	for i := 0; i < spinnerFrames*2 && r.a.connTest.frame == before; i++ {
		r.tick()
	}
	if r.a.connTest.frame == before {
		t.Fatal("spinner frame should advance on the tick")
	}
	r.tap("conn-test") // disabled: a click does nothing
	if r.a.connTest.seq != 0 {
		t.Fatal("a disabled Test Connection button must not start another test")
	}
}

func TestSettingsSectionsAndSwitch(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.tap("rail-settings")
	for _, id := range []string{"set-sec-appearance", "set-sec-query", "set-sec-monitor", "set-sec-safety", "set-confirm"} {
		if !r.has(id) {
			t.Errorf("settings missing %s", id)
		}
	}
	txt := r.texts()
	for _, want := range []string{"Appearance", "Queries", "Monitoring", "Safety", "How often the driver is sampled."} {
		if !strings.Contains(txt, want) {
			t.Errorf("settings text %q missing", want)
		}
	}
	on := r.a.confirmDestructive
	r.tap("set-confirm")
	if r.a.confirmDestructive == on {
		t.Fatal("switch must flip the setting")
	}
	r.tap("set-confirm")
	if r.a.confirmDestructive != on {
		t.Fatal("switch must flip back")
	}
	// Sections stack top to bottom.
	last := -1.0
	for _, id := range []string{"set-sec-appearance", "set-sec-query", "set-sec-monitor", "set-sec-safety"} {
		if y := r.box(id).Y; y <= last {
			t.Errorf("%s out of order", id)
		} else {
			last = y
		}
	}
	r.png("settings")
}

func TestDialogFootersPutPrimaryLast(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	right := func(id string) float64 { b := r.box(id); return b.X + b.W }
	r.tap("rail-connections")
	if !(right("conn-connect") > right("conn-close") && right("conn-close") > right("conn-save")) {
		t.Error("connection manager: Connect must be rightmost")
	}
	r.tap("conn-close")
	r.must(r.a.askConfirm(r.ctx, &confirmState{Title: "Drop it?", Body: "Gone forever.", Label: "Drop", OK: func(c context.Context) error { return nil }}))
	if right("confirm-ok") <= right("confirm-cancel") {
		t.Error("confirm: the action must sit right of Cancel")
	}
	r.tap("confirm-cancel")
	r.tap("rail-settings")
	if !r.has("set-close") || r.a.modalTitle() != "Settings" {
		t.Error("settings footer")
	}
}
