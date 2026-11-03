// Command migrate-ui-builder migrates the UI builder examples from the old
// string-property API to generated property methods. Run from the scripts
// module with go run ./cmd/migrate-ui-builder.
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
	"slices"
	"strconv"
	"strings"
)

const (
	propertyValueArguments = 2
	migratedFileMode       = 0o600
)

func main() {
	root := flag.String("root", "..", "repository root")
	flag.Parse()
	if err := migrate(*root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func migrate(root string) error {
	for _, example := range []string{"examples/builder/ui-builder", "examples/builder/jmap-mail"} {
		base := filepath.Join(root, example)
		entries, err := os.ReadDir(base)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
				continue
			}
			path := filepath.Join(base, entry.Name())
			if err := migrateFile(path); err != nil {
				return err
			}
		}
	}
	return nil
}

func migrateFile(path string) error {
	source, err := os.ReadFile(path) //nolint:gosec // reason: paths come only from the two fixed example roots.
	if err != nil {
		return err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, source, parser.ParseComments)
	if err != nil {
		return err
	}
	uiName, importErr := uiImportName(file)
	keywords := keywordValues()
	changed := false
	var migrateErr error
	ast.Inspect(file, func(node ast.Node) bool {
		if migrateErr != nil {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		oldMethod := selector.Sel.Name
		property := oldMethod
		valueIndex := 0
		switch oldMethod {
		case "CSS", "Style":
			if len(call.Args) != propertyValueArguments {
				migrateErr = fmt.Errorf("%s:%s: expected property and value arguments", path, fset.Position(call.Pos()))
				return false
			}
			propertyLiteral, ok := call.Args[0].(*ast.BasicLit)
			if !ok || propertyLiteral.Kind != token.STRING {
				migrateErr = fmt.Errorf(
					"%s:%s: runtime property requires manual migration to CSSRaw",
					path,
					fset.Position(call.Pos()),
				)
				return false
			}
			property, err = strconv.Unquote(propertyLiteral.Value)
			if err != nil {
				migrateErr = err
				return false
			}
			property = goName(property)
			valueIndex = 1
		case "Styles":
			migrateErr = fmt.Errorf("%s:%s: Styles map requires manual migration", path, fset.Position(call.Pos()))
			return false
		case "Display", "Padding", "BorderRadius", "FlexDirection", "AlignItems", "Gap",
			"BackgroundColor", "Color", "FontSize", "FontWeight", "Margin", "Width", "Height":
		default:
			return true
		}
		if len(call.Args) != valueIndex+1 {
			migrateErr = fmt.Errorf("%s:%s: expected one style value", path, fset.Position(call.Pos()))
			return false
		}
		if oldMethod != "CSS" && oldMethod != "Style" {
			if value, ok := unwrapKeyword(call.Args[valueIndex], uiName); ok {
				call.Args[valueIndex] = value
				changed = true
				return true
			}
			if isUIValue(call.Args[valueIndex], uiName) {
				return true
			}
		}
		if importErr != nil {
			migrateErr = fmt.Errorf("%s: %w", path, importErr)
			return false
		}
		selector.Sel.Name = property
		call.Args = []ast.Expr{typedValue(property, call.Args[valueIndex], uiName, keywords)}
		changed = true
		return true
	})
	if migrateErr != nil {
		return migrateErr
	}
	if !changed {
		return nil
	}
	var output bytes.Buffer
	if err := format.Node(&output, fset, file); err != nil {
		return fmt.Errorf("format migrated %s: %w", path, err)
	}
	if bytes.Equal(source, output.Bytes()) {
		return nil
	}
	return os.WriteFile(path, output.Bytes(), migratedFileMode)
}

func unwrapKeyword(expr ast.Expr, uiName string) (ast.Expr, bool) {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return nil, false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "RawValue" {
		return nil, false
	}
	packageName, ok := selector.X.(*ast.Ident)
	if !ok || packageName.Name != uiName {
		return nil, false
	}
	value, ok := call.Args[0].(*ast.SelectorExpr)
	if !ok {
		return nil, false
	}
	packageName, ok = value.X.(*ast.Ident)
	if !ok || packageName.Name != uiName {
		return nil, false
	}
	return value, true
}

func isUIValue(expr ast.Expr, uiName string) bool {
	var selector *ast.SelectorExpr
	switch value := expr.(type) {
	case *ast.CallExpr:
		var ok bool
		selector, ok = value.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
	case *ast.SelectorExpr:
		selector = value
	default:
		return false
	}
	packageName, ok := selector.X.(*ast.Ident)
	return ok && packageName.Name == uiName
}

func uiImportName(file *ast.File) (string, error) {
	for _, item := range file.Imports {
		path, err := strconv.Unquote(item.Path.Value)
		if err != nil {
			return "", err
		}
		if path != "github.com/chinmay-sawant/ownframe/ui" {
			continue
		}
		if item.Name != nil {
			return item.Name.Name, nil
		}
		return "ui", nil
	}
	return "", fmt.Errorf("missing ownframe/ui import")
}

func typedValue(property string, value ast.Expr, ui string, keywords map[string]string) ast.Expr {
	if literal, ok := value.(*ast.BasicLit); ok && literal.Kind == token.STRING {
		text, err := strconv.Unquote(literal.Value)
		if err == nil {
			if keyword := keywordValue(property, text, keywords); keyword != "" {
				return packageSelector(ui, keyword)
			}
			if property == "Color" || property == "BackgroundColor" {
				if isHex(text) {
					return call(ui, "Hex", literal)
				}
			}
			if value := scalarValue(property, text, ui); value != nil {
				return value
			}
			if value := spacingValue(property, text, ui); value != nil {
				return value
			}
		}
	}
	return call(ui, "RawValue", value)
}

func scalarValue(property, text, ui string) ast.Expr {
	units := map[string]string{
		"px": "Px", "rem": "Rem", "%": "Percent",
	}
	for unit, constructor := range units {
		if !strings.HasSuffix(text, unit) {
			continue
		}
		number := strings.TrimSuffix(text, unit)
		if _, err := strconv.ParseFloat(number, 64); err != nil {
			return nil
		}
		if unit == "%" && !supportsPercentage(property) || unit != "%" && !supportsLength(property) {
			return nil
		}
		return call(ui, constructor, &ast.BasicLit{Kind: token.FLOAT, Value: number})
	}
	if number, err := strconv.ParseFloat(text, 64); err == nil {
		if supportsNumber(property) {
			return call(ui, "Number", &ast.BasicLit{Kind: token.FLOAT, Value: text})
		}
		if number == 0 && supportsLength(property) {
			return call(ui, "Px", &ast.BasicLit{Kind: token.INT, Value: "0"})
		}
	}
	return nil
}

func spacingValue(property, text, ui string) ast.Expr {
	if !supportsSpacing(property) {
		return nil
	}
	parts := strings.Fields(text)
	if len(parts) < 2 || len(parts) > 4 {
		return nil
	}
	values := make([]ast.Expr, 0, len(parts))
	for _, part := range parts {
		value := scalarValue(property, part, ui)
		if value == nil {
			return nil
		}
		call, ok := value.(*ast.CallExpr)
		if !ok {
			return nil
		}
		method, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || method.Sel.Name != "Px" {
			return nil
		}
		values = append(values, value)
	}
	return call(ui, "Edges", values...)
}

func keywordValue(property, text string, choices map[string]string) string {
	for _, value := range strings.Fields(choices[property]) {
		if value == text {
			return property + goName(value)
		}
	}
	return ""
}

func keywordValues() map[string]string {
	return map[string]string{
		"Display": strings.Join([]string{
			"none block inline inline-block flex inline-flex grid inline-grid",
			"table table-row table-cell contents",
		}, " "),
		"FlexDirection": "row row-reverse column column-reverse",
		"JustifyContent": strings.Join([]string{
			"normal start end flex-start flex-end center space-between",
			"space-around space-evenly stretch",
		}, " "),
		"AlignItems":   "normal start end flex-start flex-end center baseline stretch",
		"Position":     "static relative absolute fixed sticky",
		"Overflow":     "visible hidden clip scroll auto",
		"OverflowX":    "visible hidden clip scroll auto",
		"OverflowY":    "visible hidden clip scroll auto",
		"BoxSizing":    "border-box content-box",
		"TextAlign":    "start end left right center justify",
		"WhiteSpace":   "normal nowrap pre pre-wrap pre-line break-spaces",
		"OverflowWrap": "normal break-word anywhere",
		"FontWeight":   "normal bold bolder lighter",
		"FontStyle":    "normal italic oblique",
		"Visibility":   "visible hidden collapse",
		"TextOverflow": "clip ellipsis",
		"Width":        "auto min-content max-content fit-content",
		"Height":       "auto min-content max-content fit-content",
	}
}

const lengthProperties = `
width height min-width min-height max-width max-height inline-size block-size
min-inline-size min-block-size max-inline-size max-block-size top right bottom left inset
inset-block inset-inline inset-block-start inset-block-end inset-inline-start
inset-inline-end padding padding-top padding-right padding-bottom padding-left
padding-block padding-inline padding-block-start padding-block-end padding-inline-start
padding-inline-end margin margin-top margin-right margin-bottom margin-left margin-block
margin-inline margin-block-start margin-block-end margin-inline-start margin-inline-end
gap row-gap column-gap font-size line-height letter-spacing word-spacing text-indent
border-radius border-top-left-radius border-top-right-radius border-bottom-left-radius
border-bottom-right-radius border-width border-top-width border-right-width
border-bottom-width border-left-width outline-width outline-offset flex-basis column-width
perspective stroke-width
`

const percentageProperties = `
width height min-width min-height max-width max-height inline-size min-inline-size
max-inline-size block-size min-block-size max-block-size top right bottom left inset
inset-block inset-inline inset-block-start inset-block-end inset-inline-start
inset-inline-end margin margin-top margin-right margin-bottom margin-left margin-block
margin-inline margin-block-start margin-block-end margin-inline-start margin-inline-end
padding padding-top padding-right padding-bottom padding-left padding-block padding-inline
padding-block-start padding-block-end padding-inline-start padding-inline-end gap row-gap
column-gap font-size line-height border-radius border-top-left-radius
border-top-right-radius border-bottom-left-radius border-bottom-right-radius flex-basis
text-indent
`

const numberProperties = `
opacity flex-grow flex-shrink order z-index font-weight line-height column-count orphans
widows fill-opacity stroke-opacity
`

const spacingProperties = `
padding margin gap border-radius padding-block padding-inline margin-block margin-inline
`

func supportsLength(property string) bool     { return hasProperty(lengthProperties, property) }
func supportsPercentage(property string) bool { return hasProperty(percentageProperties, property) }
func supportsNumber(property string) bool     { return hasProperty(numberProperties, property) }
func supportsSpacing(property string) bool    { return hasProperty(spacingProperties, property) }

func hasProperty(properties, property string) bool {
	return slices.Contains(strings.Fields(properties), hyphenName(property))
}
func isHex(text string) bool {
	if len(text) != 4 && len(text) != 5 && len(text) != 7 && len(text) != 9 || text[0] != '#' {
		return false
	}
	for _, c := range text[1:] {
		if !strings.ContainsRune("0123456789abcdefABCDEF", c) {
			return false
		}
	}
	return true
}

func hyphenName(name string) string {
	var out strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			out.WriteByte('-')
		}
		out.WriteRune(r)
	}
	return out.String()
}

func goName(property string) string {
	parts := strings.Split(strings.TrimPrefix(property, "-"), "-")
	for i, part := range parts {
		if part != "" {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func call(pkg, name string, args ...ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{Fun: packageSelector(pkg, name), Args: args}
}

func packageSelector(pkg, name string) ast.Expr {
	return &ast.SelectorExpr{X: ast.NewIdent(pkg), Sel: ast.NewIdent(name)}
}
