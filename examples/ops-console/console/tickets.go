package console

func seedTickets() []Ticket {
	return []Ticket{
		{
			ID:     "t1",
			Title:  "Seal mismatch",
			Bay:    "Bay 4",
			Status: statusOpen,
			Note:   "Trailer 441 arrived with a broken seal.",
			Action: "open-t1",
		},
		{
			ID:     "t2",
			Title:  "Gate hold",
			Bay:    "Lane C",
			Status: statusHold,
			Note:   "Paperwork missing for truck 88.",
			Action: "open-t2",
		},
		{
			ID:     "t3",
			Title:  "Split pallet",
			Bay:    "Bay 12",
			Status: statusOpen,
			Note:   "Two crates off the stack at inbound.",
			Action: "open-t3",
		},
	}
}

func (a *App) ticket(id string) (Ticket, bool) {
	for _, t := range a.tickets {
		if t.ID == id {
			return t, true
		}
	}

	return Ticket{}, false
}

func (a *App) openTicket(id string) {
	t, ok := a.ticket(id)
	if !ok {
		return
	}

	a.view.Current = t
	a.view.Note = t.Note
	a.page.SetFormValue("note", t.Note)
	a.show(screenDetail)
}
