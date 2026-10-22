package replay

import "testing"

func TestValidateUserFont(t *testing.T) {
	t.Parallel()
	if err := ValidateUserFont(nil); err == nil {
		t.Fatal("empty allowed")
	}
	if err := ValidateUserFont(make([]byte, MaxUserFontBytes+1)); err == nil {
		t.Fatal("oversize allowed")
	}
	if err := ValidateUserFont([]byte("not a font")); err == nil {
		t.Fatal("garbage allowed")
	}
}
