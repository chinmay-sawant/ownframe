package chadcn

import (
	"fmt"
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// InputOTPOptions describes a controlled one-time-password control.
type InputOTPOptions struct {
	// ID is the group's stable host ID. Each slot derives its own ID from it.
	ID string
	// Label is the group's accessible name. It defaults to "One-time password".
	Label string
	// Length is the number of slots. Values at or below zero use six slots.
	Length int
	// Value fills the slots from left to right.
	Value string
	// Disabled marks every slot as non-interactive.
	Disabled bool
}

const defaultOTPLength = 6

// InputOTP renders one input per digit in a shared bordered row. The host
// reads the typed digits through Page.Form using the derived slot IDs.
func InputOTP(options InputOTPOptions) *ui.Element {
	length := options.Length
	if length <= 0 {
		length = defaultOTPLength
	}
	label := options.Label
	if label == "" {
		label = "One-time password"
	}
	digits := []rune(options.Value)
	group := ui.Div().Class("of-input-otp").Attr("role", "group").Attr("aria-label", label)
	if options.ID != "" {
		group.ID(options.ID)
	}
	if options.Disabled {
		group.Attr("aria-disabled", ariaBool(true))
	}
	for index := range length {
		group.Children(otpSlot(options, digits, index, length))
	}
	return group
}

func otpSlot(options InputOTPOptions, digits []rune, index, length int) *ui.Element {
	value := ""
	if index < len(digits) {
		value = string(digits[index])
	}
	slot := ui.Input().Class("of-input-otp-slot").Attr("type", "text").
		Attr("inputmode", "numeric").Attr("maxlength", "1").
		Attr("value", value).
		Attr("aria-label", fmt.Sprintf("Digit %d of %d", index+1, length))
	if index == 0 {
		slot.Attr("autocomplete", "one-time-code")
	} else {
		slot.Attr("autocomplete", "off")
	}
	if options.ID != "" {
		slot.ID(partID(options.ID, "-"+strconv.Itoa(index)))
	}
	return setDisabled(slot, options.Disabled)
}
