package state

import "reflect"

func setSeg(cur any, segs []string, value any) (any, bool) {
	if len(segs) == 0 {
		return value, true
	}
	seg := segs[0]
	rest := segs[1:]
	if m, ok := cur.(map[string]any); ok {
		if len(rest) == 0 {
			m[seg] = value
			return m, true
		}
		child, exists := m[seg]
		if !exists {
			child = map[string]any{}
		}
		nc, ok := setSeg(child, rest, value)
		if !ok {
			return cur, false
		}
		m[seg] = nc
		return m, true
	}
	rv := reflect.ValueOf(cur)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return cur, false
		}
		elem := rv.Elem()
		if elem.Kind() != reflect.Struct {
			return cur, false
		}
		f := elem.FieldByName(seg)
		if !f.IsValid() {
			return cur, false
		}
		if len(rest) == 0 {
			v := reflect.ValueOf(value)
			if v.Type().AssignableTo(f.Type()) {
				f.Set(v)
			} else if v.Type().ConvertibleTo(f.Type()) {
				f.Set(v.Convert(f.Type()))
			} else {
				return cur, false
			}
			return cur, true
		}
		nc, ok := setSeg(f.Interface(), rest, value)
		if !ok {
			return cur, false
		}
		fv := reflect.ValueOf(nc)
		if fv.Type().AssignableTo(f.Type()) {
			f.Set(fv)
		} else if fv.Type().ConvertibleTo(f.Type()) {
			f.Set(fv.Convert(f.Type()))
		} else {
			return cur, false
		}
		return cur, true
	}
	if rv.Kind() == reflect.Struct {
		return setStructSeg(cur, segs, value)
	}

	return cur, false
}
