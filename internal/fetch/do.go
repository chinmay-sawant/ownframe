package fetch

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
)

// Do sends one request. An empty method is GET.
// A nil context returns an error whose text contains "nil context".
// ctx cancellation is passed to http.NewRequestWithContext.
func Do(ctx context.Context, method, rawURL string, header map[string]string, body []byte) (Response, error) {
	if ctx == nil {
		return Response{}, ErrNilContext
	}
	if err := checkURL(ctx, rawURL); err != nil {
		return Response{}, err
	}
	if err := checkMethodHeader(method, header); err != nil {
		return Response{}, err
	}
	if method == "" {
		method = http.MethodGet
	}

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, reader)
	if err != nil {
		return Response{}, err
	}
	for key, value := range header {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, ErrCredentialsInURL) {
			return Response{}, ErrCredentialsInURL
		}
		if errors.Is(err, ErrScheme) {
			return Response{}, ErrScheme
		}
		if errors.Is(err, ErrTooManyRedirects) {
			return Response{}, ErrTooManyRedirects
		}
		if strings.Contains(err.Error(), "response headers exceeded") {
			return Response{}, ErrHeadersTooLarge
		}

		return Response{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	out, err := io.ReadAll(io.LimitReader(resp.Body, MaxBodyBytes+1))
	if err != nil {
		return Response{}, err
	}
	if len(out) > MaxBodyBytes {
		return Response{}, ErrBodyTooLarge
	}

	headerBytes := 0
	for key, values := range resp.Header {
		headerBytes += len(key)
		for _, value := range values {
			headerBytes += len(value)
		}
	}
	if headerBytes > MaxHeaderBytes {
		return Response{}, ErrHeadersTooLarge
	}

	return Response{Status: resp.StatusCode, Header: flat(resp.Header), Body: out}, nil
}
