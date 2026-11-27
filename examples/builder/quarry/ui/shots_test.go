package ui

import (
	"fmt"
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/store"
)

// TestShotsAllViews writes one screenshot per view and theme into the test
// artifact directory. It only runs when QUARRY_SHOTS=1 is set, because it is
// a visual-audit driver, not a regression check.
func TestShotsAllViews(t *testing.T) {
	if os.Getenv("QUARRY_SHOTS") == "" {
		t.Skip("set QUARRY_SHOTS=1 to capture the visual audit screenshots")
	}
	for _, dark := range []bool{true, false} {
		name := map[bool]string{true: "dark", false: "light"}[dark]
		r := newRig(t, 1440, 960, "")
		if !dark {
			r.tap("tb-theme")
		}
		shot := func(n string) { r.png(fmt.Sprintf("%s-%s", n, name)) }
		shot("01-empty-workbench")
		r.a.activeTabPtr().ed = newEditor("SELECT o.id, o.status, o.total_cents, u.full_name\nFROM orders o JOIN users u ON u.id = o.user_id\nWHERE o.total_cents > 1000 -- big orders\nORDER BY o.id")
		r.chord("enter", false)
		r.settle()
		shot("02-query-results")
		r.tap("rs-explain")
		shot("03-explain-empty")
		r.tap("ed-explain")
		r.settle()
		shot("03b-explain")
		r.tap("rs-messages")
		shot("04-messages")
		r.tap("rs-results")
		r.tap("rs-inspector")
		r.tap("cell-1-1")
		shot("05-query-inspector")
		r.openTableByName("audit_events")
		shot("06-table")
		r.tap("tbl-inspector")
		r.tap("cell-0-4")
		shot("07-table-inspector")
		r.must(r.a.menuActionFor("tree-table:main.orders", "structure"))
		shot("08-structure-columns")
		r.tap("st-sub-indexes")
		shot("09-structure-indexes")
		r.tap("st-sub-ddl")
		shot("10-structure-ddl")
		r.openMonitor()
		for i := 0; i < 20; i++ {
			r.a.PushSyntheticSample(i)
		}
		r.must(r.a.redraw(r.ctx))
		shot("11-monitor")
		r.chord("k", false)
		shot("12-palette")
		r.key("escape")
		r.tap("rail-connections")
		shot("13-connection-manager")
		r.key("escape")
		r.tap("rail-history")
		shot("14-history")
		r.key("escape")
		r.tap("rail-settings")
		shot("15-settings")
		r.key("escape")
		r.tap("tb-conn")
		shot("16-conn-menu")
		r.key("escape")
		r.openTableByName("users")
		r.a.activeTabPtr().ed = nil
		r.must(r.a.openExport(r.ctx, true))
		shot("17-export")
		r.key("escape")
		r.must(r.a.openImport(r.ctx))
		shot("18-import")
		r.key("escape")
		p := newRig(t, 1440, 700, "mock:postgres")
		if !dark {
			p.tap("tb-theme")
		}
		for _, s := range p.a.schema {
			for _, g := range s.Children {
				p.a.expanded["group:"+s.Name+":"+g.Name] = true
			}
		}
		p.must(p.a.redraw(p.ctx))
		p.png("19-tree-pg-" + name)
		for i := 0; i < 30; i++ {
			_, _ = r.a.prefs.AddHistory(r.ctx, store.HistoryEntry{ConnectionID: r.a.cfg.ID, SQL: fmt.Sprintf("SELECT count(*) FROM orders WHERE id > %d", i), Rows: 1})
		}
		r.tap("rail-history")
		shot("20-history-long")
		r.key("escape")
	}
}
