package telegram_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/telegram/telegram"
)

func dumpBoxes(app *telegram.App) string {
	var b strings.Builder

	for _, box := range app.Boxes() {
		fmt.Fprintf(
			&b,
			"\n%s id=%q action=%q text=%q xywh=%.1f,%.1f,%.1f,%.1f",
			box.Tag,
			box.ID,
			box.Action,
			box.Text,
			box.X,
			box.Y,
			box.W,
			box.H,
		)
	}

	if b.Len() == 0 {
		return "\n(no boxes)"
	}

	return b.String()
}

func boxByID(t *testing.T, app *telegram.App, id string) ownframe.Box {
	t.Helper()

	for _, box := range app.Boxes() {
		if box.ID == id {
			return box
		}
	}

	t.Fatalf("no box id=%q:%s", id, dumpBoxes(app))

	return ownframe.Box{}
}
