package main

import (
	_ "embed"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	//go:embed blackhole.kage
	holeSource []byte
	//go:embed bloom.kage
	bloomSource []byte
	//go:embed post.kage
	postSource []byte
)

type shaderPipeline struct {
	hole  *ebiten.Shader
	bloom *ebiten.Shader
	post  *ebiten.Shader
}

func compileShaders() (*shaderPipeline, error) {
	h, err := ebiten.NewShader(holeSource)
	if err != nil {
		return nil, fmt.Errorf("compile blackhole shader: %w", err)
	}
	b, err := ebiten.NewShader(bloomSource)
	if err != nil {
		return nil, fmt.Errorf("compile bloom shader: %w", err)
	}
	p, err := ebiten.NewShader(postSource)
	if err != nil {
		return nil, fmt.Errorf("compile post shader: %w", err)
	}
	return &shaderPipeline{hole: h, bloom: b, post: p}, nil
}
