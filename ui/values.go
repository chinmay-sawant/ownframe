package ui

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

type cssValue interface{ cssText() (string, error) }

// unitKind identifies the unit of a single-number CSS value. unitText marks a
// value that is already spelled as text, so the zero valueData stays invalid.
type unitKind uint8

const (
	unitText unitKind = iota
	unitNone
	unitPx
	unitPercent
	unitVh
	unitRem
)

// suffix returns the unit suffix appended to a numeric value.
func (u unitKind) suffix() string {
	switch u {
	case unitNone:
		return ""
	case unitPx:
		return "px"
	case unitPercent:
		return "%"
	case unitVh:
		return "vh"
	case unitRem:
		return "rem"
	default:
		return ""
	}
}

// valueData is the value behind every typed CSS constructor. A single-number
// value keeps its number and unit so the element stores it unformatted; text
// values (colors, keywords, composites) keep their CSS spelling.
type valueData struct {
	text string
	num  float64
	unit unitKind
	err  error
}

func (v valueData) cssText() (string, error) {
	if v.err != nil {
		return "", v.err
	}
	if v.unit != unitText {
		return formatScalar(v.num, v.unit), nil
	}
	if v.text == "" {
		return "", fmt.Errorf("ui: uninitialized CSS value")
	}
	return v.text, nil
}

// cssData exposes the value without formatting it. It is promoted to every
// typed value, so setValue can store a single-number value numerically and
// leave the conversion to text for the renderer.
func (v valueData) cssData() valueData { return v }

// floatBits is the bit size of every formatted value: the numeric
// constructors all hold float64.
const floatBits = 64

// appendScalar appends the CSS text of one numeric value. The bytes match
// strconv.FormatFloat(n, 'f', -1, 64) followed by the unit suffix, which is
// how Px, Percent, Vh, Rem, and Number have always rendered.
func appendScalar(dst []byte, n float64, unit unitKind) []byte {
	dst = strconv.AppendFloat(dst, n, 'f', -1, floatBits)
	return append(dst, unit.suffix()...)
}

// formatScalar returns one numeric value as a string. Composite constructors
// use it; rendering appends into a reused buffer instead.
func formatScalar(n float64, unit unitKind) string {
	var buf [32]byte
	return string(appendScalar(buf[:0], n, unit))
}

// Length is a CSS length constructed with Px, Rem, Vh, or LengthVar.
// Its zero value is invalid.
type Length struct{ valueData }

// Percentage is a CSS percentage. Its zero value is invalid.
type Percentage struct{ valueData }

// Color is a CSS color constructed with Hex, Transparent, or ColorVar.
// Its zero value is invalid.
type Color struct{ valueData }

// Numeric is a unitless CSS number constructed with Number. Its zero value is invalid.
type Numeric struct{ valueData }

// Spacing is one to four lengths constructed with Edges. Its zero value is invalid.
type Spacing struct{ valueData }

// Px returns a length in CSS pixels. Non-finite inputs fail at Render or NewPage.
func Px(n float64) Length { return Length{valueData: scalar(n, unitPx)} }

// Vh returns a length as a percentage of the viewport height.
func Vh(n float64) Length { return Length{valueData: scalar(n, unitVh)} }

// Rem returns a length relative to the root font size.
func Rem(n float64) Length { return Length{valueData: scalar(n, unitRem)} }

// Percent returns a percentage, where Percent(100) means 100%.
func Percent(n float64) Percentage { return Percentage{valueData: scalar(n, unitPercent)} }

// Number returns a unitless number. The engine checks property-specific ranges.
func Number(n float64) Numeric { return Numeric{valueData: scalar(n, unitNone)} }

func scalar(n float64, unit unitKind) valueData {
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return valueData{err: fmt.Errorf("ui: CSS number must be finite")}
	}
	return valueData{num: n, unit: unit}
}

// Edges returns a spacing shorthand in CSS order: top, right, bottom, left.
// Supply one to four lengths; fewer values follow CSS shorthand expansion.
func Edges(values ...Length) Spacing {
	const maxEdges = 4
	if len(values) == 0 || len(values) > maxEdges {
		return Spacing{valueData: valueData{err: fmt.Errorf("ui: Edges needs one to four lengths")}}
	}
	parts := make([]string, 0, len(values))
	for _, v := range values {
		text, err := v.cssText()
		if err != nil {
			return Spacing{valueData: valueData{err: err}}
		}
		parts = append(parts, text)
	}
	return Spacing{valueData: valueData{text: strings.Join(parts, " ")}}
}

// Hex returns a CSS hex color with 3, 4, 6, or 8 hex digits, including the # prefix.
func Hex(text string) Color {
	const shortRGB, shortRGBA, fullRGB, fullRGBA = 4, 5, 7, 9
	valid := len(text) == shortRGB || len(text) == shortRGBA || len(text) == fullRGB || len(text) == fullRGBA
	if valid && text[0] == '#' {
		for _, c := range text[1:] {
			if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
				valid = false
				break
			}
		}
	} else {
		valid = false
	}
	if !valid {
		return Color{valueData: valueData{err: fmt.Errorf("ui: invalid CSS hex color %q", text)}}
	}
	return Color{valueData: valueData{text: text}}
}

// Transparent returns a fully transparent CSS color.
func Transparent() Color { return Color{valueData: valueData{text: "transparent"}} }

// ColorVar references a theme color with a validated color fallback.
// The theme stylesheet supplies the variable's runtime value.
func ColorVar(name string, fallback Color) Color {
	return Color{valueData: variableValue(name, fallback)}
}

// LengthVar references a theme length with a validated length fallback.
func LengthVar(name string, fallback Length) Length {
	return Length{valueData: variableValue(name, fallback)}
}

func variableValue(name string, fallback cssValue) valueData {
	text, err := fallback.cssText()
	if err != nil {
		return valueData{err: err}
	}
	return Var(name, text).valueData
}

// BorderStroke is a solid border constructed by SolidBorder.
type BorderStroke struct{ valueData }

// SolidBorder combines a width and color without raw CSS syntax.
func SolidBorder(width Length, color Color) BorderStroke {
	w, err := width.cssText()
	if err != nil {
		return BorderStroke{valueData: valueData{err: err}}
	}
	c, err := color.cssText()
	if err != nil {
		return BorderStroke{valueData: valueData{err: err}}
	}
	return BorderStroke{valueData: valueData{text: w + " solid " + c}}
}

// Transform is a typed transform constructed by TranslateY.
type Transform struct{ valueData }

// TranslateY moves an element vertically by a CSS length.
func TranslateY(distance Length) Transform {
	text, err := distance.cssText()
	if err != nil {
		return Transform{valueData: valueData{err: err}}
	}
	return Transform{valueData: valueData{text: "translateY(" + text + ")"}}
}

type expression struct{ valueData }

// RawValue explicitly opts out of property-specific value typing. It accepts a
// single declaration value, not additional declarations or rules. Blinkless
// checks CSS semantics; this helper checks declaration boundaries only.
func RawValue(text string) expression {
	return expression{valueData: valueData{text: text, err: validateCSSValue(text)}}
}

// Var references a custom property. An optional fallback is one CSS value.
// Invalid names and values fail at Render or NewPage. Variable resolution and
// cycles follow Blinkless semantics, not compile-time validation.
func Var(name string, fallback ...string) expression {
	if !validCustomProperty(name) {
		return expression{valueData: valueData{err: fmt.Errorf("ui: invalid custom property name %q", name)}}
	}
	if len(fallback) > 1 {
		return expression{valueData: valueData{err: fmt.Errorf("ui: Var accepts at most one fallback")}}
	}
	text := "var(" + name
	if len(fallback) == 1 {
		if err := validateCSSValue(fallback[0]); err != nil {
			return expression{valueData: valueData{err: err}}
		}
		text += ", " + fallback[0]
	}
	return RawValue(text + ")")
}

type wideKeyword uint8

const (
	// Inherit takes the parent's computed value.
	Inherit wideKeyword = iota + 1
	// Initial takes the property's initial value.
	Initial
	// Unset selects inherit or initial according to the property.
	Unset
	// Revert rolls back the current cascade origin where the engine supports it.
	Revert
	// RevertLayer rolls back the current cascade layer where the engine supports it.
	RevertLayer
)

func (v wideKeyword) cssText() (string, error) {
	switch v {
	case Inherit:
		return "inherit", nil
	case Initial:
		return "initial", nil
	case Unset:
		return "unset", nil
	case Revert:
		return "revert", nil
	case RevertLayer:
		return "revert-layer", nil
	default:
		return "", fmt.Errorf("ui: invalid CSS-wide keyword %d", v)
	}
}

// inlineStyleEntries is the number of declarations an element keeps without a
// separate allocation. Most elements in a generated tree set two to four.
const inlineStyleEntries = 4

// styleEntry is one CSS declaration. A single-number value keeps its number
// and unit so the renderer formats it straight into the output buffer; a
// composite or raw value keeps its text.
type styleEntry struct {
	name string
	text string
	num  float64
	unit unitKind
}

// appendValue appends the declaration's CSS value to dst.
func (e styleEntry) appendValue(dst []byte) []byte {
	if e.unit == unitText {
		return append(dst, e.text...)
	}
	return appendScalar(dst, e.num, e.unit)
}

// cssTextValue returns the declaration's CSS value as a string. Composite
// constructors and patch comparisons use it; rendering appends instead.
func (e styleEntry) cssTextValue() string {
	if e.unit == unitText {
		return e.text
	}
	return formatScalar(e.num, e.unit)
}

// equal reports whether two entries name the same declaration with the same
// value. Numeric values compare bit-for-bit so Px(-0) and Px(0) stay distinct,
// and a numeric value spelled like a raw one, such as Px(10) and
// RawValue("10px"), compares equal, matching the string maps this replaced.
func (e styleEntry) equal(other styleEntry) bool {
	if e.name != other.name {
		return false
	}
	if e.unit == other.unit {
		if e.unit == unitText {
			return e.text == other.text
		}
		return math.Float64bits(e.num) == math.Float64bits(other.num)
	}
	return e.cssTextValue() == other.cssTextValue()
}

// styleSet stores an element's declarations in insertion order. The first
// inlineStyleEntries live in the element itself, so an ordinary element
// allocates nothing; later entries spill into a slice.
type styleSet struct {
	inline [inlineStyleEntries]styleEntry
	extra  []styleEntry
	n      int
}

func (s *styleSet) len() int { return s.n }

// at returns the entry at index i, which must be within len.
func (s *styleSet) at(i int) styleEntry {
	if i < len(s.inline) {
		return s.inline[i]
	}
	return s.extra[i-len(s.inline)]
}

// get returns the entry named name.
func (s *styleSet) get(name string) (styleEntry, bool) {
	for i := 0; i < s.n; i++ {
		if entry := s.at(i); entry.name == name {
			return entry, true
		}
	}
	return styleEntry{}, false
}

// set adds entry, replacing a declaration with the same name.
func (s *styleSet) set(entry styleEntry) {
	for i := 0; i < s.n; i++ {
		if s.at(i).name != entry.name {
			continue
		}
		if i < len(s.inline) {
			s.inline[i] = entry
		} else {
			s.extra[i-len(s.inline)] = entry
		}
		return
	}
	if s.n < len(s.inline) {
		s.inline[s.n] = entry
	} else {
		s.extra = append(s.extra, entry)
	}
	s.n++
}

// equal reports whether two sets hold the same declarations in any order.
func (s *styleSet) equal(other *styleSet) bool {
	if s.n != other.n {
		return false
	}
	for i := 0; i < s.n; i++ {
		entry := s.at(i)
		otherEntry, ok := other.get(entry.name)
		if !ok || !entry.equal(otherEntry) {
			return false
		}
	}
	return true
}

func (e *Element) setValue(property string, v cssValue) *Element {
	if e == nil {
		return e
	}
	if v == nil || isNilCSSValue(v) {
		e.recordError(fmt.Errorf("ui: nil value for CSS property %q", property))
		return e
	}
	// A single-number value stays numeric until the renderer writes it.
	if carrier, ok := v.(interface{ cssData() valueData }); ok {
		if data := carrier.cssData(); data.err == nil && data.unit != unitText {
			return e.setScalarStyle(property, data.num, data.unit)
		}
	}
	text, err := v.cssText()
	if err != nil {
		e.recordError(fmt.Errorf("ui: CSS property %q: %w", property, err))
		return e
	}
	return e.setCSS(property, text)
}

// Public value types also admit pointers through their value-receiver methods.
// Reject typed nil pointers before invoking the interface method.
func isNilCSSValue(v cssValue) bool {
	value := reflect.ValueOf(v)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

func (e *Element) setCSS(property, value string) *Element {
	if e == nil {
		return e
	}
	if err := validateCSSValue(value); err != nil {
		e.recordError(fmt.Errorf("ui: CSS property %q: %w", property, err))
		return e
	}
	e.styles.set(styleEntry{name: property, text: value})
	return e
}

// setScalarStyle stores a single-number declaration without formatting it.
// The constructor already rejected non-finite numbers, so no validation is
// needed here; the renderer writes the number on demand.
func (e *Element) setScalarStyle(property string, number float64, unit unitKind) *Element {
	if e == nil {
		return e
	}
	e.styles.set(styleEntry{name: property, num: number, unit: unit})
	return e
}

// CSSRaw accepts a property and value supplied at runtime. Unknown properties
// fail at Render or NewPage. Prefer generated methods for compile-time name
// checking. Use CustomProperty to declare a CSS variable.
func (e *Element) CSSRaw(property, value string) *Element {
	if e == nil {
		return e
	}
	property = strings.ToLower(strings.TrimSpace(property))
	if !SupportsProperty(property) {
		e.recordError(fmt.Errorf("ui: CSS property %q is not implemented by the compatibility matrix", property))
		return e
	}
	return e.setCSS(property, value)
}

// CustomProperty declares a case-sensitive CSS variable such as --accent.
// Names use the ASCII subset --[A-Za-z0-9_-]+. Values must fit in one declaration.
func (e *Element) CustomProperty(name, value string) *Element {
	if e == nil {
		return e
	}
	if !validCustomProperty(name) {
		e.recordError(fmt.Errorf("ui: invalid custom property name %q", name))
		return e
	}
	return e.setCSS(name, value)
}

func validCustomProperty(name string) bool {
	if !strings.HasPrefix(name, "--") || len(name) <= 2 {
		return false
	}
	for _, c := range name[2:] {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '-' && c != '_' {
			return false
		}
	}
	return true
}

// Check boundaries before concatenating a style attribute. Quoted delimiters
// and delimiters inside functions, such as data URLs, remain valid values.
func validateCSSValue(text string) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("ui: empty CSS value")
	}
	stack := []byte{}
	var quote byte
	for i := 0; i < len(text); i++ {
		c := text[i]
		if c == 0 || (c < ' ' && c != '\t' && c != '\r' && c != '\n') {
			return fmt.Errorf("ui: control character in CSS value")
		}
		if c == '\\' {
			if i+1 >= len(text) {
				return fmt.Errorf("ui: unfinished CSS escape")
			}
			// Escaped newlines and NUL are rejected rather than hiding a boundary.
			next := text[i+1]
			if next < ' ' {
				return fmt.Errorf("ui: invalid CSS escape")
			}
			i++
			continue
		}
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			continue
		}
		if c == '/' && i+1 < len(text) && text[i+1] == '*' {
			return fmt.Errorf("ui: CSS comments are not accepted in declaration values")
		}
		switch c {
		case '{', '}':
			return fmt.Errorf("ui: CSS rule delimiter in declaration value")
		case ';':
			if len(stack) == 0 {
				return fmt.Errorf("ui: multiple CSS declarations in value")
			}
		case '(', '[':
			stack = append(stack, c)
		case ')', ']':
			if len(stack) == 0 {
				return fmt.Errorf("ui: unmatched CSS delimiter")
			}
			opening := stack[len(stack)-1]
			if (c == ')' && opening != '(') || (c == ']' && opening != '[') {
				return fmt.Errorf("ui: mismatched CSS delimiters")
			}
			stack = stack[:len(stack)-1]
		}
	}
	if quote != 0 || len(stack) != 0 {
		return fmt.Errorf("ui: unfinished CSS value")
	}
	return nil
}
