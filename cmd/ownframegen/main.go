// Command ownframegen compiles HTML templates with Go-style actions into
// typed, zero-reflection Go view structs, fluent builders, and typed
// ElementPatch builders for OwnFrame.
//
// Usage:
//
//	ownframegen -in view.html -out view_gen.go -pkg main -type DashboardView
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const outputFilePerm = 0o600

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "ownframegen: %v\n", err)
		os.Exit(1)
	}
}

type cliOptions struct {
	inPath   string
	outPath  string
	pkgName  string
	typeName string
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("ownframegen", flag.ContinueOnError)
	var opts cliOptions
	fs.StringVar(&opts.inPath, "in", "", "input .html file, glob pattern, or directory")
	fs.StringVar(&opts.outPath, "out", "", "output .go file path (defaults to <stem>_gen.go, or '-' for stdout)")
	fs.StringVar(&opts.pkgName, "pkg", "main", "Go package name for the generated file")
	fs.StringVar(&opts.typeName, "type", "", "generated Go view struct name (single-file input)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if strings.TrimSpace(opts.inPath) == "" {
		return errors.New("missing required -in flag")
	}
	if strings.TrimSpace(opts.pkgName) == "" {
		return errors.New("package name -pkg cannot be empty")
	}

	inputs, err := resolveInputs(opts.inPath)
	if err != nil {
		return err
	}
	if len(inputs) == 0 {
		return fmt.Errorf("no .html files matched %q", opts.inPath)
	}
	if len(inputs) > 1 && opts.typeName != "" {
		return fmt.Errorf("-type %q cannot be used with multiple input files (%d matched)", opts.typeName, len(inputs))
	}

	views := make([]*ParsedView, 0, len(inputs))
	for _, path := range inputs {
		//nolint:gosec // CLI tool intentionally reads user-specified template path.
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}

		typeName := opts.typeName
		if typeName == "" {
			typeName = defaultTypeName(path)
		}

		view, parseErr := ParseTemplate(filepath.Base(path), typeName, string(raw))
		if parseErr != nil {
			return fmt.Errorf("parse %s: %w", path, parseErr)
		}
		views = append(views, view)
	}

	src, err := EmitFile(opts.pkgName, views)
	if err != nil {
		return fmt.Errorf("emit Go source: %w", err)
	}

	outPath := opts.outPath
	if outPath == "" {
		outPath = defaultOutPath(inputs[0])
	}
	if outPath == "-" {
		_, err = stdout.Write(src)

		return err
	}

	if err := os.WriteFile(outPath, src, outputFilePerm); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}

	return nil
}

func resolveInputs(pattern string) ([]string, error) {
	if strings.ContainsAny(pattern, "*?[") {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("glob %q: %w", pattern, err)
		}
		sort.Strings(matches)

		return filterHTMLFiles(matches), nil
	}

	info, err := os.Stat(pattern)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{pattern}, nil
	}

	entries, err := os.ReadDir(pattern)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(e.Name()), ".html") {
			out = append(out, filepath.Join(pattern, e.Name()))
		}
	}
	sort.Strings(out)

	return out, nil
}

func filterHTMLFiles(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if strings.EqualFold(filepath.Ext(p), ".html") {
			out = append(out, p)
		}
	}

	return out
}

func defaultTypeName(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	name := toPascalCase(base)
	if name == "" {
		return "View"
	}
	if strings.HasSuffix(name, "View") {
		return name
	}

	return name + "View"
}

func defaultOutPath(firstInput string) string {
	ext := filepath.Ext(firstInput)
	stem := strings.TrimSuffix(firstInput, ext)

	return stem + "_gen.go"
}
