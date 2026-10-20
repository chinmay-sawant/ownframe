package corebackend

import (
	"context"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/storage"
	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/workbook"
)

// Seed writes the shipped fixtures once and loads one workbook: the one
// named, or the first seeded workbook when name is empty.
func (b *Backend) Seed(ctx context.Context, name string) error {
	if _, err := b.st.SeedDummy(ctx); err != nil {
		return err
	}

	return b.load(ctx, name)
}

// SeedDemo writes only the demo workbook and loads it. Tests use it when
// they do not need the 100,000-row stress fixture.
func (b *Backend) SeedDemo(ctx context.Context) error {
	if _, err := b.st.SeedWith(ctx, storage.SeedVersion, workbook.SeedDemo()); err != nil {
		return err
	}

	return b.load(ctx, "")
}

func (b *Backend) load(ctx context.Context, name string) error {
	info, err := b.pick(ctx, name)
	if err != nil {
		return err
	}

	wb, err := b.st.LoadWorkbook(ctx, info.ID)
	if err != nil {
		return err
	}

	wb.SetHistoryLimit(1)
	b.wb = wb

	return nil
}
