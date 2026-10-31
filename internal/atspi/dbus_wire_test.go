package atspi

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"
)

func TestReadMessageOversizedFieldsLen(t *testing.T) {
	t.Parallel()

	var hdr [16]byte
	hdr[0] = dbusEndianLittle
	hdr[1] = dbusTypeMethodCall
	hdr[2] = 0 // flags
	hdr[3] = dbusProtoVersion
	binary.LittleEndian.PutUint32(hdr[4:8], 0)                // bodyLen = 0
	binary.LittleEndian.PutUint32(hdr[8:12], 1)               // serial = 1
	binary.LittleEndian.PutUint32(hdr[12:16], maxFieldsLen+1) // fieldsLen > 64KB

	_, err := readMessage(bytes.NewReader(hdr[:]))
	if err == nil {
		t.Fatal("expected error for fieldsLen > 64KB, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds 64KB limit") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestReadMessageOversizedBodyLen(t *testing.T) {
	t.Parallel()

	var hdr [16]byte
	hdr[0] = dbusEndianLittle
	hdr[1] = dbusTypeMethodCall
	hdr[2] = 0 // flags
	hdr[3] = dbusProtoVersion
	binary.LittleEndian.PutUint32(hdr[4:8], maxBodyLen+1) // bodyLen > 4MB
	binary.LittleEndian.PutUint32(hdr[8:12], 1)           // serial = 1
	binary.LittleEndian.PutUint32(hdr[12:16], 0)          // fieldsLen = 0

	_, err := readMessage(bytes.NewReader(hdr[:]))
	if err == nil {
		t.Fatal("expected error for bodyLen > 4MB, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds 4MB limit") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestReadMessageSkipsBigEndianFrame(t *testing.T) {
	t.Parallel()

	wire := encodeBigEndianMethodCall(
		"/org/a11y/atspi/accessible/root",
		"org.a11y.atspi.Accessible",
		"GetName",
		"s",
		encodeBigEndianStringBody("hello"),
	)
	next := encodeMessage(&dbusMsg{
		msgType: dbusTypeMethodCall,
		serial:  9,
		member:  "NameAcquired",
	})
	r := bytes.NewReader(append(wire, next...))

	skipped, err := readMessage(r)
	if err != nil {
		t.Fatalf("big-endian frame: %v", err)
	}
	if skipped == nil || skipped.msgType != 0 || skipped.serial != 0 || len(skipped.body) != 0 {
		t.Fatalf("big-endian frame = %+v, want empty message", skipped)
	}

	decoded, err := readMessage(r)
	if err != nil {
		t.Fatalf("frame after big-endian message: %v", err)
	}
	if decoded.serial != 9 || decoded.member != "NameAcquired" {
		t.Fatalf("next message = %+v", decoded)
	}
}

func TestReadMessageTruncatedBigEndianFrame(t *testing.T) {
	t.Parallel()

	wire := encodeBigEndianMethodCall(
		"/org/a11y/atspi/accessible/root",
		"org.a11y.atspi.Accessible",
		"GetName",
		"s",
		encodeBigEndianStringBody("hello"),
	)
	truncated := wire[:len(wire)-1]

	if _, err := readMessage(bytes.NewReader(truncated)); err == nil {
		t.Fatal("expected error for truncated big-endian frame, got nil")
	}
}

func TestDecodeStringBody(t *testing.T) {
	t.Parallel()

	// Short body (< 4 bytes)
	for _, short := range [][]byte{nil, {}, {1}, {1, 2, 3}} {
		if _, err := decodeStringBody(short); err == nil {
			t.Fatalf("expected error for short body len %d", len(short))
		}
	}

	// Overflow / MaxUint32
	var maxU32 [8]byte
	binary.LittleEndian.PutUint32(maxU32[:4], ^uint32(0))
	if _, err := decodeStringBody(maxU32[:]); err == nil {
		t.Fatal("expected error for MaxUint32 length")
	}

	// Length exceeds limit (> 4MB)
	var oversized [8]byte
	binary.LittleEndian.PutUint32(oversized[:4], maxBodyLen+1)
	if _, err := decodeStringBody(oversized[:]); err == nil {
		t.Fatal("expected error for length > maxBodyLen")
	}

	// Length exceeds available bytes
	var truncated [8]byte
	binary.LittleEndian.PutUint32(truncated[:4], 100)
	if _, err := decodeStringBody(truncated[:]); err == nil {
		t.Fatal("expected error for truncated body")
	}

	// Valid string body
	encoded := encodeStringBody("hello world")
	decoded, err := decodeStringBody(encoded)
	if err != nil {
		t.Fatalf("decodeStringBody failed: %v", err)
	}
	if decoded != "hello world" {
		t.Fatalf("decodeStringBody = %q, want 'hello world'", decoded)
	}
}

func TestDecodeRequestNameBody(t *testing.T) {
	t.Parallel()

	// Short body (< 4 bytes)
	for _, short := range [][]byte{nil, {}, {1}, {1, 2, 3}} {
		if _, _, err := decodeRequestNameBody(short); err == nil {
			t.Fatalf("expected error for short body len %d", len(short))
		}
	}

	// Overflow / MaxUint32
	var maxU32 [8]byte
	binary.LittleEndian.PutUint32(maxU32[:4], ^uint32(0))
	if _, _, err := decodeRequestNameBody(maxU32[:]); err == nil {
		t.Fatal("expected error for MaxUint32 length")
	}

	// Length exceeds limit (> 4MB)
	var oversized [8]byte
	binary.LittleEndian.PutUint32(oversized[:4], maxBodyLen+1)
	if _, _, err := decodeRequestNameBody(oversized[:]); err == nil {
		t.Fatal("expected error for length > maxBodyLen")
	}

	// Length exceeds available bytes
	var truncated [8]byte
	binary.LittleEndian.PutUint32(truncated[:4], 100)
	if _, _, err := decodeRequestNameBody(truncated[:]); err == nil {
		t.Fatal("expected error for truncated body")
	}

	// Valid request name body
	encoded := encodeRequestNameBody("org.a11y.Bus", 0x3)
	name, flags, err := decodeRequestNameBody(encoded)
	if err != nil {
		t.Fatalf("decodeRequestNameBody failed: %v", err)
	}
	if name != "org.a11y.Bus" {
		t.Fatalf("name = %q, want 'org.a11y.Bus'", name)
	}
	if flags != 0x3 {
		t.Fatalf("flags = %d, want 3", flags)
	}
}

func TestDecodeEmbedBody(t *testing.T) {
	t.Parallel()

	// Short body (< 4 bytes)
	for _, short := range [][]byte{nil, {}, {1}, {1, 2, 3}} {
		if _, _, err := decodeEmbedBody(short); err == nil {
			t.Fatalf("expected error for short embed len %d", len(short))
		}
	}

	// Overflow / MaxUint32
	var maxU32 [8]byte
	binary.LittleEndian.PutUint32(maxU32[:4], ^uint32(0))
	if _, _, err := decodeEmbedBody(maxU32[:]); err == nil {
		t.Fatal("expected error for MaxUint32 nameLen")
	}

	// Valid embed body
	encoded := encodeEmbedBody(":1.42", "/org/a11y/atspi/accessible/root")
	name, root, err := decodeEmbedBody(encoded)
	if err != nil {
		t.Fatalf("decodeEmbedBody failed: %v", err)
	}
	if name != ":1.42" {
		t.Fatalf("name = %q, want ':1.42'", name)
	}
	if root != "/org/a11y/atspi/accessible/root" {
		t.Fatalf("root = %q, want '/org/a11y/atspi/accessible/root'", root)
	}
}

func TestMessageRoundtrip(t *testing.T) {
	t.Parallel()

	orig := &dbusMsg{
		msgType:     dbusTypeMethodCall,
		flags:       dbusFlagNoAutoStart,
		serial:      42,
		replySerial: 10,
		path:        "/org/freedesktop/DBus",
		iface:       "org.freedesktop.DBus",
		member:      "Hello",
		dest:        "org.freedesktop.DBus",
		sender:      ":1.0",
		sig:         "s",
		body:        encodeStringBody("payload"),
	}

	wire := encodeMessage(orig)
	decoded, err := readMessage(bytes.NewReader(wire))
	if err != nil {
		t.Fatalf("readMessage failed: %v", err)
	}

	if decoded.msgType != orig.msgType || decoded.flags != orig.flags || decoded.serial != orig.serial {
		t.Fatalf("header mismatch: got %+v, want %+v", decoded, orig)
	}
	if decoded.replySerial != orig.replySerial {
		t.Fatalf("replySerial: got %d, want %d", decoded.replySerial, orig.replySerial)
	}
	if decoded.path != orig.path || decoded.iface != orig.iface || decoded.member != orig.member {
		t.Fatalf("fields mismatch: got %+v, want %+v", decoded, orig)
	}
	if decoded.dest != orig.dest || decoded.sender != orig.sender || decoded.sig != orig.sig {
		t.Fatalf("fields mismatch: got %+v, want %+v", decoded, orig)
	}
	bodyStr, err := decodeStringBody(decoded.body)
	if err != nil || bodyStr != "payload" {
		t.Fatalf("body mismatch: got %q, err %v", bodyStr, err)
	}
}

func encodeBigEndianMethodCall(path, iface, member, sig string, body []byte) []byte {
	var fields bytes.Buffer
	writeBigEndianStringField(&fields, dbusHeaderPath, "o", path)
	writeBigEndianStringField(&fields, dbusHeaderInterface, "s", iface)
	writeBigEndianStringField(&fields, dbusHeaderMember, "s", member)
	if sig != "" {
		writeBigEndianSignatureField(&fields, sig)
	}
	fieldsBytes := fields.Bytes()

	var msg bytes.Buffer
	msg.WriteByte(dbusEndianBig)
	msg.WriteByte(dbusTypeMethodCall)
	msg.WriteByte(0)
	msg.WriteByte(dbusProtoVersion)

	var u32 [4]byte
	binary.BigEndian.PutUint32(u32[:], uint32(len(body))) //nolint:gosec // reason: test bodies are tiny.
	msg.Write(u32[:])
	binary.BigEndian.PutUint32(u32[:], 7)
	msg.Write(u32[:])
	binary.BigEndian.PutUint32(u32[:], uint32(len(fieldsBytes))) //nolint:gosec // reason: test headers are tiny.
	msg.Write(u32[:])

	msg.Write(fieldsBytes)
	pad(&msg, dbusAlign8)
	msg.Write(body)
	return msg.Bytes()
}

func writeBigEndianStringField(fields *bytes.Buffer, code byte, sig, val string) {
	pad(fields, dbusAlign8)
	fields.WriteByte(code)
	fields.WriteByte(byte(len(sig))) //nolint:gosec // reason: test signatures are one char.
	fields.WriteString(sig)
	fields.WriteByte(0)

	pad(fields, dbusAlign4)
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(val))) //nolint:gosec // reason: test values are tiny.
	fields.Write(lenBuf[:])
	fields.WriteString(val)
	fields.WriteByte(0)
}

func writeBigEndianSignatureField(fields *bytes.Buffer, sig string) {
	pad(fields, dbusAlign8)
	fields.WriteByte(dbusHeaderSignature)
	fields.WriteByte(1)
	fields.WriteString("g")
	fields.WriteByte(0)
	fields.WriteByte(byte(len(sig))) //nolint:gosec // reason: test signatures are one char.
	fields.WriteString(sig)
	fields.WriteByte(0)
}

func encodeBigEndianStringBody(s string) []byte {
	var body bytes.Buffer
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(s))) //nolint:gosec // reason: test values are tiny.
	body.Write(lenBuf[:])
	body.WriteString(s)
	body.WriteByte(0)
	return body.Bytes()
}
