package atspi

import (
	"context"
	"slices"
	"unicode/utf8"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

const atspiEditableText = "org.a11y.atspi.EditableText"

func editableTextCapable(obj Object) bool {
	return obj.Role == RoleEntry && obj.State.Enabled && obj.State.Sensitive &&
		obj.State.Editable && !obj.State.ReadOnly && !obj.Password &&
		slices.Contains(obj.Actions, ActionSetText)
}

func (v *serverView) handleEditableText(ctx context.Context, msg *dbusMsg) callResult {
	obj, ok := v.object(msg.path)
	if !ok {
		return errResult("org.freedesktop.DBus.Error.UnknownObject", "unknown object "+msg.path)
	}
	if !editableTextCapable(obj) || v.actions == nil {
		return errResult("org.freedesktop.DBus.Error.UnknownInterface", "object cannot edit text")
	}
	edit, value, failure := decodeTextEdit(msg)
	if failure.errName != "" {
		return failure
	}
	err := v.actions(ctx, host.AccessibilityRequest{
		Document: v.app.Document,
		Node:     obj.ID,
		Kind:     host.AccessibilitySetValue,
		Value:    value,
		TextEdit: edit,
	})
	if msg.member == "CopyText" {
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.Failed", err.Error())
		}
		return emptyResult()
	}
	return okResult("b", encodeBoolValue(err == nil))
}

func decodeTextEdit(msg *dbusMsg) (*host.AccessibilityTextEdit, string, callResult) {
	d := dbusDecoder{body: msg.body}
	invalid := errResult("org.freedesktop.DBus.Error.InvalidArgs", "invalid editable text arguments")
	switch msg.member {
	case "SetTextContents":
		text, err := d.stringVal()
		if err != nil || !utf8.ValidString(text) {
			return nil, "", invalid
		}
		return nil, text, callResult{}
	case "InsertText", "CopyText", "CutText", "DeleteText", "PasteText":
	default:
		return nil, "", errResult("org.freedesktop.DBus.Error.UnknownMethod", "unknown method "+msg.member)
	}

	start, err := d.int32Val()
	if err != nil || start < 0 {
		return nil, "", invalid
	}
	edit := &host.AccessibilityTextEdit{Start: int(start), End: int(start)}
	switch msg.member {
	case "InsertText":
		text, textErr := d.stringVal()
		length, lengthErr := d.int32Val()
		if textErr != nil || lengthErr != nil || length < -1 || !utf8.ValidString(text) {
			return nil, "", invalid
		}
		if length >= 0 {
			text = truncateRunes(text, int(length))
		}
		edit.Kind = host.AccessibilityReplaceText
		return edit, text, callResult{}
	case "PasteText":
		edit.Kind = host.AccessibilityPasteText
	default:
		end, err := d.int32Val()
		if err != nil || end < start {
			return nil, "", invalid
		}
		edit.End = int(end)
		switch msg.member {
		case "CopyText":
			edit.Kind = host.AccessibilityCopyText
		case "CutText":
			edit.Kind = host.AccessibilityCutText
		case "DeleteText":
			edit.Kind = host.AccessibilityReplaceText
		}
	}
	return edit, "", callResult{}
}

// truncateRunes returns at most maxRunes UTF-8 runes from s without splitting
// any multi-byte character.
func truncateRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	count := 0
	for i := range s {
		if count == maxRunes {
			return s[:i]
		}
		count++
	}
	return s
}

const editableTextXML = `<interface name="org.a11y.atspi.EditableText">
<property name="version" type="u" access="read"/>
<method name="SetTextContents"><arg type="s" direction="in"/><arg type="b" direction="out"/></method>
<method name="InsertText"><arg type="i" direction="in"/><arg type="s" direction="in"/><arg type="i" direction="in"/><arg type="b" direction="out"/></method>
<method name="CopyText"><arg type="i" direction="in"/><arg type="i" direction="in"/></method>
<method name="CutText"><arg type="i" direction="in"/><arg type="i" direction="in"/><arg type="b" direction="out"/></method>
<method name="DeleteText"><arg type="i" direction="in"/><arg type="i" direction="in"/><arg type="b" direction="out"/></method>
<method name="PasteText"><arg type="i" direction="in"/><arg type="b" direction="out"/></method>
</interface>`
