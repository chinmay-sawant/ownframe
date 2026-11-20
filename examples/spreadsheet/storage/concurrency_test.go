package storage

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

func TestConcurrentReadsNoDeadlock(t *testing.T) {
	st := openMemory(t)
	ctx := context.Background()

	if _, err := st.SeedWith(ctx, 1, workbook.SeedStress(500)); err != nil {
		t.Fatal(err)
	}

	page, err := st.ListWorkbooks(ctx, 0, 0)
	if err != nil {
		t.Fatal(err)
	}

	id := page.Items[0].ID

	done := make(chan error, 4)

	for i := 0; i < 4; i++ {
		go func(i int) {
			opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if i%2 == 0 {
				_, err := st.ListWorkbooks(opCtx, 0, 0)
				done <- err

				return
			}

			_, err := st.LoadWorkbook(opCtx, id)
			done <- err
		}(i)
	}

	for i := 0; i < 4; i++ {
		if err := <-done; err != nil {
			t.Fatalf("concurrent op: %v", err)
		}
	}
}
