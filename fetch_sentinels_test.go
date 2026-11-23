package ownframe_test

import (
	"context"
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestFetchSentinelsMatchWithErrorsIs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	if _, err := ownframe.XHR(ctx, "BREW", "http://example.com/", nil, nil); !errors.Is(err, ownframe.ErrMethod) {
		t.Fatalf("method: %v", err)
	}
	hdr := map[string]string{"Cookie": "a=b"}
	if _, err := ownframe.XHR(ctx, "GET", "http://example.com/", hdr, nil); !errors.Is(err, ownframe.ErrBlockedHeader) {
		t.Fatalf("header: %v", err)
	}
	if _, err := ownframe.Fetch(ctx, "http://u:p@example.com/"); !errors.Is(err, ownframe.ErrCredentialsInURL) {
		t.Fatalf("userinfo: %v", err)
	}
	for name, err := range map[string]error{
		"body": ownframe.ErrBodyTooLarge, "headers": ownframe.ErrHeadersTooLarge, "redirects": ownframe.ErrTooManyRedirects,
	} {
		if err == nil {
			t.Fatalf("%s sentinel is nil", name)
		}
	}
}
