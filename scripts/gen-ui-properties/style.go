package main

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Explicit value families keep length and number helpers off unrelated properties.
// Properties without a family still accept CSS-wide keywords, Var, and RawValue.
func valueFamilies() map[string][]string {
	return map[string][]string{
		"Length": strings.Fields(`
			width height min-width min-height max-width max-height inline-size block-size min-inline-size
			min-block-size max-inline-size max-block-size top right bottom left inset inset-block
			inset-inline inset-block-start inset-block-end inset-inline-start inset-inline-end padding
			padding-top padding-right padding-bottom padding-left padding-block padding-inline
			padding-block-start padding-block-end padding-inline-start padding-inline-end margin margin-top
			margin-right margin-bottom margin-left margin-block margin-inline margin-block-start
			margin-block-end margin-inline-start margin-inline-end gap row-gap column-gap font-size
			line-height letter-spacing word-spacing text-indent border-radius border-top-left-radius
			border-top-right-radius border-bottom-left-radius border-bottom-right-radius border-width
			border-top-width border-right-width border-bottom-width border-left-width outline-width
			outline-offset flex-basis column-width perspective stroke-width
		`),
		"Percentage": strings.Fields(`
			width height min-width min-height max-width max-height inline-size min-inline-size
			max-inline-size block-size min-block-size max-block-size top right bottom left inset inset-block
			inset-inline inset-block-start inset-block-end inset-inline-start inset-inline-end margin
			margin-top margin-right margin-bottom margin-left margin-block margin-inline margin-block-start
			margin-block-end margin-inline-start margin-inline-end padding padding-top padding-right
			padding-bottom padding-left padding-block padding-inline padding-block-start padding-block-end
			padding-inline-start padding-inline-end gap row-gap column-gap font-size line-height
			border-radius border-top-left-radius border-top-right-radius border-bottom-left-radius
			border-bottom-right-radius flex-basis text-indent
		`),
		"Color": strings.Fields(`
			color background-color border-color border-top-color border-right-color border-bottom-color
			border-left-color outline-color text-decoration-color text-emphasis-color fill stroke
		`),
		"Numeric": strings.Fields(`
			opacity flex flex-grow flex-shrink order z-index font-weight line-height column-count orphans widows
			fill-opacity stroke-opacity
		`),
		"BorderStroke": strings.Fields("border border-top border-right border-bottom border-left"),
		"Transform":    strings.Fields("transform"),
		"Spacing": strings.Fields(`
			padding margin gap border-radius padding-block padding-inline margin-block margin-inline
		`),
	}
}

func keywordSpecs() map[string][]string {
	return map[string][]string{
		"display": strings.Fields(`
			none block inline inline-block flex inline-flex grid inline-grid table table-row table-cell
			contents
		`),
		"flex-direction": strings.Fields("row row-reverse column column-reverse"),
		"justify-content": strings.Fields(`
			normal start end flex-start flex-end center space-between space-around space-evenly stretch
		`),
		"align-items": strings.Fields("normal start end flex-start flex-end center baseline stretch"),
		"align-self": strings.Fields(`
			auto normal start end flex-start flex-end center baseline stretch
		`),
		"align-content": strings.Fields(`
			normal start end flex-start flex-end center baseline stretch space-between space-around
			space-evenly
		`),
		"flex-wrap":       strings.Fields("nowrap wrap wrap-reverse"),
		"position":        strings.Fields("static relative absolute fixed sticky"),
		"overflow":        strings.Fields("visible hidden clip scroll auto"),
		"overflow-x":      strings.Fields("visible hidden clip scroll auto"),
		"overflow-y":      strings.Fields("visible hidden clip scroll auto"),
		"box-sizing":      strings.Fields("border-box content-box"),
		"border-collapse": strings.Fields("collapse separate"),
		"text-decoration": strings.Fields("none underline line-through"),
		"text-align":      strings.Fields("start end left right center justify"),
		"white-space":     strings.Fields("normal nowrap pre pre-wrap pre-line break-spaces"),
		"overflow-wrap":   strings.Fields("normal break-word anywhere"),
		"font-weight":     strings.Fields("normal bold bolder lighter"),
		"font-family":     strings.Fields("serif sans-serif monospace"),
		"font-style":      strings.Fields("normal italic oblique"),
		"visibility":      strings.Fields("visible hidden collapse"),
		"text-overflow":   strings.Fields("clip ellipsis"),
		"width":           strings.Fields("auto min-content max-content fit-content"),
		"height":          strings.Fields("auto min-content max-content fit-content"),
	}
}

func goName(property string) string {
	parts := strings.Split(strings.TrimPrefix(property, "-"), "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}

func generateStyles(buf *bytes.Buffer, props []string) error {
	families := valueFamilies()
	keywords := keywordSpecs()
	names := make(map[string]string, len(props))
	for _, prop := range props {
		name := goName(prop)
		if prior, ok := names[name]; ok {
			return fmt.Errorf("method collision: %s and %s", prior, prop)
		}
		names[name] = prop
	}
	for prop := range keywords {
		if _, ok := names[goName(prop)]; !ok {
			return fmt.Errorf("keyword metadata for unsupported property %s", prop)
		}
	}
	for family, list := range families {
		for _, prop := range list {
			if _, ok := names[goName(prop)]; !ok {
				return fmt.Errorf("%s metadata for unsupported property %s", family, prop)
			}
		}
	}
	for _, prop := range props {
		name := goName(prop)
		fmt.Fprintf(buf, "\n// %sValue is a value for the CSS %s property.\n",
			name,
			prop,
		)
		fmt.Fprintf(buf, "type %sValue interface { cssValue; is%s() }\n",
			name,
			name,
		)
		fmt.Fprintf(buf, "\n// %s sets the CSS %s property. Use RawValue for complex CSS syntax.\n",
			name,
			prop,
		)
		fmt.Fprintf(buf, "func (e *Element) %s(v %sValue) *Element { return e.setValue(%q, v) }\n",
			name,
			name,
			prop,
		)
		for _, universal := range []string{"expression", "wideKeyword"} {
			fmt.Fprintf(buf, "func (%s) is%s() {}\n",
				universal,
				name,
			)
		}
		for _, family := range slices.Sorted(maps.Keys(families)) {
			for _, p := range families[family] {
				if p == prop {
					fmt.Fprintf(buf, "func (%s) is%s() {}\n",
						family,
						name,
					)
				}
			}
		}
		values := keywords[prop]
		if len(values) == 0 {
			continue
		}
		typ := strings.ToLower(name[:1]) + name[1:] + "Keyword"
		fmt.Fprintf(buf, "type %s uint16\nconst (\n", typ)
		for i, val := range values {
			fmt.Fprintf(buf, "// %s%s sets %s to %s.\n%s%s %s = %d\n",
				name,
				goName(val),
				prop,
				val,
				name,
				goName(val),
				typ,
				i+1,
			)
		}
		fmt.Fprintf(buf, ")\nfunc (%s) is%s() {}\nfunc (v %s) cssText() (string, error) { switch v {\n",
			typ,
			name,
			typ,
		)
		for _, val := range values {
			fmt.Fprintf(buf, "case %s%s: return %q, nil\n",
				name,
				goName(val),
				val,
			)
		}
		fmt.Fprintf(buf, "default: return \"\", fmt.Errorf(\"ui: invalid %s keyword %%d\", v) } }\n", prop)
	}
	return nil
}
