package app

import "testing"

func TestChannelsReact(t *testing.T) {
	t.Parallel()
	a := cha(t)
	ck(t, a, "channels-react-av1-0")
	p := a.View().Channels.Posts[0]
	ch(t, p.Reactions[0].Mine && p.Reactions[0].Count == 5)
	ck(t, a, "channels-post-av1")
	ck(t, a, "channels-react-av1-0")
	p = a.View().Channels.Posts[0]
	ch(t, !p.Reactions[0].Mine && p.Reactions[0].Count == 4 && p.Expanded)
	ck(t, a, "channels-picker-av1")
	ck(t, a, "channels-add-av1-2")
	p = a.View().Channels.Posts[0]
	ch(t, !p.Picker && len(p.Reactions) == 3 && p.Reactions[2].Mine)
}
