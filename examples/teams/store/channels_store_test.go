package store

import (
	"reflect"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/teams/channels"
)

// TestChannelsRoundTrip seeds an empty database, sends a reply, toggles a
// reaction, and checks both changes and the team list survive Save and Load.
func TestChannelsRoundTrip(t *testing.T) {
	st, err := Open(Memory)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = st.Close() }()

	if err := st.Seed(); err != nil {
		t.Fatal(err)
	}

	d, ok, err := st.Load()
	if err != nil || !ok {
		t.Fatalf("Load: %v ok=%v", err, ok)
	}

	page, err := ownframe.New(ownframe.Config{Width: 400, Height: 300,
		HTML: `<input id="channels-reply-av1">`})
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()

	if err := page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	page.SetFormValue("channels-reply-av1", "Round-trip reply")
	channels.Handle(ctx, page, &d.Channels, "channels-reply-send-av1")
	channels.Handle(ctx, page, &d.Channels, "channels-react-av1-0")

	if err := st.Save(d); err != nil {
		t.Fatal(err)
	}

	got, ok, err := st.Load()
	if err != nil || !ok {
		t.Fatalf("Load: %v ok=%v", err, ok)
	}

	if !reflect.DeepEqual(got.Channels.Teams, d.Channels.Teams) {
		t.Errorf("teams = %+v", got.Channels.Teams)
	}

	if n := len(got.Channels.Teams); n != 5 {
		t.Errorf("teams = %d, want 5", n)
	}

	av1 := got.Channels.AllPosts()["av-general"][0]
	last, r := av1.Replies[5], av1.Reactions[0]

	if last.Text != "Round-trip reply" || !last.Own || !r.Mine || r.Count != 5 {
		t.Errorf("reply = %+v reaction = %+v", last, r)
	}

	if got.Channels.ActiveTeam != "avengers" || got.Channels.ActiveChannel != "av-general" {
		t.Errorf("active = %s/%s", got.Channels.ActiveTeam, got.Channels.ActiveChannel)
	}
}
