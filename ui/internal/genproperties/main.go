// Command genproperties refreshes the builder registry from the OwnFrame
// compatibility-matrix snapshot. Run from the ui package with go generate.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"os"
	"sort"
	"strings"
)

const matrix = "../documentation/compatibility-matrix.md"
const output = "properties_generated.go"
const expectedImplementedPropertyCount = 555
const propertiesPerCaseLine = 5

func main() {
	if err := generate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generate() error {
	f, err := os.Open(matrix)
	if err != nil {
		return err
	}
	defer f.Close()

	var props []string
	seen := make(map[string]bool)
	inSection := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "### 2.1 Implemented") {
			inSection = true
			continue
		}
		if inSection && strings.HasPrefix(line, "### ") {
			break
		}
		if !inSection || !strings.HasPrefix(line, "| \x60") {
			continue
		}
		row := strings.TrimPrefix(line, "| \x60")
		end := strings.Index(row, "\x60 |")
		if end < 0 {
			continue
		}
		name := row[:end]
		if seen[name] {
			return fmt.Errorf("duplicate property %q", name)
		}
		seen[name] = true
		props = append(props, name)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if len(props) != expectedImplementedPropertyCount {
		return fmt.Errorf("expected 555 implemented properties; found %d", len(props))
	}
	sort.Strings(props)

	var buf bytes.Buffer
	fmt.Fprintln(&buf, "// Code generated from documentation/compatibility-matrix.md (2026-10-20 snapshot).")
	fmt.Fprintln(&buf, "// DO NOT EDIT: run go generate ./ui after updating the matrix.")
	fmt.Fprintln(&buf, "package ui")
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "import \"strings\"")
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "// SupportsProperty reports whether a CSS property is marked Implemented in")
	fmt.Fprintln(&buf, "// the pinned Blinkless compatibility snapshot. This is name coverage only;")
	fmt.Fprintln(&buf, "// individual value and layout semantics are enforced by Blinkless.")
	fmt.Fprintln(&buf, "func SupportsProperty(property string) bool {")
	fmt.Fprintln(&buf, "\tswitch strings.ToLower(strings.TrimSpace(property)) {")
	fmt.Fprint(&buf, "\tcase ")
	for i, name := range props {
		if i != 0 {
			if i%propertiesPerCaseLine == 0 {
				fmt.Fprint(&buf, ",\n\t\t")
			} else {
				fmt.Fprint(&buf, ", ")
			}
		}
		fmt.Fprintf(&buf, "%q", name)
	}
	fmt.Fprintln(&buf, ":")
	fmt.Fprintln(&buf, "\t\treturn true")
	fmt.Fprintln(&buf, "\tdefault:")
	fmt.Fprintln(&buf, "\t\treturn false")
	fmt.Fprintln(&buf, "\t}")
	fmt.Fprintln(&buf, "}")
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "// SupportedProperties lists all 555 implemented CSS property names, sorted.")
	fmt.Fprintln(&buf, "func SupportedProperties() []string {")
	fmt.Fprintln(&buf, "\treturn []string{")
	for _, name := range props {
		fmt.Fprintf(&buf, "\t\t%q,\n", name)
	}
	fmt.Fprintln(&buf, "\t}")
	fmt.Fprintln(&buf, "}")
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	return os.WriteFile(output, formatted, 0600)
}
