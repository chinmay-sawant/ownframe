//go:build ios

package mobile

import "github.com/chinmay-sawant/ownframe/examples/pet-care/store"

func init() {
	path, err := store.DefaultPath()
	if err == nil {
		err = start(path)
	}
	if err != nil {
		panic(err)
	}
}
