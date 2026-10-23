package player

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// search GETs one iTunes search page.
func (a *App) search(ctx context.Context, term, entity string, limit int) ([]searchResult, error) {
	q := url.Values{}
	q.Set("term", term)
	q.Set("entity", entity)
	q.Set("limit", strconv.Itoa(limit))

	res, err := ownframe.Fetch(ctx, a.base+"/search?"+q.Encode())
	if err != nil {
		return nil, err
	}

	if res.Status != http.StatusOK {
		return nil, fmt.Errorf("player: search status %d", res.Status)
	}

	var out searchResponse
	if err := json.Unmarshal(res.Body, &out); err != nil {
		return nil, err
	}

	return out.Results, nil
}

// searchNow runs a live search when the box has a non-empty value.
func (a *App) searchNow(ctx context.Context) error {
	q := strings.TrimSpace(a.page.Form().FormValue("q"))
	if q == "" {
		return nil
	}

	a.view.Query = q

	if err := a.Load(ctx, q); err != nil {
		a.page.SetData(&a.view)
		a.page.Probe().SetChangedProperties([]string{"Query", "Status"})
	}

	return nil
}
