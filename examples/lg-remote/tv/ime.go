package tv

import (
	"errors"
	"strconv"
	"strings"
)

const imeService = "com.webos.service.ime/"

// clearCount is how many characters Clear deletes when the TV will not
// replace the field in one step.
const clearCount = 200

// ime sends text editing to the TV's focused input box. spec is
// "text:<chars>", "delete:<count>", or "enter:".
func (c *Client) ime(spec string) (string, error) {
	kind, arg, _ := strings.Cut(spec, ":")
	return c.imeSend(kind, arg)
}

func (c *Client) imeSend(kind, arg string) (string, error) {

	switch kind {
	case "text":
		if arg == "" {
			return "", errors.New("nothing to type")
		}
		_, err := c.Request(imeService+"insertText", map[string]any{"text": arg, "replace": 0})

		return "Typed on the TV", imeError(err)
	case "delete":
		n, err := strconv.Atoi(arg)
		if err != nil || n < 1 {
			n = 1
		}
		_, err = c.Request(imeService+"deleteCharacters", map[string]any{"count": n})

		return "Deleted on the TV", imeError(err)
	case "clear":
		// replace:true swaps the whole field for the empty string. Some TVs
		// refuse that, so fall back to deleting a long run of characters.
		if _, err := c.Request(imeService+"insertText", map[string]any{"text": "", "replace": true}); err == nil {
			return "Cleared on the TV", nil
		}
		_, err := c.Request(imeService+"deleteCharacters", map[string]any{"count": clearCount})

		return "Cleared on the TV", imeError(err)
	case "enter":
		_, err := c.Request(imeService+"sendEnterKey", nil)

		return "Enter sent to the TV", imeError(err)
	default:
		return "", errors.New("unknown keyboard command")
	}
}

// imeError says what to do when the TV refuses text. The usual cause is that
// no text box on the TV has its keyboard open.
func imeError(err error) error {
	if err == nil {
		return nil
	}

	text := err.Error()
	if strings.Contains(text, "did not answer") || strings.Contains(text, "closed the connection") {
		return err
	}
	if strings.Contains(strings.ToLower(text), "permission") {
		return errors.New("the TV refused typing for this app. Remove the app from the TV's paired devices and connect again")
	}

	return errors.New("the TV did not take the text (" + text + "). Open a search box on the TV first")
}
