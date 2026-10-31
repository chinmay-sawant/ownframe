package window

import (
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/leakcheck"
	"github.com/hajimehoshi/ebiten/v2"
)

// TestMain fails the binary when window tests leak goroutines (GEN-26).
func TestMain(m *testing.M) {
	// Desktop tile tests use retained surfaces; GPU sessions can select clearing explicitly.
	ebiten.SetScreenClearedEveryFrame(false)
	// Unit tests must not claim or query the developer's real accessibility bus.
	_ = os.Setenv("AT_SPI_BUS_ADDRESS", "disabled")
	os.Exit(leakcheck.Verify(m))
}
