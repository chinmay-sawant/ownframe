package files

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// Handle applies one click action; it reports whether the action was ours.
func Handle(_ context.Context, page *ownframe.Page, d *Data, action string) bool {
	switch {
	case action == "files-query":
		d.Query = page.Form().FormValue("files-search")
		d.Active = ""
		d.rebuild()
	case strings.HasPrefix(action, "files-filter-"):
		return setFilter(d, strings.TrimPrefix(action, "files-filter-"))
	case strings.HasPrefix(action, "files-open-"):
		d.Active = strings.TrimPrefix(action, "files-open-")
	case strings.HasPrefix(action, "files-detail-star-"):
		star(d, strings.TrimPrefix(action, "files-detail-star-"))
	case strings.HasPrefix(action, "files-star-"):
		star(d, strings.TrimPrefix(action, "files-star-"))
	default:
		return false
	}

	return true
}

// setFilter switches the list to one named filter and closes the detail.
func setFilter(d *Data, name string) bool {
	switch name {
	case "recent", "teams", "shared", "starred":
	default:
		return false
	}

	d.Filter = name
	d.Active = ""
	d.rebuild()

	return true
}

// star flips one file's star, refreshes the view, and clears Active when
// the rebuilt view dropped it.
func star(d *Data, id string) {
	i := d.index(id)
	if i < 0 {
		return
	}

	d.all[i].Starred = !d.all[i].Starred
	d.rebuild()

	if d.Active == "" {
		return
	}

	for _, f := range d.Files {
		if f.ID == d.Active {
			return
		}
	}

	d.Active = ""
}
