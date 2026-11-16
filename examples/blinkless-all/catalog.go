package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Catalog is the status of every CSS property, read from the blinkless
// checkout. It is the one source for the counts at the top of the window.
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

// catalogPath finds the catalog in the sibling blinkless checkout.
func catalogPath() string {
	rel := filepath.Join("blinkless", "testdata", "css", "catalog", "properties.json")
	candidates := []string{
		filepath.Join("..", rel),
		rel,
		filepath.Join("..", "..", "..", rel),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return candidates[0]
}
