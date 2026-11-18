package main

import "fmt"

// App holds the catalog and the tabs. It builds one page per tab.
type App struct {
	cat  *Catalog
	tabs []Tab
}

// NewApp loads the catalog at path and the six tabs.
func NewApp(catalog string) (*App, error) {
	cat, err := LoadCatalog(catalog)
	if err != nil {
		return nil, err
	}

	a := &App{cat: cat, tabs: allTabs()}

	for _, t := range a.tabs {
		for _, n := range t.Names() {
			if _, ok := cat.Status(n); !ok {
				return nil, fmt.Errorf("tab %s names %q, which is not in the catalog", t.ID, n)
			}
		}
	}

	return a, nil
}

// FirstTab is the tab open at start.
func (a *App) FirstTab() string { return a.tabs[0].ID }
