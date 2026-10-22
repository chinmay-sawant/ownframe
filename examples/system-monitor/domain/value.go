package domain

import "strconv"

// Value is one measurement. Valid separates a real zero from a reading the
// source could not take.
type Value struct {
	N     float64
	Valid bool
}

// Get returns the number and whether it is valid.
func (v Value) Get() (float64, bool) { return v.N, v.Valid }

// Or returns the number when valid, else fallback.
func (v Value) Or(fallback float64) float64 {
	if !v.Valid {
		return fallback
	}

	return v.N
}

// String renders the value for the UI. An invalid value prints as "-" so a
// missing sensor never looks like a zero.
func (v Value) String() string {
	if !v.Valid {
		return "-"
	}

	return strconv.FormatFloat(v.N, 'f', 2, 64)
}

// Percent builds a 0 to 100 reading.
func Percent(n float64) Value { return Value{N: n, Valid: true} }

// percentScale turns a 0..1 ratio into a 0..100 percent reading.
const percentScale = 100

// Bytes builds a byte count reading.
func Bytes(n float64) Value { return Value{N: n, Valid: true} }

// Rate builds a bytes per second reading.
func Rate(n float64) Value { return Value{N: n, Valid: true} }

// Count builds a plain count reading.
func Count(n float64) Value { return Value{N: n, Valid: true} }

// Celsius builds a temperature reading.
func Celsius(n float64) Value { return Value{N: n, Valid: true} }
