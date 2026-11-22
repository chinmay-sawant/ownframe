package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

const (
	blinklessModule = "github.com/chinmay-sawant/blinkless"
	catalogDirEnv   = "BLINKLESS_CATALOG_DIR"
)

// Catalog is the status of every CSS property, read from the pinned blinkless
// module (or an explicit override / sibling fallback). It is the one source for
// the counts at the top of the window.
type Catalog struct {
	status map[string]string
}

// Status values used by the catalog.
const (
	StatusImplemented = "implemented"
	StatusPartial     = "partial"
	StatusUnsupported = "unsupported"
	StatusIgnored     = "intentionally ignored"
)

// LoadCatalog reads the properties.json file at path.
func LoadCatalog(path string) (*Catalog, error) {
	raw, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	var doc struct {
		Properties []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"properties"`
	}

	if err = json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("catalog %s: %w", path, err)
	}

	c := &Catalog{status: make(map[string]string, len(doc.Properties))}
	for _, r := range doc.Properties {
		c.status[r.Name] = r.Status
	}

	return c, nil
}

// Total is the number of properties in the catalog.
func (c *Catalog) Total() int { return len(c.status) }

// Count is how many properties have this status.
func (c *Catalog) Count(status string) int {
	n := 0

	for _, s := range c.status {
		if s == status {
			n++
		}
	}

	return n
}

// Status returns the status of one property and whether the catalog has it.
func (c *Catalog) Status(name string) (string, bool) {
	s, ok := c.status[name]

	return s, ok
}

// Tally counts the statuses of the given properties.
func (c *Catalog) Tally(names []string) map[string]int {
	out := map[string]int{}

	for _, n := range names {
		s, ok := c.status[n]
		if !ok {
			s = "missing"
		}

		out[s]++
	}

	return out
}

func catalogRelPath() string {
	return filepath.Join("testdata", "css", "catalog", "properties.json")
}

func siblingCatalogCandidates() []string {
	rel := filepath.Join("blinkless", catalogRelPath())

	return []string{
		filepath.Join("..", rel),
		rel,
		filepath.Join("..", "..", "..", rel),
	}
}

// catalogPath resolves properties.json from BLINKLESS_CATALOG_DIR when set,
// otherwise from the pinned blinkless module directory first, falling back to
// a sibling blinkless checkout only if the pinned module lookup fails.
func catalogPath() string {
	return resolveCatalogPath(os.Getenv(catalogDirEnv), pinnedModuleFile, siblingCatalogCandidates())
}

func resolveCatalogPath(envDir string, findModule func(string) string, candidates []string) string {
	rel := catalogRelPath()

	if envDir = strings.TrimSpace(envDir); envDir != "" {
		for _, path := range []string{
			filepath.Join(envDir, rel),
			filepath.Join(envDir, "properties.json"),
			envDir,
		} {
			if isRegularFile(path) {
				return path
			}
		}

		return filepath.Join(envDir, rel)
	}

	if findModule != nil {
		if path := findModule(rel); path != "" {
			return path
		}
	}

	for _, path := range candidates {
		if isRegularFile(path) {
			return path
		}
	}

	if len(candidates) > 0 {
		return candidates[0]
	}

	return rel
}

// pinnedModuleFile resolves rel inside the pinned blinkless module directory,
// checking compiled build info against the local module cache first (zero
// subprocesses, works from any working directory) before falling back to
// `go list -m`.
func pinnedModuleFile(rel string) string {
	if path := buildInfoModuleFile(rel); path != "" {
		return path
	}

	return moduleFile(rel)
}

func buildInfoModuleFile(rel string) string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}

	for _, dep := range info.Deps {
		if dep.Path != blinklessModule {
			continue
		}

		modPath, modVersion := dep.Path, dep.Version
		if dep.Replace != nil {
			if dep.Replace.Version == "" && dep.Replace.Path != "" {
				candidate := filepath.Join(dep.Replace.Path, rel)
				if isRegularFile(candidate) {
					return candidate
				}

				return ""
			}

			modPath, modVersion = dep.Replace.Path, dep.Replace.Version
		}

		if modVersion == "" || modVersion == "(devel)" {
			return ""
		}

		escaped := escapeModulePath(modPath)
		for _, root := range moduleCacheDirs() {
			candidate := filepath.Join(root, escaped+"@"+modVersion, rel)
			if isRegularFile(candidate) {
				return candidate
			}
		}

		return ""
	}

	return ""
}

func moduleCacheDirs() []string {
	var dirs []string

	if cache := strings.TrimSpace(os.Getenv("GOMODCACHE")); cache != "" {
		dirs = append(dirs, cache)
	}

	if gopath := strings.TrimSpace(os.Getenv("GOPATH")); gopath != "" {
		for _, entry := range filepath.SplitList(gopath) {
			if entry != "" {
				dirs = append(dirs, filepath.Join(entry, "pkg", "mod"))
			}
		}
	}

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		dirs = append(dirs, filepath.Join(home, "go", "pkg", "mod"))
	}

	return dirs
}

func escapeModulePath(path string) string {
	var b strings.Builder
	b.Grow(len(path))

	for i := range len(path) {
		c := path[i]
		if c >= 'A' && c <= 'Z' {
			b.WriteByte('!')
			b.WriteByte(c + ('a' - 'A'))

			continue
		}

		b.WriteByte(c)
	}

	return b.String()
}

// moduleFile returns rel inside the pinned blinkless module via `go list -m`,
// or "" when the go tool cannot find it.
func moduleFile(rel string) string {
	const moduleLookupTimeout = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), moduleLookupTimeout)
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		"go",
		"list",
		"-m",
		"-f",
		"{{.Dir}}",
		blinklessModule,
	)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	path := filepath.Join(strings.TrimSpace(string(out)), rel)
	if !isRegularFile(path) {
		return ""
	}

	return path
}

func isRegularFile(path string) bool {
	clean := filepath.Clean(path)
	root, err := os.OpenRoot(filepath.Dir(clean))
	if err != nil {
		return false
	}
	defer func() { _ = root.Close() }()

	info, err := root.Stat(filepath.Base(clean))

	return err == nil && !info.IsDir()
}
