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
	sortedFamilies := slices.Sorted(maps.Keys(families))
	propFamilies := make(map[string][]string, len(props))
	for _, family := range sortedFamilies {
		for _, prop := range families[family] {
			if _, ok := names[goName(prop)]; !ok {
				return fmt.Errorf("%s metadata for unsupported property %s", family, prop)
			}
			propFamilies[prop] = append(propFamilies[prop], family)
		}
	}

	fmt.Fprintln(buf, "\n// Value is a CSS value for properties without property-specific typed helpers.")
	fmt.Fprintln(buf, "type Value interface { cssValue; isValue() }")
	fmt.Fprintln(buf, "func (expression) isValue() {}")
	fmt.Fprintln(buf, "func (wideKeyword) isValue() {}")

	emittedCombos := make(map[string]bool)
	for _, prop := range props {
		if len(keywords[prop]) > 0 || len(propFamilies[prop]) == 0 {
			continue
		}
		combo := propFamilies[prop]
		iface := strings.Join(combo, "") + "Value"
		if emittedCombos[iface] {
			continue
		}
		emittedCombos[iface] = true
		marker := "is" + iface
		fmt.Fprintf(buf, "\ntype %s interface { cssValue; %s() }\n", iface, marker)
		fmt.Fprintf(buf, "func (expression) %s() {}\n", marker)
		fmt.Fprintf(buf, "func (wideKeyword) %s() {}\n", marker)
		for _, family := range combo {
			fmt.Fprintf(buf, "func (%s) %s() {}\n", family, marker)
		}
	}

	for _, prop := range props {
		values := keywords[prop]
		if len(values) == 0 {
			continue
		}
		name := goName(prop)
		iface := name + "Value"
		marker := "is" + name
		typ := strings.ToLower(name[:1]) + name[1:] + "Keyword"
		fmt.Fprintf(buf, "\ntype %s interface { cssValue; %s() }\n", iface, marker)
		fmt.Fprintf(buf, "type %s uint16\nconst (\n", typ)
		for i, val := range values {
			if i == 0 {
				fmt.Fprintf(buf, "%s%s %s = iota + 1\n", name, goName(val), typ)
			} else {
				fmt.Fprintf(buf, "%s%s\n", name, goName(val))
			}
		}
		fmt.Fprintf(buf, ")\nfunc (expression) %s() {}\nfunc (wideKeyword) %s() {}\nfunc (%s) %s() {}\n",
			marker,
			marker,
			typ,
			marker,
		)
		for _, family := range propFamilies[prop] {
			fmt.Fprintf(buf, "func (%s) %s() {}\n", family, marker)
		}
		fmt.Fprintf(buf, "func (v %s) cssText() (string, error) {\nnames := [...]string{\"\"", typ)
		for _, val := range values {
			fmt.Fprintf(buf, ", %q", val)
		}
		fmt.Fprintf(buf, "}\nif v == 0 || int(v) >= len(names) {\nreturn \"\", fmt.Errorf(\"ui: invalid %s keyword %%d\", v)\n}\nreturn names[v], nil\n}\n", prop)
	}

	fmt.Fprintln(buf)
	for _, prop := range props {
		name := goName(prop)
		iface := "Value"
		if len(keywords[prop]) > 0 {
			iface = name + "Value"
		} else if len(propFamilies[prop]) > 0 {
			iface = strings.Join(propFamilies[prop], "") + "Value"
		}
		fmt.Fprintf(buf, "func (e *Element) %s(v %s) *Element { return e.setValue(%q, v) }\n",
			name,
			iface,
			prop,
		)
	}
	return nil
}
