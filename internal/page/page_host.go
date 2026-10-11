package page

import "github.com/chinmay-sawant/ownframe/internal/host"

var (
	_ host.Focuser         = (*Page)(nil)
	_ host.ContextMenu     = (*Page)(nil)
	_ host.CursorShape     = (*Page)(nil)
	_ host.ScrollRequester = (*Page)(nil)
	_ host.LongPresser     = (*Page)(nil)
	_ host.Swiper          = (*Page)(nil)
	_ host.Accessible      = (*Page)(nil)
)
