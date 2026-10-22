package wire

import (
	"context"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/ui"
)

// command kinds the worker loop runs off the UI loop.
type cmdKind uint8

const (
	cmdAdd cmdKind = iota
	cmdControl
	cmdActive
	cmdPage
	cmdSummary
	cmdDark
)

// command is one UI request for the worker.
type command struct {
	kind cmdKind
	add  ui.AddRequest
	ctl  ui.Control
	page ui.PageRequest
	gen  uint64
	dark bool
}

// Start launches the worker loop and the engine's worker pool. Commands
// queued before it wait in the channel.
func (b *Backend) Start(ctx context.Context) {
	runCtx, cancel := context.WithCancel(ctx)
	b.ctx, b.cancel = runCtx, cancel
	b.eng.Start(runCtx)
	go b.loop()
}

// loop runs one command at a time, so the store sees one writer.
func (b *Backend) loop() {
	defer close(b.done)

	for {
		select {
		case <-b.ctx.Done():
			return
		case c := <-b.cmds:
			b.run(c)
		}
	}
}
