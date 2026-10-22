// merge-go joins explicitly selected Go files from one package. It refuses
// platform files and build constraints; reviewers choose the responsibility.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func main() {
	write := flag.Bool("write", false, "write the destination and remove its source files")
	flag.Parse()
	if err := merge(flag.Args(), *write); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func merge(args []string, write bool) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: merge-go [-write] destination.go source.go source.go ...")
	}
	dest := filepath.Clean(args[0])
	imports := map[string]string{}
	var body bytes.Buffer
	pkg, packageDoc := "", ""
	seen := map[string]bool{}
	for _, source := range args[1:] {
		source = filepath.Clean(source)
		if seen[source] || filepath.Dir(source) != filepath.Dir(dest) {
			return fmt.Errorf("duplicate or cross-package source: %s", source)
		}
		seen[source] = true
		if strings.HasSuffix(source, "_test.go") != strings.HasSuffix(dest, "_test.go") {
			return fmt.Errorf("cannot mix test and production files: %s", source)
		}
		data, err := os.ReadFile(source)
		if err != nil {
			return err
		}
		if bytes.Contains(data, []byte("//go:")) || bytes.Contains(data, []byte("// +build")) {
			return fmt.Errorf("preserve directives manually: %s", source)
		}
		name := strings.TrimSuffix(strings.TrimSuffix(source, ".go"), "_test")
		for _, suffix := range []string{"aix", "android", "darwin", "dragonfly", "freebsd", "illumos", "ios", "js", "linux", "netbsd", "openbsd", "plan9", "solaris", "wasip1", "windows", "386", "amd64", "arm", "arm64", "loong64", "mips", "mipsle", "mips64", "mips64le", "ppc64", "ppc64le", "riscv64", "s390x", "wasm"} {
			if strings.HasSuffix(name, "_"+suffix) {
				return fmt.Errorf("preserve platform suffix: %s", source)
			}
		}
		set := token.NewFileSet()
		file, err := parser.ParseFile(set, source, data, parser.ParseComments)
		if err != nil {
			return err
		}
		if pkg != "" && pkg != file.Name.Name {
			return fmt.Errorf("package mismatch in %s", source)
		}
		pkg = file.Name.Name
		if file.Doc != nil {
			if packageDoc != "" {
				return fmt.Errorf("more than one package comment; review manually")
			}
			packageDoc = string(data[set.Position(file.Doc.Pos()).Offset:set.Position(file.Doc.End()).Offset]) + "\n"
		}
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			alias := ""
			if spec.Name != nil {
				alias = spec.Name.Name
			}
			if prior, ok := imports[path]; ok && prior != alias {
				return fmt.Errorf("conflicting aliases for %s", path)
			}
			if path == "C" {
				return fmt.Errorf("preserve cgo preambles manually: %s", source)
			}
			imports[path] = alias
		}
		// Preserve source text and trailing comments; remove only import declarations.
		start := set.Position(file.Name.End()).Offset
		for _, decl := range file.Decls {
			if imp, ok := decl.(*ast.GenDecl); ok && imp.Tok == token.IMPORT {
				body.Write(data[start:set.Position(imp.Pos()).Offset])
				start = set.Position(imp.End()).Offset
			}
		}
		body.Write(data[start:])
		body.WriteString("\n")
	}
	if _, err := os.Stat(dest); err == nil && !seen[dest] {
		return fmt.Errorf("destination already exists and is not a source: %s", dest)
	}
	var joined bytes.Buffer
	fmt.Fprintf(&joined, "%spackage %s\n\n", packageDoc, pkg)
	paths := make([]string, 0, len(imports))
	for path := range imports {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	if len(paths) != 0 {
		joined.WriteString("import (\n")
		for _, path := range paths {
			fmt.Fprintf(&joined, "%s %q\n", imports[path], path)
		}
		joined.WriteString(")\n")
	}
	joined.Write(body.Bytes())
	formatted, err := format.Source(joined.Bytes())
	if err != nil {
		return err
	}
	if !write {
		_, err = os.Stdout.Write(formatted)
		return err
	}
	if err := os.WriteFile(dest, formatted, 0o644); err != nil {
		return err
	}
	for _, source := range args[1:] {
		if filepath.Clean(source) != dest {
			if err := os.Remove(source); err != nil {
				return err
			}
		}
	}
	fmt.Printf("merged %d files into %s\n", len(seen), dest)
	return nil
}
