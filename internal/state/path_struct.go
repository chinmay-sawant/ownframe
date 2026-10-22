package state

import "reflect"

func setStructSeg(cur any, segs []string, value any) (any, bool) {
	rv := reflect.ValueOf(cur)
	if rv.Kind() != reflect.Struct {
		return cur, false
	}
	cp := reflect.New(rv.Type()).Elem()
	cp.Set(rv)
	seg := segs[0]
	rest := segs[1:]
	f := cp.FieldByName(seg)
	if !f.IsValid() || !f.CanSet() || !f.CanInterface() {
		return cur, false
	}
	if len(rest) == 0 {
		v := reflect.ValueOf(value)
		if !v.IsValid() {
			f.Set(reflect.Zero(f.Type()))
			return cp.Interface(), true
		}
		if v.Type().AssignableTo(f.Type()) {
			f.Set(v)
			return cp.Interface(), true
		}
		if v.Type().ConvertibleTo(f.Type()) {
			f.Set(v.Convert(f.Type()))
			return cp.Interface(), true
		}
		return cur, false
	}
	if f.Kind() == reflect.Map && f.IsNil() {
		f.Set(reflect.MakeMap(f.Type()))
	}
	nc, ok := setSeg(f.Interface(), rest, value)
	if !ok {
		return cur, false
	}
	fv := reflect.ValueOf(nc)
	if !fv.IsValid() {
		f.Set(reflect.Zero(f.Type()))
		return cp.Interface(), true
	}
	switch {
	case fv.Type().AssignableTo(f.Type()):
		f.Set(fv)
	case fv.Type().ConvertibleTo(f.Type()):
		f.Set(fv.Convert(f.Type()))
	default:
		return cur, false
	}
	return cp.Interface(), true
}
