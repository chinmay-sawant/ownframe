package masking

import "testing"

func TestBlendGatesMatchShaderIDs(t *testing.T) {
	t.Parallel()
	for mode, m := range blendModes {
		gate := m.sep && !BlendNormal(mode)
		if gate && m.id == 0 {
			t.Fatalf("mode %q gated but shader id is 0", mode)
		}
		if !gate && m.id != 0 {
			t.Fatalf("mode %q has shader id %d outside gate", mode, m.id)
		}
		if BlendSeparable(mode) != m.sep || BlendModeID(mode) != m.id {
			t.Fatalf("mode %q table mismatch", mode)
		}
		if BlendSupported(mode) != true {
			t.Fatalf("mode %q not supported", mode)
		}
	}
	if BlendSupported("bogus") || BlendSeparable("bogus") || BlendModeID("bogus") != 0 {
		t.Fatal("bogus mode accepted")
	}
}
