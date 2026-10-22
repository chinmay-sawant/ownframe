package main

import (
	"fmt"

	"github.com/chinmay-sawant/ownframe/examples/teams/channels"
)

func makePosts(n int) []channels.Post {
	out := make([]channels.Post, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, channels.Post{
			ID:       fmt.Sprintf("bench-%d", i),
			Author:   "Bench User",
			Initials: "BU",
			Color:    "blue",
			Time:     "10:00 AM",
			Subject:  fmt.Sprintf("Post %d subject", i),
			Text:     "Synthetic body for channel scaling. Short text keeps layout comparable across runs.",
			Likes:    i % benchLike,
		})
	}
	return out
}
