// Package mobile binds the Depth Alert page to Android and iOS builds.
package mobile

func Dummy() {}

func init() {
	if err := start(); err != nil {
		panic(err)
	}
}
