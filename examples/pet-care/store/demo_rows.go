package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/pet-care/model"
)

// demoDate offsets keep the seeded rows spread across this week.
const (
	groomingInDays = 6
	habitatInDays  = 2
)

func seedDemoRows(ctx context.Context, tx *sql.Tx) error {
	for _, p := range demoPets {
		_, err := tx.ExecContext(ctx, `INSERT INTO pets VALUES(?,?,?,?,?,?)`, p.ID, p.Name, p.Species, p.Breed, p.BirthDate, p.Notes)
		if err != nil {
			return err
		}
	}
	rows := []model.Record{
		{ID: "demo-care-1", PetID: "demo-maple", Category: "Vet check-in", Date: demoDate(0), Note: "Bring the vaccination record.", Scheduled: true},
		{ID: "demo-care-2", PetID: "demo-maple", Category: "Grooming", Date: demoDate(groomingInDays), Note: "Brush and trim appointment.", Scheduled: true},
		{ID: "demo-care-3", PetID: "demo-miso", Category: "Annual checkup", Date: demoDate(-2), Note: "Example overdue reminder.", Scheduled: true},
		{ID: "demo-care-4", PetID: "demo-pip", Category: "Habitat refresh", Date: demoDate(habitatInDays), Note: "Replace bedding and tidy the pen.", Scheduled: true},
		{ID: "demo-care-5", PetID: "demo-miso", Category: "Brushing", Date: demoDate(-1), Note: "A short evening grooming session.", Complete: true},
		{ID: "demo-care-6", PetID: "demo-pip", Category: "Nail trim", Date: demoDate(-5), Note: "Completed at home.", Scheduled: true, Complete: true},
		{ID: "demo-care-7", PetID: "demo-maple", Category: "Walk", Date: demoDate(-1), Note: "Easy loop around the park."},
	}
	for _, r := range rows {
		_, err := tx.ExecContext(ctx, `INSERT INTO records VALUES(?,?,?,?,?,?,?)`, r.ID, r.PetID, r.Category, r.Date, r.Note, r.Scheduled, r.Complete)
		if err != nil {
			return err
		}
	}
	return nil
}

func demoDate(days int) string { return time.Now().AddDate(0, 0, days).Format("2006-01-02") }
