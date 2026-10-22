//go:build !js

// Command live-log-viewer follows local log files, or a seeded dummy
// stream by default, in an ownframe window. Entries land in SQLite under
// the user config directory unless -dir or -temp overrides it.
package main

import (
	"context"
	"flag"
	"log"
	"path/filepath"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/live-log-viewer/ui"
)

// files collects repeated -file flags.
type files []string

func (f *files) String() string { return strings.Join(*f, ",") }

func (f *files) Set(v string) error {
	*f = append(*f, v)

	return nil
}

func main() {
	dir := flag.String("dir", "", "data directory; empty uses the user config directory")
	temp := flag.Bool("temp", false, "use a fresh temporary data directory")
	perf := flag.Bool("perf", false, "show redraw timings in the status bar")
	fromEnd := flag.Bool("from-end", true, "start -file sources at their current end")
	var list files
	flag.Var(&list, "file", "log file to follow; repeatable, default is dummy mode")
	flag.Parse()

	st := openStore(*dir, *temp)
	defer func() {
		if err := st.Close(); err != nil {
			log.Print(err)
		}
	}()

	stop := startIngest(st, list, *fromEnd)
	defer stop()

	app, err := ui.New(ui.Options{
		Feed:      ui.NewStoreFeed(st),
		ExportDir: filepath.Join(st.Dir(), "exports"),
		Perf:      *perf,
	})
	if err != nil {
		log.Print(err)

		return
	}

	defer func() {
		if err := app.Close(); err != nil {
			log.Print(err)
		}
	}()

	if err := ownframe.Run(context.Background(), app.Page()); err != nil {
		log.Print(err)
	}
}
