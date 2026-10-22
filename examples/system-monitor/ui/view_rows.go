package ui

// buildPanels renders the overview cards from the panel graphs.
func buildPanels(s *state) []Panel {
	out := make([]Panel, 0, len(panelOrder))

	for _, id := range panelOrder {
		p := s.panels[id]
		card := Panel{
			ID:    id,
			Name:  p.name,
			Value: unavailable,
			Bars:  make([]int, graphColsMax),
		}

		if p.have {
			card.Value = p.last.Text
			card.Sub = p.last.Sub
			card.Peak = peakText(p)
		}

		out = append(out, card)
	}

	return out
}

// peakText renders the graph's largest sample in the card's unit.
func peakText(p *panel) string {
	v, ok := p.graph.Max()
	if !ok {
		return ""
	}

	if p.id == "cpu" || p.id == "mem" {
		return "peak " + formatPercent(v, true)
	}

	return "peak " + formatRate(v, true)
}

// Row and detail name preview lengths, in runes.
const (
	rowNameLen    = 44
	detailNameLen = 60
)

// buildRows renders the visible page of the frozen snapshot.
func buildRows(rows []Process, sel string) []Row {
	out := make([]Row, 0, len(rows))

	for _, p := range rows {
		out = append(out, Row{
			ID:    p.ID,
			PID:   formatCount(p.PID),
			Name:  truncate(p.Name, rowNameLen),
			State: nonEmpty(p.State),
			CPU:   cpuText(p.CPU, p.CPUKnown),
			Mem:   formatBytes(float64(p.Mem), p.MemKnown),
			Sel:   p.ID == sel,
		})
	}

	return out
}

// buildSel renders the detail card.
func buildSel(s *state) SelView {
	v := SelView{Any: s.sel.id != ""}
	if !v.Any {
		return v
	}

	v.ID = s.sel.id
	v.Name = truncate(nonEmpty(s.sel.proc.Name), detailNameLen)
	v.PID = formatCount(s.sel.proc.PID)
	v.Status = detailStatus(s.sel)
	v.Loading = s.sel.trk.Loading
	v.Err = s.sel.trk.Err
	v.Fields = detailFields(s.sel)

	return v
}
