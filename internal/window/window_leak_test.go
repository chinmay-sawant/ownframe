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
	os.Exit(leakcheck.Verify(m))
}
