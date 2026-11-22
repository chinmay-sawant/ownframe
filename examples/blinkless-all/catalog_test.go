package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFakeCatalog(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}

	const body = `{"properties":[{"name":"color","status":"implemented"}]}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestResolveCatalogPathPrefersPinnedModuleOverSibling(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	pinned := filepath.Join(tmp, "modcache", catalogRelPath())
	sibling := filepath.Join(tmp, "sibling", "blinkless", catalogRelPath())
	writeFakeCatalog(t, pinned)
	writeFakeCatalog(t, sibling)

	got := resolveCatalogPath("", func(string) string { return pinned }, []string{sibling})
	if got != pinned {
		t.Fatalf("resolveCatalogPath = %q, want pinned %q", got, pinned)
	}
}

func TestResolveCatalogPathEnvOverride(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	pinned := filepath.Join(tmp, "modcache", catalogRelPath())
	overrideRoot := filepath.Join(tmp, "override")
	overrideFile := filepath.Join(overrideRoot, catalogRelPath())
	directDir := filepath.Join(tmp, "direct")
	directFile := filepath.Join(directDir, "properties.json")

	writeFakeCatalog(t, pinned)
	writeFakeCatalog(t, overrideFile)
	writeFakeCatalog(t, directFile)

	for _, tc := range []struct {
		name string
		env  string
		want string
	}{
		{"repo root", overrideRoot, overrideFile},
		{"catalog dir", directDir, directFile},
		{"direct file", directFile, directFile},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := resolveCatalogPath(tc.env, func(string) string { return pinned }, nil)
			if got != tc.want {
				t.Fatalf("resolveCatalogPath(%q) = %q, want %q", tc.env, got, tc.want)
			}
		})
	}
}

func TestResolveCatalogPathFallsBackToSibling(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	sibling := filepath.Join(tmp, "blinkless", catalogRelPath())
	writeFakeCatalog(t, sibling)

	got := resolveCatalogPath("", func(string) string { return "" }, []string{sibling})
	if got != sibling {
		t.Fatalf("resolveCatalogPath fallback = %q, want sibling %q", got, sibling)
	}
}

func TestCatalogPathMatchesPinnedModule(t *testing.T) {
	t.Setenv(catalogDirEnv, "")

	want := moduleFile(catalogRelPath())
	if want == "" {
		t.Skip("go list could not resolve pinned blinkless module directory")
	}

	got := catalogPath()
	if filepath.Clean(got) != filepath.Clean(want) {
		t.Fatalf("catalogPath() = %q, want pinned module catalog %q", got, want)
	}

	if _, err := LoadCatalog(got); err != nil {
		t.Fatalf("LoadCatalog(%q): %v", got, err)
	}

	if !strings.HasSuffix(got, catalogRelPath()) {
		t.Fatalf("catalogPath() = %q, missing suffix %q", got, catalogRelPath())
	}
}
