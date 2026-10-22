package console

func (a *App) syncView() {
	a.view.Tickets = a.filtered()
	a.view.Empty = a.view.Screen == screenList && len(a.view.Tickets) == 0
	a.view.OpenN, a.view.HoldN, a.view.ClosedN = a.counts()
	a.view.Nav = a.navItems()

	if t, ok := a.ticket(a.view.Current.ID); ok {
		a.view.Current = t
	}
}

func (a *App) filtered() []Ticket {
	out := make([]Ticket, 0, len(a.tickets))
	for _, t := range a.tickets {
		if a.view.Filter == filterAll || a.view.Filter == t.Status {
			out = append(out, t)
		}
	}

	return out
}

func (a *App) counts() (openN, holdN, closedN int) {
	for _, t := range a.tickets {
		switch t.Status {
		case statusOpen:
			openN++
		case statusHold:
			holdN++
		case statusClosed:
			closedN++
		}
	}

	return openN, holdN, closedN
}

func (a *App) setFilter(name string) {
	switch name {
	case filterAll, statusOpen, statusHold, statusClosed:
		a.view.Filter = name
	}
}

func (a *App) setStatus(status string) {
	switch status {
	case statusOpen, statusHold, statusClosed:
	default:
		return
	}

	for i := range a.tickets {
		if a.tickets[i].ID == a.view.Current.ID {
			a.tickets[i].Status = status
			a.view.Notice = "Status updated."

			return
		}
	}
}

func (a *App) saveNote() {
	for i := range a.tickets {
		if a.tickets[i].ID == a.view.Current.ID {
			a.tickets[i].Note = a.view.Note
			a.view.Notice = "Note saved."

			return
		}
	}
}
