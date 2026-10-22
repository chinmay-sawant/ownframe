package tv

import (
	"io"
	"net/http"
	"strings"
)

func nameFrom(text string) string {
	loc := headerValue(text, "location")
	if loc == "" {
		return ""
	}

	return nameAt(loc)
}

func nameAt(rawURL string) string {
	client := http.Client{Timeout: nameTimeout}
	resp, err := client.Get(rawURL)
	if err != nil {
		return ""
	}

	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, nameReadLimit))
	if err != nil {
		return ""
	}

	return tagText(string(body), "friendlyName")
}

func headerValue(text, name string) string {
	for _, line := range strings.Split(text, "\n") {
		key, val, ok := strings.Cut(line, ":")
		if !ok || !strings.EqualFold(strings.TrimSpace(key), name) {
			continue
		}

		return strings.TrimSpace(val)
	}

	return ""
}

func tagText(body, tag string) string {
	low := strings.ToLower(body)
	open := "<" + strings.ToLower(tag) + ">"
	close := "</" + strings.ToLower(tag) + ">"
	start := strings.Index(low, open)
	if start < 0 {
		return ""
	}

	start += len(open)
	end := strings.Index(low[start:], close)
	if end < 0 {
		return ""
	}

	return strings.TrimSpace(body[start : start+end])
}
