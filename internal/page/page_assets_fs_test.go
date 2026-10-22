package page

import (
	"testing"
	"testing/fstest"
)

func TestAssetResolverSandboxes(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{"img/a.png": {Data: []byte("png")}}
	get := AssetResolver(fsys)
	if data, err := get("img/a.png"); err != nil || string(data) != "png" {
		t.Fatalf("valid: %v %q", err, data)
	}
	for _, bad := range []string{"", "../x", "a/../../x", "/abs", "a\\b", "C:x"} {
		if _, err := get(bad); err == nil {
			t.Fatalf("%q allowed", bad)
		}
	}
}
