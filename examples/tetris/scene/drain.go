package scene

import "context"

// drainBudget bounds how many worker results one frame applies, so a
// burst cannot stall the tick.
const drainBudget = 4

// drain applies finished store work. A result whose id is not the current
// request for its kind is stale and is dropped.
func (s *Scene) drain(ctx context.Context) error {
	if s.store == nil {
		return nil
	}

	for i := 0; i < drainBudget; i++ {
		res, ok := s.store.Poll()
		if !ok {
			return nil
		}

		if err := s.apply(ctx, res); err != nil {
			return err
		}
	}

	return nil
}

// apply handles one finished request.
func (s *Scene) apply(ctx context.Context, res Result) error {
	switch res.Kind {
	case ResultSettings:
		if res.ID == s.settingsReq {
			s.applySettings(ctx, res)
		}

	case ResultTheme:
		if res.ID != s.themeReq {
			return nil
		}

		if res.Err != nil {
			s.setStatus("THEME NOT SAVED")

			return nil
		}

		s.setStatus("THEME SAVED")

	case ResultScores, ResultDummy:
		if err := s.applyScores(ctx, res); err != nil {
			return err
		}

	case ResultSaved:
		s.applySaved(res)

	case ResultSnapshot:
		if res.ID == s.snapReq && res.Err != nil {
			s.setStatus("SNAPSHOT FAILED")
		}
	}

	return nil
}
