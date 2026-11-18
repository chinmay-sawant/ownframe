package tv

// KeyboardState is what the TV last said about its on-screen keyboard.
type KeyboardState int

const (
	// KeyboardUnknown means the TV has not reported yet.
	KeyboardUnknown KeyboardState = iota
	// KeyboardClosed means no text box on the TV is waiting for text.
	KeyboardClosed
	// KeyboardOpen means a text box on the TV has focus and takes typing.
	KeyboardOpen
	// KeyboardUnreported means this TV does not publish keyboard focus.
	KeyboardUnreported
)

const imeWatchID = "ime_watch"

// watchKeyboard subscribes once per connection to the TV's remote keyboard
// events. The TV answers whenever a text box gains or loses focus.
func (c *Client) watchKeyboard() error {
	c.mu.Lock()
	if c.imeWatching {
		c.mu.Unlock()
		return nil
	}
	c.imeWatching = true
	c.mu.Unlock()

	return c.send(map[string]any{
		"id":      imeWatchID,
		"type":    "subscribe",
		"uri":     "ssap://" + imeService + "registerRemoteKeyboard",
		"payload": map[string]any{"subscribe": true},
	})
}

func (c *Client) noteKeyboard(m Message) {
	state := KeyboardClosed
	switch m.Type {
	case "error":
		state = KeyboardUnreported
	default:
		p := mapOf(m.Payload)
		if w, ok := p["currentWidget"].(map[string]any); ok {
			if focus, _ := w["focus"].(bool); focus {
				state = KeyboardOpen
			}
		}
	}

	c.mu.Lock()
	c.imeState = state
	c.mu.Unlock()
}

// KeyboardState reports the last keyboard focus the TV published. It never
// waits on the network.
func (s *Session) KeyboardState() KeyboardState {
	c := s.cli.Load()
	if c == nil {
		return KeyboardUnknown
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	return c.imeState
}
