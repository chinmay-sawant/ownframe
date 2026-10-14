package replay

import (
	_ "embed"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// groupBlendSource is the Kage program that composites one buffered group
// over its backdrop with the CSS blend formula.
//
//go:embed group_blend.kage
var groupBlendSource string

var groupBlendShader struct {
	once sync.Once
	sh   *ebiten.Shader
}

// groupShader compiles the group compositor once. It returns nil when the
// program does not compile, and the caller falls back to source-over so the
// page still paints.
func groupShader() *ebiten.Shader {
	groupBlendShader.once.Do(func() {
		sh, err := ebiten.NewShader([]byte(groupBlendSource))
		if err == nil {
			groupBlendShader.sh = sh
		}
	})

	return groupBlendShader.sh
}
