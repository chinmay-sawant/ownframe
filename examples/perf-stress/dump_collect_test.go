package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/perf-benchmarks/benchutil"
)

// dumpApp is one headless page plus its redraw func.
type dumpApp struct {
	name   string
	page   *ownframe.Page
	redraw func(context.Context) error
}

// collectDumps redraws every app headless and returns one Stats snapshot
// per app, so scripts/perf-dump.sh archives a run headless.
func collectDumps(t *testing.T, ctx context.Context, apps []dumpApp) []json.RawMessage {
	t.Helper()

	var docs []json.RawMessage

	for _, a := range apps {
		if err := a.redraw(ctx); err != nil {
			t.Fatal(err)
		}

		b, err := benchutil.Dump(a.name, a.page)
		if err != nil {
			t.Fatal(err)
		}

		docs = append(docs, b)
	}

	return docs
}
