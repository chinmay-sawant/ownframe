package atspi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"sort"
)

// dbusEncoder appends D-Bus values to a message body. Every method
// accounts for the alignment the wire format requires, using the
// buffer length as the absolute body offset.
type dbusEncoder struct {
	buf bytes.Buffer
}

func (e *dbusEncoder) align(n int) {
	pad(&e.buf, n)
}

func (e *dbusEncoder) stringVal(s string) {
	e.align(dbusAlign4)
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(s))) //nolint:gosec // reason: values are bounded by maxBodyLen, fits uint32.
	e.buf.Write(lenBuf[:])
	e.buf.WriteString(s)
	e.buf.WriteByte(0)
}

func (e *dbusEncoder) signatureVal(sig string) {
	e.buf.WriteByte(byte(len(sig))) //nolint:gosec // reason: signatures are a few chars, always fits a byte.
	e.buf.WriteString(sig)
	e.buf.WriteByte(0)
}

func (e *dbusEncoder) uint32Val(v uint32) {
	e.align(dbusAlign4)
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], v)
	e.buf.Write(buf[:])
}

func (e *dbusEncoder) int32Val(v int32) {
	e.uint32Val(uint32(v)) //nolint:gosec // reason: D-Bus int32 is the same wire width and two's-complement bits.
}

func (e *dbusEncoder) int16Val(v int16) {
	e.align(dbusAlign2)
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], uint16(v)) //nolint:gosec // reason: D-Bus int16 is the same wire width and two's-complement bits.
	e.buf.Write(buf[:])
}

func (e *dbusEncoder) boolVal(v bool) {
	if v {
		e.uint32Val(1)
		return
	}
	e.uint32Val(0)
}

func (e *dbusEncoder) doubleVal(v float64) {
	e.align(dbusAlign8)
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], math.Float64bits(v))
	e.buf.Write(buf[:])
}

// variant writes a type signature and then the value aligned to the
// contained type.
func (e *dbusEncoder) variant(sig string, write func(*dbusEncoder)) {
	e.signatureVal(sig)
	e.align(dbusAlignment(sig))
	write(e)
}

// array writes the payload length in bytes, pads to the element
// alignment, and then calls write once for the whole array payload.
func (e *dbusEncoder) array(sig string, write func(*dbusEncoder)) {
	e.align(dbusAlign4)
	lenPos := e.buf.Len()
	var zero [4]byte
	e.buf.Write(zero[:])
	e.align(dbusAlignment(sig))
	start := e.buf.Len()
	write(e)
	binary.LittleEndian.PutUint32(e.buf.Bytes()[lenPos:lenPos+4], uint32(e.buf.Len()-start)) //nolint:gosec // reason: array payloads are bounded by maxBodyLen, fits uint32.
}

// structValue starts at an 8-byte boundary, as the D-Bus struct rules
// require.
func (e *dbusEncoder) structValue(write func(*dbusEncoder)) {
	e.align(dbusAlign8)
	write(e)
}

func (e *dbusEncoder) bytes() []byte {
	return e.buf.Bytes()
}

// dbusAlignment returns the alignment of the first complete type in a
// signature.
func dbusAlignment(sig string) int {
	if sig == "" {
		return 1
	}
	switch sig[0] {
	case 'y', 'g', 'v':
		return 1
	case 'n', 'q':
		return dbusAlign2
	case 'b', 'i', 'u', 's', 'o', 'a', 'h':
		return dbusAlign4
	case 'x', 't', 'd', '(', '{':
		return dbusAlign8
	default:
		return 1
	}
}

type dbusDecoder struct {
	body []byte
	off  int
}

func (d *dbusDecoder) align(n int) error {
	rem := d.off % n
	if rem == 0 {
		return nil
	}
	d.off += n - rem
	if d.off > len(d.body) {
		return errors.New("dbus value: alignment past body")
	}
	return nil
}

func (d *dbusDecoder) stringVal() (string, error) {
	if err := d.align(dbusAlign4); err != nil {
		return "", err
	}
	if d.off+4 > len(d.body) {
		return "", errors.New("dbus value: short string length")
	}
	n := binary.LittleEndian.Uint32(d.body[d.off : d.off+4])
	d.off += 4
	if uint64(n)+1 > uint64(len(d.body)-d.off) { //nolint:gosec // reason: both sides are bounded body lengths.
		return "", errors.New("dbus value: short string data")
	}
	s := string(d.body[d.off : d.off+int(n)])
	d.off += int(n) + 1
	return s, nil
}

func (d *dbusDecoder) uint32Val() (uint32, error) {
	if err := d.align(dbusAlign4); err != nil {
		return 0, err
	}
	if d.off+4 > len(d.body) {
		return 0, errors.New("dbus value: short uint32")
	}
	v := binary.LittleEndian.Uint32(d.body[d.off : d.off+4])
	d.off += 4
	return v, nil
}

func (d *dbusDecoder) int32Val() (int32, error) {
	v, err := d.uint32Val()
	return int32(v), err //nolint:gosec // reason: the bits are reinterpreted as a signed D-Bus int32.
}

// variant returns the contained signature and a decoder for its value.
// The caller only needs scalar variants, so the value decoder shares
// the parent body and starts at the aligned offset.
func (d *dbusDecoder) variant() (string, *dbusDecoder, error) {
	if d.off >= len(d.body) {
		return "", nil, errors.New("dbus value: short variant signature")
	}
	n := int(d.body[d.off])
	d.off++
	if d.off+n+1 > len(d.body) {
		return "", nil, errors.New("dbus value: short variant signature data")
	}
	sig := string(d.body[d.off : d.off+n])
	d.off += n + 1
	if err := d.align(dbusAlignment(sig)); err != nil {
		return "", nil, err
	}
	return sig, d, nil
}

func encodeStringValue(s string) []byte {
	var e dbusEncoder
	e.stringVal(s)
	return e.bytes()
}

func encodeUint32Value(v uint32) []byte {
	var e dbusEncoder
	e.uint32Val(v)
	return e.bytes()
}

func encodeInt32Value(v int32) []byte {
	var e dbusEncoder
	e.int32Val(v)
	return e.bytes()
}

func encodeBoolValue(v bool) []byte {
	var e dbusEncoder
	e.boolVal(v)
	return e.bytes()
}

func encodeStructSO(name, path string) []byte {
	var e dbusEncoder
	e.structValue(func(e *dbusEncoder) {
		e.stringVal(name)
		e.stringVal(path)
	})
	return e.bytes()
}

func encodeArraySO(refs [][2]string) []byte {
	var e dbusEncoder
	e.array("(so)", func(e *dbusEncoder) {
		for _, ref := range refs {
			e.structValue(func(e *dbusEncoder) {
				e.stringVal(ref[0])
				e.stringVal(ref[1])
			})
		}
	})
	return e.bytes()
}

func encodeArrayUint32(vals []uint32) []byte {
	var e dbusEncoder
	e.array("u", func(e *dbusEncoder) {
		for _, v := range vals {
			e.uint32Val(v)
		}
	})
	return e.bytes()
}

func encodeArrayString(vals []string) []byte {
	var e dbusEncoder
	e.array("s", func(e *dbusEncoder) {
		for _, v := range vals {
			e.stringVal(v)
		}
	})
	return e.bytes()
}

func encodeEmptyRelationSet() []byte {
	var e dbusEncoder
	e.array("(ua(so))", func(*dbusEncoder) {})
	return e.bytes()
}

func encodeStringMap(vals map[string]string) []byte {
	var e dbusEncoder
	keys := make([]string, 0, len(vals))
	for k := range vals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	e.array("{ss}", func(e *dbusEncoder) {
		for _, k := range keys {
			e.align(dbusAlign8)
			e.stringVal(k)
			e.stringVal(vals[k])
		}
	})
	return e.bytes()
}

func encodeRect(x, y, w, h int32) []byte {
	var e dbusEncoder
	e.structValue(func(e *dbusEncoder) {
		e.int32Val(x)
		e.int32Val(y)
		e.int32Val(w)
		e.int32Val(h)
	})
	return e.bytes()
}

func encodeActionArray(actions [][3]string) []byte {
	var e dbusEncoder
	e.array("(sss)", func(e *dbusEncoder) {
		for _, a := range actions {
			e.structValue(func(e *dbusEncoder) {
				e.stringVal(a[0])
				e.stringVal(a[1])
				e.stringVal(a[2])
			})
		}
	})
	return e.bytes()
}

// decodeSO decodes a (so) struct body.
func decodeSO(body []byte) (string, string, error) {
	d := dbusDecoder{body: body}
	if err := d.align(dbusAlign8); err != nil {
		return "", "", err
	}
	name, err := d.stringVal()
	if err != nil {
		return "", "", err
	}
	path, err := d.stringVal()
	if err != nil {
		return "", "", err
	}
	return name, path, nil
}

func encodePropertyValue(sig string, write func(*dbusEncoder)) []byte {
	var e dbusEncoder
	e.variant(sig, write)
	return e.bytes()
}

func encodeErrorBody(message string) []byte {
	return encodeStringValue(message)
}
