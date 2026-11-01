package ui

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

type cssValue interface{ cssText() (string, error) }

type valueData struct {
	text string
	err  error
}

func (v valueData) cssText() (string, error) {
	if v.err != nil {
		return "", v.err
	}
	if v.text == "" {
		return "", fmt.Errorf("ui: uninitialized CSS value")
	}
	return v.text, nil
}

// Length is a CSS length constructed with Px or Rem. Its zero value is invalid.
type Length struct{ valueData }

// Percentage is a CSS percentage. Its zero value is invalid.
type Percentage struct{ valueData }

// Color is a CSS color constructed with Hex. Its zero value is invalid.
type Color struct{ valueData }

// Numeric is a unitless CSS number constructed with Number. Its zero value is invalid.
type Numeric struct{ valueData }

// Spacing is one to four lengths constructed with Edges. Its zero value is invalid.
type Spacing struct{ valueData }

// Px returns a length in CSS pixels. Non-finite inputs fail at Render or NewPage.
func Px(n float64) Length { return Length{valueData: scalar(n, "px")} }

// Rem returns a length relative to the root font size.
func Rem(n float64) Length { return Length{valueData: scalar(n, "rem")} }

// Percent returns a percentage, where Percent(100) means 100%.
func Percent(n float64) Percentage { return Percentage{valueData: scalar(n, "%")} }

// Number returns a unitless number. The engine checks property-specific ranges.
func Number(n float64) Numeric { return Numeric{valueData: scalar(n, "")} }

func scalar(n float64, unit string) valueData {
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return valueData{err: fmt.Errorf("ui: CSS number must be finite")}
	}
	return valueData{text: strconv.FormatFloat(n, 'f', -1, 64) + unit}
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

func (e *Element) setValue(property string, v cssValue) *Element {
	if e == nil {
		return e
	}
	if v == nil || isNilCSSValue(v) {
		e.recordError(fmt.Errorf("ui: nil value for CSS property %q", property))
		return e
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
	if e.styles == nil {
		e.styles = make(map[string]string)
	}
	e.styles[property] = value
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
