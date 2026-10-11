package store

import (
	"fmt"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/model"
	"strings"
	"time"
)

func ValidateBackup(b model.Backup) error {
	if b.Version != 1 {
		return fmt.Errorf("unsupported backup version")
	}
	ids := map[string]bool{}
	for _, p := range b.Pets {
		if strings.TrimSpace(p.ID) == "" || strings.TrimSpace(p.Name) == "" || strings.TrimSpace(p.Species) == "" || ids[p.ID] {
			return fmt.Errorf("invalid or duplicate pet")
		}
		ids[p.ID] = true
	}
	for _, r := range b.Records {
		if strings.TrimSpace(r.ID) == "" || strings.TrimSpace(r.PetID) == "" || strings.TrimSpace(r.Category) == "" || ids[r.ID] || !ids[r.PetID] {
			return fmt.Errorf("invalid or duplicate care record")
		}
		if _, err := time.Parse("2006-01-02", r.Date); err != nil {
			return fmt.Errorf("invalid care date")
		}
		ids[r.ID] = true
	}
	return nil
}
