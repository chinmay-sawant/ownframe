package tv

import (
	"errors"
	"strings"
)

// Exec runs one spec such as "button:HOME" or "ssap:audio/volumeUp".
func (c *Client) Exec(spec string) (string, error) {
	kind, name, _ := strings.Cut(spec, ":")

	switch kind {
	case "button":
		return name, c.Button(name)
	case "ssap":
		_, err := c.Request(name, nil)
		return name, err
	case "launch":
		return c.launch(name)
	case "hdmi":
		return c.hdmi(name)
	case "hub":
		return c.hub()
	case "app":
		return c.appButton(name)
	case "hotstar":
		return c.hotstar()
	case "move":
		return c.pointerMove(name)
	case "click":
		return "Click", c.writePointer("type:click\n\n")
	case "scroll":
		return c.pointerScroll(name)
	case "power":
		_, err := c.Request("system/turnOff", nil)
		if err != nil && c.Button("POWER") != nil {
			return "", err
		}
		return "Turning the TV off", nil
	case "ime":
		if name == "watch:" {
			return "Watching the TV keyboard", c.watchKeyboard()
		}
		return c.ime(name)
	case "pair":
		if c.Model != "" {
			return c.Model, nil
		}
		return "Paired", nil
	case "mute":
		return c.toggleMute()
	case "vol":
		return c.volume(name)
	case "ch":
		return c.channel(name)
	case "media":
		return c.media(name)
	case "screen":
		return c.screen(name)
	default:
		return "", errors.New("unknown control")
	}
}
