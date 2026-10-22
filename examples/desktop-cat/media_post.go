//go:build !js && !android && !ios

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/chinmay-sawant/ownframe/examples/desktop-cat/cat"
)

func postMedia(ctx context.Context, endpoint string, n cat.Notification) error {
	ctx, cancel := context.WithTimeout(ctx, postTimeout)
	defer cancel()
	data, err := json.Marshal(n)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	if u, err := url.Parse(endpoint); err == nil && u.Scheme != "" && u.Host != "" {
		request.Header.Set("Origin", u.Scheme+"://"+u.Host)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusAccepted {
		return fmt.Errorf("media notification returned %s", response.Status)
	}
	return nil
}
