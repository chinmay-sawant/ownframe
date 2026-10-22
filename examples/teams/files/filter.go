package files

import "strings"

// matches reports whether f belongs under the filter and the query.
func matches(f File, filter, query string) bool {
	if query != "" && !strings.Contains(strings.ToLower(f.Name), query) {
		return false
	}

	switch filter {
	case "teams":
		return f.Team
	case "shared":
		return f.Shared
	case "starred":
		return f.Starred
	}

	return true
}

// rebuild sets Files to the view of all that Filter and Query select.
func (d *Data) rebuild() {
	query := strings.ToLower(strings.TrimSpace(d.Query))
	rows := make([]File, 0, len(d.all))

	for _, f := range d.all {
		if matches(f, d.Filter, query) {
			rows = append(rows, f)
		}
	}

	d.Files = rows
}

// index returns the position of id in all, or -1.
func (d *Data) index(id string) int {
	for i := range d.all {
		if d.all[i].ID == id {
			return i
		}
	}

	return -1
}
