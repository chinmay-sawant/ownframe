package page

import "reflect"

// bindTarget returns the settable field name on the struct p.data points at.
// A missing or unexported field is false.
func bindTarget(p *Page, name string) (reflect.Value, bool) {
	if p == nil || name == "" || p.data == nil {
		return reflect.Value{}, false
	}

	v := reflect.ValueOf(p.data)
	if v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	f := v.Elem().FieldByName(name)
	if !f.IsValid() || !f.CanSet() {
		return reflect.Value{}, false
	}

	return f, true
}

// bindKind is 'b' for a checkbox bool field, 's' for a string field, or 0.
// Text-like inputs, textareas, selects, and radios are 's'.
func bindKind(c Control) byte {
	if c.Type == "checkbox" {
		return 'b'
	}
	if c.Tag == "textarea" || c.Tag == "select" || c.Type == "radio" || textLike(c.Type) {
		return 's'
	}

	return 0
}

// bindRead copies the bound field into c and reports whether c is bound.
// A radio is checked when the string field holds the control value.
func bindRead(p *Page, c Control) (Control, bool) {
	f, ok := bindTarget(p, c.Bind)
	if !ok {
		return c, false
	}

	switch bindKind(c) {
	case 'b':
		if f.Kind() != reflect.Bool {
			return c, false
		}
		c.Checked = f.Bool()
	case 's':
		if f.Kind() != reflect.String {
			return c, false
		}
		if c.Type == "radio" {
			c.Checked = f.String() == c.Value
		} else {
			c.Value = f.String()
		}
	default:
		return c, false
	}

	return c, true
}

// bindWrite copies c into the bound field and reports whether c is bound.
// An unchecked radio writes nothing.
func bindWrite(p *Page, c Control) bool {
	f, ok := bindTarget(p, c.Bind)
	if !ok {
		return false
	}

	switch bindKind(c) {
	case 'b':
		if f.Kind() != reflect.Bool {
			return false
		}
		f.SetBool(c.Checked)
		p.formDirty = true
	case 's':
		if f.Kind() != reflect.String {
			return false
		}
		if c.Type != "radio" || c.Checked {
			f.SetString(c.Value)
			p.formDirty = true
		}
	default:
		return false
	}

	return true
}
