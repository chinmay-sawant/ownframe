package state

import (
	"reflect"
	"strings"
)

func valueAtPath(data any, path string) any {
	if path == "" {
		return data
	}
	cur := data
	for _, seg := range strings.Split(path, ".") {
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
