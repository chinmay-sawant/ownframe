package depthalert

import "context"

func (a *App) Tick(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a.page.Probe().UseFrameDirty()
	a.mu.Lock()
	current := a.view
	changed := current != a.rendered
	a.mu.Unlock()
	if a.ops.generation != a.page.Probe().Generation() {
		if err := a.ops.bind(a.page); err != nil {
			return err
		}
	}
	if changed {
		a.page.SetData(current)
		a.ops.update(a.page, current)
		a.mu.Lock()
		a.rendered = current
		a.mu.Unlock()
	}
	return nil
}
