package replay

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestGroupShaderCompiles pins the Kage program: the compositor's modes are
// useless if the shader does not parse.
func TestGroupShaderCompiles(t *testing.T) {
	t.Parallel()

	if _, err := ebiten.NewShader([]byte(groupBlendSource)); err != nil {
		t.Fatal(err)
	}
}
