package app

import (
	"github.com/chinmay-sawant/ownframe/examples/pet-care/model"
	"strings"
)

func recordRows(records []model.Record, pets []model.Pet) []RecordRow {
	rows := make([]RecordRow, 0, len(records))
	for _, r := range records {
		name := "Unknown pet"
		for _, p := range pets {
			if p.ID == r.PetID {
				name = p.Name
				break
			}
		}
		status := "Recorded"
		if r.Scheduled && !r.Complete {
			status = "Upcoming"
			if r.Date == model.Today() {
				status = "Due today"
			}
			if r.Date < model.Today() {
				status = "Overdue"
			}
		}
		if r.Complete {
			status = "Complete"
		}
		rows = append(rows, RecordRow{Record: r, PetName: name, Status: status})
	}
	return rows
}

func visibleRecords(rows []RecordRow, screen, petID, query string) []RecordRow {
	out := make([]RecordRow, 0, len(rows))
	for _, r := range rows {
		if screen == "schedule" && !r.Scheduled {
			continue
		}
		if screen == "dashboard" && (!r.Scheduled || r.Complete) {
			continue
		}
		if screen == "profile" && petID != "" && r.PetID != petID {
			continue
		}
		if screen == "history" && query != "" && !strings.Contains(strings.ToLower(r.Category+" "+r.PetName), strings.ToLower(query)) {
			continue
		}
		out = append(out, r)
	}
	return out
}
