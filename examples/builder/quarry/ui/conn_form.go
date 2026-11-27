package ui

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/ui"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
)

const (
	maxPort        = 65535
	spinnerFrames  = 3
	spinnerStepMs  = 180
	dsnPlaceholder = "postgres://user@host:5432/database"
)

// fieldErrs maps a form control id to the message shown under it.
type fieldErrs map[string]string

// validateConn checks a connection draft the way the driver would, so the form
// can name the broken field instead of failing at connect time.
func validateConn(d connDraft) fieldErrs {
	errs := fieldErrs{}
	if strings.TrimSpace(d.Name) == "" {
		errs["cd-name"] = "Name is required."
	}
	if d.Engine == driver.EngineSQLite {
		path := strings.TrimSpace(d.Path)
		switch fi, err := os.Stat(path); {
		case path == "":
			errs["cd-path"] = "Choose a database file."
		case err != nil:
			errs["cd-path"] = "File not found: " + path
		case fi.IsDir():
			errs["cd-path"] = "That path is a folder, not a database file."
		}
		return errs
	}
	if d.Mock {
		return errs
	}
	if strings.TrimSpace(d.Host) == "" {
		errs["cd-host"] = "Host is required."
	}
	switch n, err := strconv.Atoi(strings.TrimSpace(d.Port)); {
	case strings.TrimSpace(d.Port) == "":
		errs["cd-port"] = "Port is required."
	case err != nil:
		errs["cd-port"] = "Port must be a number."
	case n < 1 || n > maxPort:
		errs["cd-port"] = "Port must be between 1 and " + strconv.Itoa(maxPort) + "."
	}
	if strings.TrimSpace(d.Database) == "" {
		errs["cd-db"] = "Database name is required."
	}
	return errs
}

// parseDSN reads a postgres:// or postgresql:// URL into a config. Unlike the
// command-line parser it rejects URLs without a host and out-of-range ports.
func parseDSN(raw string) (driver.ConnectionConfig, error) {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		return driver.ConnectionConfig{}, fmt.Errorf("not a PostgreSQL URL. Expected %s", dsnPlaceholder)
	}
	if u.Hostname() == "" {
		return driver.ConnectionConfig{}, fmt.Errorf("the URL has no host. Expected %s", dsnPlaceholder)
	}
	if p := u.Port(); p != "" {
		if n, perr := strconv.Atoi(p); perr != nil || n < 1 || n > maxPort {
			return driver.ConnectionConfig{}, fmt.Errorf("the URL port must be between 1 and %d", maxPort)
		}
	}
	return parsePostgresURL(raw)
}

// shown reports the message for id once the field was edited or the form was
// submitted.
func (d connDraft) shown(errs fieldErrs, id string) string {
	if d.Submitted || d.Touched[id] {
		return errs[id]
	}
	return ""
}

func (d *connDraft) touch(id string) {
	if d.Touched == nil {
		d.Touched = map[string]bool{}
	}
	d.Touched[id] = true
}

// connFieldChanged records an edit in a connection field and repaints only
// when the visible validation changed.
func (a *App) connFieldChanged(ctx context.Context, id string) error {
	d := &a.connDraft
	before := d.visibleErrs()
	a.syncConnDraft()
	d.touch(id)
	if d.visibleErrs() == before {
		return nil
	}
	return a.redraw(ctx)
}

// visibleErrs is a comparable fingerprint of the shown messages.
func (d connDraft) visibleErrs() string {
	errs := validateConn(d)
	var sb strings.Builder
	for _, id := range []string{"cd-name", "cd-path", "cd-host", "cd-port", "cd-db", "cd-dsn"} {
		if m := d.shown(errs, id); m != "" {
			sb.WriteString(id + "=" + m + ";")
		}
	}
	return sb.String()
}

// ---- field layout ----

// field is one labelled form row: the control, then either its error or its
// hint on the line below.
func field(id, label string, control *ui.Element, errText, hint string) *ui.Element {
	col := ui.Div().Class("qr-field-col").Children(control)
	switch {
	case errText != "":
		col.Children(ui.Div().ID(id+"-err").Class("qr-field-err").Attr("role", ariaRole("status")).Text(errText))
	case hint != "":
		col.Children(ui.Div().ID(id + "-hint").Class("qr-field-hint").Text(hint))
	}
	return ui.Div().ID(id+"-row").Class("qr-form-row qr-form-row-top").Children(
		ui.Span().Class("qr-form-label").Text(label), col)
}

// marked flags a control whose message is showing.
func marked(el *ui.Element, errText string) *ui.Element {
	if errText == "" {
		return el
	}
	return el.Class("qr-input qr-input-bad").Attr("aria-invalid", "true")
}

// ---- password reveal, DSN fill ----

func (a *App) togglePassword(ctx context.Context) error {
	a.syncConnDraft()
	a.connDraft.ShowPass = !a.connDraft.ShowPass
	return a.redraw(ctx)
}

// fillFromDSN copies a pasted PostgreSQL URL into the form fields. The URL
// field is cleared afterwards so the password does not stay in plain text.
func (a *App) fillFromDSN(ctx context.Context) error {
	a.syncConnDraft()
	d := &a.connDraft
	d.touch("cd-dsn")
	cfg, err := parseDSN(d.DSN)
	if err != nil {
		d.DSNErr = err.Error()
		return a.redraw(ctx)
	}
	d.DSNErr, d.DSN = "", ""
	d.Engine, d.Host, d.Database, d.User, d.Password = driver.EnginePostgres, cfg.Host, cfg.Database, cfg.User, cfg.Password
	d.Port = strconv.Itoa(cfg.Port)
	if cfg.Port == 0 {
		d.Port = "5432"
	}
	if cfg.SSLMode != "" {
		d.SSL = cfg.SSLMode
	}
	if strings.TrimSpace(d.Name) == "" || d.Name == "New connection" {
		d.Name = cfg.Name
	}
	d.Result, d.ResultOK = "", false
	return a.redraw(ctx)
}

// ---- async test connection ----

type connTestDone struct {
	seq int
	ok  bool
	msg string
}

// connTestState runs one Test Connection at a time off the UI loop.
type connTestState struct {
	ch    chan connTestDone
	seq   int
	frame int
}

// testConn validates the draft, then opens and pings the driver in the
// background. The result lands through drainConnTest on the frame tick.
func (a *App) testConn(ctx context.Context) error {
	a.syncConnDraft()
	d := &a.connDraft
	d.Submitted = true
	if len(validateConn(*d)) > 0 {
		d.Result, d.ResultOK = "Fix the highlighted fields, then test again.", false
		return a.redraw(ctx)
	}
	if d.Testing {
		return nil
	}
	if a.connTest.ch == nil {
		a.connTest.ch = make(chan connTestDone, connTestBuf)
	}
	a.connTest.seq++
	seq, cfg, ch := a.connTest.seq, d.config(), a.connTest.ch
	d.Testing, d.Result = true, ""
	parent := context.WithoutCancel(ctx) // the test outlives this handler; connectTestTime bounds it
	go func() {
		tctx, cancel := context.WithTimeout(parent, connectTestTime)
		defer cancel()
		start := time.Now()
		drv, err := driver.Open(tctx, cfg)
		if err == nil {
			err = drv.Ping(tctx)
			if err == nil {
				v, _ := drv.Version(tctx)
				ch <- connTestDone{seq: seq, ok: true, msg: fmt.Sprintf("Connected in %s. %s", fmtDur(time.Since(start)), v)}
			}
			_ = drv.Close()
		}
		if err != nil {
			ch <- connTestDone{seq: seq, msg: "Connection failed: " + err.Error()}
		}
	}()
	return a.redraw(ctx)
}

// drainConnTest applies a finished test and animates the spinner. It reports
// whether the dialog needs a repaint.
func (a *App) drainConnTest() bool {
	dirty := false
	for a.connTest.ch != nil {
		select {
		case r := <-a.connTest.ch:
			if r.seq == a.connTest.seq {
				a.connDraft.Testing = false
				a.connDraft.Result, a.connDraft.ResultOK = r.msg, r.ok
				dirty = true
			}
			continue
		default:
		}
		break
	}
	if a.connDraft.Testing && a.modal == "conn" {
		if f := int(a.clock().UnixMilli()/spinnerStepMs) % spinnerFrames; f != a.connTest.frame {
			a.connTest.frame = f
			dirty = true
		}
	}
	return dirty && a.modal == "conn"
}

// spinner is three dots with one lit; the frame advances on the tick.
func (a *App) spinner(id string) *ui.Element {
	s := ui.Span().ID(id).Class("qr-spinner").Attr("role", "img").Attr("aria-label", "Working")
	for i := 0; i < spinnerFrames; i++ {
		cls := "qr-spin-dot"
		if i == a.connTest.frame {
			cls += " qr-spin-on"
		}
		s.Children(ui.Span().Class(cls))
	}
	return s
}

// connStatus is the Test Connection readout: spinner while running, then a
// badge and the message.
func (a *App) connStatus() *ui.Element {
	d := a.connDraft
	switch {
	case d.Testing:
		return ui.Div().ID("conn-result").Class("qr-conn-status").Attr("role", ariaRole("status")).Children(
			a.spinner("conn-spinner"), ui.Span().Class("qr-muted").Text("Testing connection..."))
	case d.Result != "":
		badge, cls, note := "Failed", "qr-badge qr-badge-err", "qr-note qr-note-err"
		if d.ResultOK {
			badge, cls, note = "OK", "qr-badge qr-badge-ok", "qr-note qr-note-ok"
		}
		return ui.Div().ID("conn-result").Class("qr-conn-status").Attr("role", ariaRole("status")).Children(
			ui.Span().ID("conn-badge").Class(cls).Text(badge), ui.Span().Class(note).Text(d.Result))
	}
	return nil
}
