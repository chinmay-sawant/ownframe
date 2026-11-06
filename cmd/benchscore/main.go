// Command benchscore compares two `go test -bench` outputs under the shared
// benchmark class budgets.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/chinmay-sawant/ownframe/internal/benchscore"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "benchscore: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("benchscore", flag.ContinueOnError)
	refPath := flags.String("ref", "", "reference benchmark output file")
	candPath := flags.String("cand", "", "candidate benchmark output file")
	manifestPath := flags.String("manifest", "", "optional JSON class manifest")
	flags.Usage = func() {
		_, _ = fmt.Fprintln(flags.Output(), "usage: benchscore -ref <reference.txt> -cand <candidate.txt> [-manifest <classes.json>]")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *refPath == "" || *candPath == "" {
		flags.Usage()
		return errors.New("both -ref and -cand are required")
	}

	reference, err := benchscore.ParseFile(*refPath)
	if err != nil {
		return err
	}
	candidate, err := benchscore.ParseFile(*candPath)
	if err != nil {
		return err
	}
	manifest, err := benchscore.LoadManifest(*manifestPath)
	if err != nil {
		return err
	}

	report := benchscore.Compare(reference, candidate, manifest)
	if err := report.Write(stdout); err != nil {
		return fmt.Errorf("write scorecard: %w", err)
	}
	if failed := len(report.FailedRows()); failed > 0 {
		return fmt.Errorf("%d of %d rows failed the budget", failed, len(report.Rows))
	}
	return nil
}
