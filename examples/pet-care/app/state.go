package app

import (
	"sync/atomic"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/store"
)

type App struct {
	page    *ownframe.Page
	db      *store.Store
	view    View
	pickers chan string
	exports chan []byte
	imports chan []byte
	dark    atomic.Bool
	back    atomic.Bool
	canBack atomic.Bool
}
