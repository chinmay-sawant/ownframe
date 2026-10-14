package displaypatch

import "testing"

func TestDebugLogReport(t *testing.T) {
	d := DebugLog{Enabled: true}
	d.Record(PatchResult{Changed: []Op{{}, {}}, Added: []Op{{}}, Removed: []Op{{}, {}, {}}})
	if d.Patched != 2 || d.Added != 1 || d.Removed != 3 {
		t.Fatalf("counts %+v", d)
	}
	if got := d.String(); got != "patched=2 added=1 removed=3" {
		t.Fatalf("got %q", got)
	}
}

func TestDebugLogDisabledIsBlank(t *testing.T) {
	d := DebugLog{Patched: 2, Added: 1, Removed: 3, Enabled: false}
	if got := d.String(); got != "" {
		t.Fatalf("got %q", got)
	}
}
