package nowplaying

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type object struct{ table *[40]uintptr }

func result(code uintptr) error {
	if int32(code) < 0 {
		return fmt.Errorf("Windows media HRESULT 0x%08X", uint32(code))
	}
	return nil
}

func (o *object) call(index int, args ...uintptr) error {
	if o == nil {
		return fmt.Errorf("media object is nil")
	}
	if index < 0 || index >= len(o.table) {
		return fmt.Errorf("media vtable index %d out of range", index)
	}
	values := append([]uintptr{uintptr(unsafe.Pointer(o))}, args...)
	code, _, _ := syscall.SyscallN(o.table[index], values...)
	runtime.KeepAlive(o)
	return result(code)
}

func (o *object) release() {
	if o != nil {
		_ = o.call(2)
	}
}

func (o *object) get(index int) (*object, error) {
	if o == nil {
		return nil, fmt.Errorf("media object is nil")
	}
	var value *object
	err := o.call(index, uintptr(unsafe.Pointer(&value)))
	return value, err
}

func (o *object) query(id string) (*object, error) {
	if o == nil {
		return nil, fmt.Errorf("media object is nil")
	}
	guid, err := windows.GUIDFromString(id)
	if err != nil {
		return nil, err
	}
	var value *object
	err = o.call(0, uintptr(unsafe.Pointer(&guid)), uintptr(unsafe.Pointer(&value)))
	return value, err
}
