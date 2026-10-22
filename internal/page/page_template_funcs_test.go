package page

import (
	"reflect"
	"testing"
)

func TestSafeFuncsReturnStrings(t *testing.T) {
	t.Parallel()
	strType := reflect.TypeOf("")
	for name, fn := range SafeFuncs() {
		if got := reflect.TypeOf(fn).Out(0); got != strType {
			t.Fatalf("%s returns %s", name, got)
		}
	}
}
