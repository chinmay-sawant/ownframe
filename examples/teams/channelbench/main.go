package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/teams/app"
)

func main() {
	n := flag.Int("posts", benchPosts, "synthetic posts in active channel")
	flag.Parse()
	s, err := app.New()
	if err != nil {
		panic(err)
	}
	v := s.View()
	v.Section = "channels"
	v.Channels.Posts = makePosts(*n)
	s.Page().SetData(v)
	var b runtime.MemStats
	runtime.ReadMemStats(&b)
	st := time.Now()
	if err := s.Redraw(context.Background()); err != nil {
		panic(err)
	}
	el := time.Since(st)
	var a runtime.MemStats
	runtime.ReadMemStats(&a)
	st2 := s.Page().Stats()
	fmt.Printf("posts=%d redraw=%s alloc=%dKB boxes=%d ops=%d layout=%s list=%s\n",
		*n, el, (a.TotalAlloc-b.TotalAlloc)/benchKB, len(s.Boxes()), st2.Ops, st2.LayoutTime, st2.DisplayListTime)
}
