package player

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// testCtx allows loopback httptest servers past the SSRF block.
func testCtx() context.Context {
	return ownframe.AllowPrivate(context.Background())
}
