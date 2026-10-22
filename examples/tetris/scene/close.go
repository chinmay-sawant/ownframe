package scene

import "time"

// ShutdownBudget bounds how long Close waits for the store worker.
const ShutdownBudget = 2 * time.Second

// snapshotWait bounds how long Close waits for the resume snapshot to
// reach the store before the worker is closed.
const snapshotWait = 500 * time.Millisecond

// Close stops the frame callback, saves a resume point, and joins the
// store worker within ShutdownBudget. A worker that outlives the budget
// is reported and left to the process exit.
func (s *Scene) Close() error {
	s.page.SetTick(nil)
	s.savePoint()
	s.awaitSnapshot()

	if s.store == nil {
		return nil
	}

	done := make(chan error, 1)

	go func() { done <- s.store.Close() }()

	select {
	case err := <-done:
		return err
	case <-time.After(ShutdownBudget):
		return ErrShutdownTimeout
	}
}

// awaitSnapshot polls briefly so the resume snapshot reaches the store
// before the worker is canceled.
func (s *Scene) awaitSnapshot() {
	if s.store == nil || s.snapReq == 0 {
		return
	}

	deadline := time.Now().Add(snapshotWait)

	for time.Now().Before(deadline) {
		res, ok := s.store.Poll()
		if !ok {
			time.Sleep(snapshotPoll)

			continue
		}

		if res.Kind == ResultSnapshot && res.ID == s.snapReq {
			return
		}
	}
}
