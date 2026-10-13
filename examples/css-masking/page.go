package main

import (
	"embed"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed index.html assets/*
var files embed.FS

func newPage() (*ownframe.Page, error) {
	html, err := files.ReadFile("index.html")
	if err != nil {
		return nil, err
	}
	p, err := ownframe.New(ownframe.Config{Title: "CSS masking", HTML: string(html), Width: 460, Height: 390})
	if err != nil {
		return nil, err
	}
	for _, name := range []string{"alpha-half.png", "luminance-half.png", "mask-type.svg", "flowers.jpg", "coast.jpg"} {
		data, err := files.ReadFile("assets/" + name)
		if err != nil {
			return nil, fmt.Errorf("read masking example asset %s: %w", name, err)
		}
		p.SetImage("assets/"+name, data)
	}
	return p, nil
}
