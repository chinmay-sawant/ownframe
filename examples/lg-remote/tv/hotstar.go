package tv

import (
	"errors"
	"strings"
)

func (c *Client) hotstar() (string, error) {
	if c.hotstarID != "" {
		if msg, err := c.launch(c.hotstarID); err == nil {
			return msg, nil
		}
		c.hotstarID = ""
	}
	p, err := c.Request("com.webos.applicationManager/listApps", nil)
	if err != nil {
		return "", err
	}
	id := hotstarID(p)
	if id == "" {
		return "", errors.New("hotstar is not installed on this TV")
	}
	msg, err := c.launch(id)
	if err == nil {
		c.hotstarID = id
	}
	return msg, err
}

func hotstarID(p map[string]any) string {
	apps, _ := p["apps"].([]any)
	for _, value := range apps {
		app, ok := value.(map[string]any)
		if !ok {
			continue
		}
		id := textField(app, "id")
		name := strings.ToLower(id + " " + textField(app, "title"))
		if id != "" && strings.Contains(name, "hotstar") {
			return id
		}
	}
	return ""
}
