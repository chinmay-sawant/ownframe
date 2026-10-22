package store

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/teams/activity"
)

// TestActivityRoundTrip checks a read item and the filter state survive Save
// and Load.
func TestActivityRoundTrip(t *testing.T) {
	st, err := Open(Memory)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	defer func() { _ = st.Close() }()

	if err := st.Seed(); err != nil {
		t.Fatalf("Seed: %v", err)
	}

	d, ok, err := st.Load()
	if err != nil || !ok {
		t.Fatalf("Load: ok=%v err=%v", ok, err)
	}

	ctx := context.Background()
	activity.Handle(ctx, nil, &d.Activity, "activity-filter-mentions")
	activity.Handle(ctx, nil, &d.Activity, "activity-open-a1")
	activity.Handle(ctx, nil, &d.Activity, "activity-reply")

	if err := st.Save(d); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, ok, err := st.Load()
	if err != nil || !ok {
		t.Fatalf("Load: ok=%v err=%v", ok, err)
	}

	if len(got.Activity.Items) != len(d.Activity.Items) {
		t.Fatalf("items = %d, want %d", len(got.Activity.Items), len(d.Activity.Items))
	}

	if got.Activity.Items[0].Unread {
		t.Fatal("a1 came back unread")
	}

	if !got.Activity.Items[0].Replied || got.Activity.Items[0].Replies != 3 {
		t.Fatalf("a1 = %+v", got.Activity.Items[0])
	}

	if got.Activity.Items[2].Where != d.Activity.Items[2].Where ||
		got.Activity.Items[2].Text != d.Activity.Items[2].Text {
		t.Fatalf("a3 = %+v", got.Activity.Items[2])
	}

	if got.Activity.Filter != "mentions" || got.Activity.Active != "a1" {
		t.Fatalf("Filter=%q Active=%q", got.Activity.Filter, got.Activity.Active)
	}
}
