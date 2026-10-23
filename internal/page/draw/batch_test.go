package draw

import "testing"

func TestRedrawBatchRules(t *testing.T) {
	t.Parallel()

	var n int
	var pending, force bool
	if QueueBatch(n, &pending, &force, false) {
		t.Fatal("closed")
	}

	OpenBatch(&n)
	if !QueueBatch(n, &pending, &force, true) || !pending || force {
		t.Fatal("skip")
	}

	OpenBatch(&n)
	if !QueueBatch(n, &pending, &force, false) || !force {
		t.Fatal("force")
	}

	run, forced := CloseBatch(&n, &pending, &force)
	if run || forced || n != 1 || !pending || !force {
		t.Fatalf("inner %v %v %d", run, forced, n)
	}

	run, forced = CloseBatch(&n, &pending, &force)
	if !run || !forced || pending || force || n != 0 {
		t.Fatalf("outer %v %v %d pending %v force %v", run, forced, n, pending, force)
	}

	RestoreBatch(&pending, &force, forced)
	if !pending || !force {
		t.Fatal("restore")
	}

	if run, _ := CloseBatch(&n, &pending, &force); run {
		t.Fatal("empty")
	}
}
