// Package mobile is the Android and iOS entry for the eyes demo.
package mobile

// Dummy exists so ebitenmobile bind will compile this package.
func Dummy() {}

func init() {
	if err := start(); err != nil {
		panic(err)
	}
}
