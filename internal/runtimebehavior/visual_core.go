package runtimebehavior

import "bytes"

// VisualCompareHook compares full and partial repaint bytes.
type VisualCompareHook func(full, partial []byte) bool

// VisualComparer stores full and partial repaint output for comparison.
type VisualComparer struct {
	Full        []byte
	Partial     []byte
	Hook        VisualCompareHook
	CompareHook VisualCompareHook
	OnMismatch  func()
	Mismatch    bool
	Equal       bool
}

// VisualCompare is an alias for VisualComparer.
type VisualCompare = VisualComparer

// NewVisualComparer builds an empty comparer.
func NewVisualComparer() *VisualComparer {
	return &VisualComparer{}
}

// NewVisualCompare builds an empty comparer.
func NewVisualCompare() *VisualComparer {
	return NewVisualComparer()
}

// SetFull stores full repaint bytes.
func (v *VisualComparer) SetFull(full []byte) {
	if v == nil {
		return
	}
	v.Full = append([]byte(nil), full...)
}

// SetPartial stores partial repaint bytes.
func (v *VisualComparer) SetPartial(partial []byte) {
	if v == nil {
		return
	}
	v.Partial = append([]byte(nil), partial...)
}

// SetImages stores both byte outputs.
func (v *VisualComparer) SetImages(full, partial []byte) {
	if v == nil {
		return
	}
	v.SetFull(full)
	v.SetPartial(partial)
}

// Compare compares stored outputs.
func (v *VisualComparer) Compare() bool {
	if v == nil {
		return true
	}
	equal := bytes.Equal(v.Full, v.Partial)
	hook := v.Hook
	if hook == nil {
		hook = v.CompareHook
	}
	if hook != nil {
		equal = hook(v.Full, v.Partial)
	}
	v.Equal = equal
	v.Mismatch = !equal
	if !equal && v.OnMismatch != nil {
		v.OnMismatch()
	}
	return equal
}
