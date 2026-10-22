package clipboard

import "testing"

func TestClearEmptiesMemory(t *testing.T) {
	UseMemory(true)
	Write("secret")
	Clear()

	if got, err := Read(); err != nil || got != "" {
		t.Fatalf("Read() = %q, %v", got, err)
	}
}
