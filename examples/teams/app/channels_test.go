package app

import (
	"context"
	"testing"
)

func ch(t *testing.T, ok bool) {
	if !ok {
		t.Fatal("ch")
	}
}

func ck(t *testing.T, a *App, id string) {
	b, ok := findBox(a.Boxes(), id)
	if !ok {
		t.Fatal(id)
	}
	if err := a.Click(context.Background(), b.X+b.W/2, b.Y+3); err != nil {
		t.Fatal(err)
	}
}

func cha(t *testing.T) *App {
	a := newApp(t)
	openSection(t, a, "channels")
	return a
}

func TestChannelsNav(t *testing.T) {
	t.Parallel()
	a := cha(t)
	ck(t, a, "channels-pin-av2")
	ck(t, a, "channels-tab-files")
	ck(t, a, "channels-star-f2")
	c := a.View().Channels
	ch(t, c.Posts[1].Pinned && c.Files[1].Starred && c.Tab == "files")
	ck(t, a, "channels-team-wakanda")
	ck(t, a, "channels-team-wakanda")
	ch(t, !a.View().Channels.Teams[1].Expanded)
	ck(t, a, "channels-team-wakanda")
	ck(t, a, "channels-channel-wk-general")
	c = a.View().Channels
	ch(t, c.Teams[1].Expanded && c.Tab == "posts" && c.ActiveChannel == "wk-general")
}

func TestChannelsThread(t *testing.T) {
	t.Parallel()
	a := cha(t)
	ck(t, a, "channels-post-av1")
	_, ok4 := findBox(a.Boxes(), "channels-reply-av1-r4")
	ch(t, a.View().Channels.Posts[0].Expanded && ok4)
	ck(t, a, "channels-post-av2")
	c := a.View().Channels
	ch(t, !c.Posts[0].Expanded && c.Posts[1].Expanded)
	a.Page().SetFormValue("channels-reply-av2", " On it ")
	ck(t, a, "channels-reply-send-av2")
	p := a.View().Channels.Posts[1]
	r := p.Replies[len(p.Replies)-1]
	ch(t, r.Own && r.Text == "On it" && a.Page().FormValue("channels-reply-av2") == "")
}
