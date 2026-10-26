package main

import (
	"context"
	"strings"
)

const searchWindowRows = 8

const searchResultLimit = 200

// searchReply carries one async workspace search back to the page loop.
type searchReply struct {
	Token  int
	Result workspaceSearchResult
}

// searchViewState builds the palette-like content search panel.
func (a *App) searchViewState() searchView {
	s := searchView{
		Open:      a.searchOpen,
		Query:     a.searchQuery,
		Case:      a.searchCase,
		Index:     a.searchIndex,
		Files:     countSearchFiles(a.searchHits),
		Note:      a.searchNote,
		Truncated: a.searchTruncated,
		Busy:      a.searchBusy,
	}
	if !s.Open || len(a.searchHits) == 0 {
		return s
	}

	start := max(0, a.searchIndex-searchWindowRows/2)
	end := min(start+searchWindowRows, len(a.searchHits))

	for i := start; i < end; i++ {
		hit := a.searchHits[i]
		s.Rows = append(s.Rows, searchRow{
			Index:    i,
			Path:     hit.Path,
			Name:     baseName(hit.Path),
			Dir:      dirName(hit.Path),
			Line:     hit.Line,
			Col:      hit.Col,
			Before:   hit.Before,
			Match:    hit.Match,
			After:    hit.After,
			Selected: i == a.searchIndex,
		})
	}

	return s
}

func countSearchFiles(hits []workspaceHit) int {
	seen := map[string]struct{}{}

	for _, hit := range hits {
		seen[hit.Path] = struct{}{}
	}

	return len(seen)
}

// openWorkspaceSearch shows the workspace content search panel.
func (a *App) openWorkspaceSearch(ctx context.Context) error {
	if !a.searchOpen {
		a.searchReturn = a.page.Form().FocusID()
		a.searchOpen = true
	}

	a.refresh()

	if err := a.page.Redraw(ctx); err != nil {
		return err
	}

	if err := a.page.Form().Focus(ctx, "workspace-search"); err != nil {
		return err
	}

	return a.page.Edit().SelectAll(ctx)
}

// closeWorkspaceSearch hides the panel and clears the query.
func (a *App) closeWorkspaceSearch(ctx context.Context) error {
	previous := a.searchReturn
	a.cancelSearch()
	a.searchOpen = false
	a.searchQuery = ""
	a.searchCase = false
	a.searchHits = nil
	a.searchIndex = 0
	a.searchNote = ""
	a.searchTruncated = false
	a.searchBusy = false
	a.searchReturn = ""
	a.refresh()

	if err := a.page.Redraw(ctx); err != nil {
		return err
	}

	if previous != "" && a.page.Form().FocusID() == "" {
		return a.page.Form().Focus(ctx, previous)
	}

	return nil
}

// runWorkspaceSearch starts one search. Results come back on the page loop
// through receiveSearch so the workspace pointer is never read concurrently
// by the renderer and the scanner.
func (a *App) runWorkspaceSearch(ctx context.Context) {
	a.cancelSearch()
	a.drainSearchResults()

	query := strings.TrimSpace(a.searchQuery)
	if query == "" {
		a.searchHits = nil
		a.searchIndex = 0
		a.searchNote = ""
		a.searchTruncated = false
		a.searchBusy = false
		a.refresh()

		return
	}

	a.searchToken++
	token := a.searchToken
	a.searchBusy = true
	a.searchNote = "Searching…"

	ctx, cancel := context.WithCancel(ctx)
	a.searchCancel = cancel
	ws := a.ws
	caseSensitive := a.searchCase
	result := a.searchResult

	go func() {
		found := searchWorkspace(ctx, ws, query, caseSensitive, searchResultLimit)

		select {
		case result <- searchReply{Token: token, Result: found}:
		default:
		}

		a.page.WantFrameIn(0)
	}()

	a.refresh()
}

// receiveSearch drains one finished search on the page loop.
func (a *App) receiveSearch(ctx context.Context) bool {
	for {
		select {
		case reply := <-a.searchResult:
			if reply.Token != a.searchToken {
				continue
			}

			a.searchHits = reply.Result.Hits
			a.searchNote = reply.Result.Note
			a.searchTruncated = reply.Result.Truncated
			a.searchBusy = false
			a.searchIndex = 0
			a.refresh()

			if err := a.page.Redraw(ctx); err != nil {
				return false
			}

			return true
		default:
			return false
		}
	}
}

// drainSearchResults drops a result the next query replaced.
func (a *App) drainSearchResults() {
	for {
		select {
		case <-a.searchResult:
		default:
			return
		}
	}
}

// cancelSearch stops the running scan.
func (a *App) cancelSearch() {
	if a.searchCancel != nil {
		a.searchCancel()
		a.searchCancel = nil
	}
}

// toggleSearchCase flips the case switch and reruns the query.
func (a *App) toggleSearchCase(ctx context.Context) {
	a.searchCase = !a.searchCase
	a.searchIndex = 0
	a.runWorkspaceSearch(ctx)
}

// searchStep moves the selected result.
func (a *App) searchStep(delta int) {
	if len(a.searchHits) == 0 {
		return
	}

	a.searchIndex = (a.searchIndex + delta + len(a.searchHits)) % len(a.searchHits)
	a.refresh()
}

// openSearchHit opens one result at its source line.
func (a *App) openSearchHit(index int) {
	if index < 0 || index >= len(a.searchHits) {
		return
	}

	hit := a.searchHits[index]
	a.searchIndex = index
	a.cancelSearch()
	a.searchOpen = false
	a.searchQuery = ""
	a.searchHits = nil
	a.openFile(hit.Path)
	a.scrollToSourceLine(hit.Line - 1)
}
