//go:build windows

package filepick

import (
	"context"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	ofnNoChangeDir   = 0x00000008
	ofnPathMustExist = 0x00000800
	ofnFileMustExist = 0x00001000
	ofnExplorer      = 0x00080000
)

// openFileName is OPENFILENAMEW from commdlg.h.
type openFileName struct {
	StructSize    uint32
	Owner         uintptr
	Instance      uintptr
	Filter        *uint16
	CustomFilter  *uint16
	MaxCustFilter uint32
	FilterIndex   uint32
	File          *uint16
	MaxFile       uint32
	FileTitle     *uint16
	MaxFileTitle  uint32
	InitialDir    *uint16
	Title         *uint16
	Flags         uint32
	FileOffset    uint16
	FileExtension uint16
	DefExt        *uint16
	CustData      uintptr
	Hook          uintptr
	TemplateName  *uint16
	ReservedPtr   uintptr
	ReservedDword uint32
	FlagsEx       uint32
}

// System32 only (LOAD_LIBRARY_SEARCH_SYSTEM32): a planted DLL in
// the app dir or on PATH cannot win the load.
var procGetOpenFileName = windows.NewLazySystemDLL("comdlg32.dll").NewProc("GetOpenFileNameW")

// Pick shows the common open dialog.
func Pick(ctx context.Context, title string) (string, error) {
	if err := requireCtx(ctx); err != nil {
		return "", err
	}
	buf := make([]uint16, 4096)
	ofn := openFileName{
		StructSize: uint32(unsafe.Sizeof(openFileName{})),
		File:       &buf[0],
		MaxFile:    uint32(len(buf)),
		Flags:      ofnExplorer | ofnFileMustExist | ofnPathMustExist | ofnNoChangeDir,
	}
	if title != "" {
		if t, err := syscall.UTF16PtrFromString(title); err == nil {
			ofn.Title = t
		}
	}
	ok, _, _ := procGetOpenFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	if ok == 0 {
		return "", ErrCanceled
	}
	picked := syscall.UTF16ToString(buf)
	if picked == "" {
		return "", ErrCanceled
	}

	return picked, nil
}
