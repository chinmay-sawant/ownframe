package ipc

import (
	"context"
	"errors"
	"strconv"
)

// onLog is the first demo.log listener: it keeps every payload it sees.
func (a *App) onLog(payload string) {
	if a.view.Received != "" {
		a.view.Received += " "
	}

	a.view.Received += payload
}

// onCount is the second demo.log listener. Send calls every listener, so this
// one counts the sends it saw while the first one keeps the payloads.
func (a *App) onCount(string) {
	a.view.Sends++
}

// onDouble is the demo.double handler: Request gets twice the number, or an
// error for text that is not a number.
func (a *App) onDouble(_ context.Context, payload string) (string, error) {
	n, err := strconv.Atoi(payload)
	if err != nil {
		return "", errors.New("not a number: " + payload)
	}

	return strconv.Itoa(n * doubleFactor), nil
}
