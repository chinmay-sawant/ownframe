package atspi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	dbusEndianLittle byte = 'l'

	dbusTypeMethodCall   byte = 1
	dbusTypeMethodReturn byte = 2
	dbusTypeError        byte = 3
	dbusTypeSignal       byte = 4

	dbusFlagNoReplyExpected byte = 0x01
	dbusFlagNoAutoStart     byte = 0x02

	dbusHeaderPath        byte = 1
	dbusHeaderInterface   byte = 2
	dbusHeaderMember      byte = 3
	dbusHeaderErrorName   byte = 4
	dbusHeaderReplySerial byte = 5
	dbusHeaderDestination byte = 6
	dbusHeaderSender      byte = 7
	dbusHeaderSignature   byte = 8

	dbusProtoVersion byte = 1
)

type dbusMsg struct {
	msgType     byte
	flags       byte
	serial      uint32
	replySerial uint32
	path        string
	iface       string
	member      string
	errorName   string
	dest        string
	sender      string
	sig         string
	body        []byte
}

func pad(b *bytes.Buffer, alignment int) {
	rem := b.Len() % alignment
	if rem != 0 {
		padLen := alignment - rem
		for i := 0; i < padLen; i++ {
			b.WriteByte(0)
		}
	}
}

func writeHeaderField(buf *bytes.Buffer, code byte, sig string, writeVal func(*bytes.Buffer)) {
	pad(buf, 8)
	buf.WriteByte(code)
	buf.WriteByte(byte(len(sig)))
	buf.WriteString(sig)
	buf.WriteByte(0)
	writeVal(buf)
}

func writeStringField(buf *bytes.Buffer, code byte, sig string, val string) {
	if val == "" {
		return
	}
	writeHeaderField(buf, code, sig, func(b *bytes.Buffer) {
		pad(b, 4)
		var lenBuf [4]byte
		binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(val)))
		b.Write(lenBuf[:])
		b.WriteString(val)
		b.WriteByte(0)
	})
}

func writeUint32Field(buf *bytes.Buffer, code byte, val uint32) {
	writeHeaderField(buf, code, "u", func(b *bytes.Buffer) {
		pad(b, 4)
		var valBuf [4]byte
		binary.LittleEndian.PutUint32(valBuf[:], val)
		b.Write(valBuf[:])
	})
}

func writeSignatureField(buf *bytes.Buffer, code byte, sigVal string) {
	if sigVal == "" {
		return
	}
	writeHeaderField(buf, code, "g", func(b *bytes.Buffer) {
		b.WriteByte(byte(len(sigVal)))
		b.WriteString(sigVal)
		b.WriteByte(0)
	})
}

func encodeMessage(msg *dbusMsg) []byte {
	var headerFields bytes.Buffer
	writeStringField(&headerFields, dbusHeaderPath, "o", msg.path)
	writeStringField(&headerFields, dbusHeaderInterface, "s", msg.iface)
	writeStringField(&headerFields, dbusHeaderMember, "s", msg.member)
	writeStringField(&headerFields, dbusHeaderErrorName, "s", msg.errorName)
	if msg.replySerial != 0 {
		writeUint32Field(&headerFields, dbusHeaderReplySerial, msg.replySerial)
	}
	writeStringField(&headerFields, dbusHeaderDestination, "s", msg.dest)
	writeStringField(&headerFields, dbusHeaderSender, "s", msg.sender)
	writeSignatureField(&headerFields, dbusHeaderSignature, msg.sig)

	fieldsBytes := headerFields.Bytes()
	fieldsLen := uint32(len(fieldsBytes))
	bodyLen := uint32(len(msg.body))

	var msgBuf bytes.Buffer
	msgBuf.WriteByte(dbusEndianLittle)
	msgBuf.WriteByte(msg.msgType)
	msgBuf.WriteByte(msg.flags)
	msgBuf.WriteByte(dbusProtoVersion)

	var u32 [4]byte
	binary.LittleEndian.PutUint32(u32[:], bodyLen)
	msgBuf.Write(u32[:])

	binary.LittleEndian.PutUint32(u32[:], msg.serial)
	msgBuf.Write(u32[:])

	binary.LittleEndian.PutUint32(u32[:], fieldsLen)
	msgBuf.Write(u32[:])

	msgBuf.Write(fieldsBytes)
	pad(&msgBuf, 8)

	if len(msg.body) > 0 {
		msgBuf.Write(msg.body)
	}

	return msgBuf.Bytes()
}

func readMessage(r io.Reader) (*dbusMsg, error) {
	var fixed [16]byte
	if _, err := io.ReadFull(r, fixed[:]); err != nil {
		return nil, err
	}
	endian := fixed[0]
	if endian != dbusEndianLittle {
		return nil, fmt.Errorf("unsupported dbus endianness: %c", endian)
	}
	msg := &dbusMsg{
		msgType: fixed[1],
		flags:   fixed[2],
	}
	bodyLen := binary.LittleEndian.Uint32(fixed[4:8])
	msg.serial = binary.LittleEndian.Uint32(fixed[8:12])
	fieldsLen := binary.LittleEndian.Uint32(fixed[12:16])

	fieldsBytes := make([]byte, fieldsLen)
	if fieldsLen > 0 {
		if _, err := io.ReadFull(r, fieldsBytes); err != nil {
			return nil, err
		}
	}

	totalHeader := 16 + fieldsLen
	padLen := (8 - (totalHeader % 8)) % 8
	if padLen > 0 {
		var padBuf [8]byte
		if _, err := io.ReadFull(r, padBuf[:padLen]); err != nil {
			return nil, err
		}
	}

	if bodyLen > 0 {
		msg.body = make([]byte, bodyLen)
		if _, err := io.ReadFull(r, msg.body); err != nil {
			return nil, err
		}
	}

	offset := 0
	for offset < len(fieldsBytes) {
		pad := (8 - ((16 + offset) % 8)) % 8
		offset += pad
		if offset >= len(fieldsBytes) {
			break
		}
		code := fieldsBytes[offset]
		offset++
		if offset >= len(fieldsBytes) {
			break
		}
		sigLen := int(fieldsBytes[offset])
		offset++
		if offset+sigLen >= len(fieldsBytes) {
			break
		}
		sig := string(fieldsBytes[offset : offset+sigLen])
		offset += sigLen + 1

		switch sig {
		case "s", "o":
			pad := (4 - ((16 + offset) % 4)) % 4
			offset += pad
			if offset+4 > len(fieldsBytes) {
				break
			}
			strLen := int(binary.LittleEndian.Uint32(fieldsBytes[offset : offset+4]))
			offset += 4
			if offset+strLen > len(fieldsBytes) {
				break
			}
			val := string(fieldsBytes[offset : offset+strLen])
			offset += strLen + 1

			switch code {
			case dbusHeaderPath:
				msg.path = val
			case dbusHeaderInterface:
				msg.iface = val
			case dbusHeaderMember:
				msg.member = val
			case dbusHeaderErrorName:
				msg.errorName = val
			case dbusHeaderDestination:
				msg.dest = val
			case dbusHeaderSender:
				msg.sender = val
			}
		case "u":
			pad := (4 - ((16 + offset) % 4)) % 4
			offset += pad
			if offset+4 > len(fieldsBytes) {
				break
			}
			uVal := binary.LittleEndian.Uint32(fieldsBytes[offset : offset+4])
			offset += 4
			if code == dbusHeaderReplySerial {
				msg.replySerial = uVal
			}
		case "g":
			if offset >= len(fieldsBytes) {
				break
			}
			gLen := int(fieldsBytes[offset])
			offset++
			if offset+gLen > len(fieldsBytes) {
				break
			}
			gVal := string(fieldsBytes[offset : offset+gLen])
			offset += gLen + 1
			if code == dbusHeaderSignature {
				msg.sig = gVal
			}
		default:
			offset = len(fieldsBytes)
		}
	}

	return msg, nil
}

func encodeStringBody(s string) []byte {
	var b bytes.Buffer
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(s)))
	b.Write(lenBuf[:])
	b.WriteString(s)
	b.WriteByte(0)
	return b.Bytes()
}

func encodeUint32Body(val uint32) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], val)
	return b[:]
}

func encodeRequestNameBody(name string, flags uint32) []byte {
	var b bytes.Buffer
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(name)))
	b.Write(lenBuf[:])
	b.WriteString(name)
	b.WriteByte(0)

	pad(&b, 4)
	binary.LittleEndian.PutUint32(lenBuf[:], flags)
	b.Write(lenBuf[:])
	return b.Bytes()
}

func decodeRequestNameBody(body []byte) (string, uint32, error) {
	if len(body) < 4 {
		return "", 0, errors.New("short body")
	}
	strLen := int(binary.LittleEndian.Uint32(body[0:4]))
	if len(body) < 4+strLen+1 {
		return "", 0, errors.New("short string body")
	}
	name := string(body[4 : 4+strLen])
	offset := 4 + strLen + 1
	rem := offset % 4
	if rem != 0 {
		offset += 4 - rem
	}
	if len(body) < offset+4 {
		return name, 0, nil
	}
	flags := binary.LittleEndian.Uint32(body[offset : offset+4])
	return name, flags, nil
}

func decodeStringBody(body []byte) (string, error) {
	if len(body) < 4 {
		return "", errors.New("short string body")
	}
	strLen := int(binary.LittleEndian.Uint32(body[0:4]))
	if len(body) < 4+strLen {
		return "", errors.New("short string body data")
	}
	return string(body[4 : 4+strLen]), nil
}

func decodeUint32Body(body []byte) (uint32, error) {
	if len(body) < 4 {
		return 0, errors.New("short uint32 body")
	}
	return binary.LittleEndian.Uint32(body[0:4]), nil
}

func encodeEmbedBody(name, rootPath string) []byte {
	var b bytes.Buffer
	var lenBuf [4]byte
	binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(name)))
	b.Write(lenBuf[:])
	b.WriteString(name)
	b.WriteByte(0)

	pad(&b, 4)
	binary.LittleEndian.PutUint32(lenBuf[:], uint32(len(rootPath)))
	b.Write(lenBuf[:])
	b.WriteString(rootPath)
	b.WriteByte(0)
	return b.Bytes()
}

func decodeEmbedBody(body []byte) (string, string, error) {
	if len(body) < 4 {
		return "", "", errors.New("short embed body")
	}
	nameLen := int(binary.LittleEndian.Uint32(body[0:4]))
	if len(body) < 4+nameLen+1 {
		return "", "", errors.New("short embed name data")
	}
	name := string(body[4 : 4+nameLen])
	offset := 4 + nameLen + 1
	rem := offset % 4
	if rem != 0 {
		offset += 4 - rem
	}
	if len(body) < offset+4 {
		return name, "", errors.New("short embed path len")
	}
	pathLen := int(binary.LittleEndian.Uint32(body[offset : offset+4]))
	offset += 4
	if len(body) < offset+pathLen {
		return name, "", errors.New("short embed path data")
	}
	rootPath := string(body[offset : offset+pathLen])
	return name, rootPath, nil
}
