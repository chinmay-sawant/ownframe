// Command gen-ui-properties refreshes the public ui builder's CSS registry
// from the OwnFrame compatibility matrix. Run through go generate ./ui.
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
const output = "../ui/properties_generated.go"
const propertiesPerCaseLine = 5
const generatedFileMode = 0o600

func main() {
	if err := generate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generate() error {
	source, err := os.ReadFile(matrix)
	if err != nil {
		return err
	}

	props := []string{}
	seen := make(map[string]bool)
	inSection := false
	scanner := bufio.NewScanner(bytes.NewReader(source))
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
	if len(props) == 0 {
		return fmt.Errorf("no implemented properties in compatibility matrix")
	}
	sort.Strings(props)

	var buf bytes.Buffer
	fmt.Fprintln(&buf, "// Code generated from documentation/compatibility-matrix.md.")
	fmt.Fprintln(&buf, "// DO NOT EDIT: run go generate ./ui after updating the matrix.")
	fmt.Fprintln(&buf, "package ui")
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, `import ("strings"; "fmt")`)
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
	fmt.Fprintln(&buf, "// SupportedProperties lists the implemented CSS property names, sorted.")
	fmt.Fprintln(&buf, "func SupportedProperties() []string {")
	fmt.Fprintln(&buf, "\treturn []string{")
	for _, name := range props {
		fmt.Fprintf(&buf, "\t\t%q,\n", name)
	}
	fmt.Fprintln(&buf, "\t}")
	fmt.Fprintln(&buf, "}")
	if err := generateStyles(&buf, props); err != nil {
		return err
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	return os.WriteFile(output, formatted, generatedFileMode)
}
