package state

import "reflect"

func valueAtPath(data any, path string) any {
	if path == "" {
		return data
	}
	segs, ok := splitPath(path)
	if !ok {
		return nil
	}
	cur := data
	for _, seg := range segs {
		switch v := cur.(type) {
		case map[string]any:
			next, ok := v[seg]
			if !ok {
				return nil
			}
			cur = next
		default:
			rv := reflect.ValueOf(cur)
			for rv.Kind() == reflect.Pointer && !rv.IsNil() {
				rv = rv.Elem()
			}
			if rv.Kind() != reflect.Struct {
				return nil
			}
			f := rv.FieldByName(seg)
			if !f.IsValid() {
				return nil
			}
			cur = f.Interface()
		}
	}
	return cur
}
