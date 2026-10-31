package telegram_test

import (
	"os"
	"testing"
)

// TestMain keeps the package tests off the developer's real accessibility
// bus. The phone tests drive full shell updates, and a live bus would claim
// a name and publish a test window during `make test`.
func TestMain(m *testing.M) {
	_ = os.Setenv("AT_SPI_BUS_ADDRESS", "disabled")
	os.Exit(m.Run())
}
