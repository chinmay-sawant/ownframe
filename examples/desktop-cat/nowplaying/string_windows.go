package nowplaying

import (
	"errors"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

// System32 only (LOAD_LIBRARY_SEARCH_SYSTEM32): a planted DLL in the
// app dir or on PATH cannot win the load.
var combase = windows.NewLazySystemDLL("combase.dll")
var createString = combase.NewProc("WindowsCreateString")
var deleteString = combase.NewProc("WindowsDeleteString")
var stringBuffer = combase.NewProc("WindowsGetStringRawBuffer")

func hstring(text string) (uintptr, error) {
	value, err := syscall.UTF16FromString(text)
	if err != nil {
		return 0, err
	}
	var handle uintptr
	code, _, _ := createString.Call(uintptr(unsafe.Pointer(&value[0])), uintptr(len(value)-1), uintptr(unsafe.Pointer(&handle)))
	return handle, result(code)
}

func (o *object) text(index int) (string, error) {
	var handle uintptr
	if err := o.call(index, uintptr(unsafe.Pointer(&handle))); err != nil {
		return "", err
	}
	defer deleteString.Call(handle)
	var length uint32
	pointer, _, _ := stringBuffer.Call(handle, uintptr(unsafe.Pointer(&length)))
	if length == 0 {
		return "", nil
	}
	if length > 4096 {
		return "", errors.New("media string too long")
	}
	value := make([]uint16, length)
	windows.NewLazySystemDLL("ntdll.dll").NewProc("RtlMoveMemory").Call(uintptr(unsafe.Pointer(&value[0])), pointer, uintptr(length)*2)
	return string(utf16.Decode(value)), nil
}
