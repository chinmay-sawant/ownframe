package fetch

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
)

func TestBlockedAddrSpecialRanges(t *testing.T) {
	t.Parallel()

	for _, s := range []string{
		"100.64.0.1", "100.127.255.254", "198.18.0.1", "198.19.255.255", "240.0.0.1",
		"255.255.255.255", "0.0.0.1", "64:ff9b::169.254.169.254", "64:ff9b::a9fe:a9fe",
		"fe80::1%eth0", "::ffff:127.0.0.1", "::ffff:10.0.0.1", "2002:a9fe:a9fe::1",
		"2001:db8::1", "fd00::1", "::1", "ff02::1",
	} {
		addr, err := netip.ParseAddr(s)
		if err != nil {
			t.Fatalf("parse %q: %v", s, err)
		}
		if !blockedAddr(addr) {
			t.Errorf("blockedAddr(%s) = false", s)
		}
	}
	for _, s := range []string{"8.8.8.8", "1.1.1.1", "2606:4700:4700::1111", "::ffff:8.8.8.8"} {
		addr := netip.MustParseAddr(s)
		if blockedAddr(addr) {
			t.Errorf("blockedAddr(%s) = true", s)
		}
	}
}

func TestAllowPrivateNilContext(t *testing.T) {
	t.Parallel()

	ctx := AllowPrivate(nil) //nolint:staticcheck // reason: the nil context is the case under test.
	if ctx != nil {
		t.Fatal("AllowPrivate(nil) must stay nil")
	}
	if _, err := Do(ctx, "", "http://example.com", nil, nil); !errors.Is(err, ErrNilContext) {
		t.Fatalf("err = %v, want ErrNilContext", err)
	}
}

func TestClientForSeparatesPrivatePool(t *testing.T) {
	t.Parallel()

	if clientFor(context.Background()) == clientFor(AllowPrivate(context.Background())) {
		t.Fatal("AllowPrivate shares the normal client's connection pool")
	}
	if client.Transport == privateClient.Transport {
		t.Fatal("clients share a transport")
	}
}

func TestBlockedLiteralZonedAndCGNAT(t *testing.T) {
	t.Parallel()

	for _, u := range []string{"http://100.64.0.1/", "http://[fe80::1%25eth0]/", "http://[64:ff9b::a9fe:a9fe]/"} {
		if _, err := Do(context.Background(), "", u, nil, nil); !errors.Is(err, ErrPrivateAddress) {
			t.Errorf("Do(%q) = %v, want ErrPrivateAddress", u, err)
		}
	}
}

func TestPrivateAndPublicCallsUseSeparateConnections(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) }))
	defer srv.Close()

	if _, err := Do(AllowPrivate(context.Background()), "", srv.URL, nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := Do(context.Background(), "", srv.URL, nil, nil); !errors.Is(err, ErrPrivateAddress) {
		t.Fatalf("normal call reached a loopback host: %v", err)
	}
}
