package main

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// listFile is one row in the list teaching panel.
type listFile struct {
	Name string
	Kind string
	Size string
}

// listState owns the filter, selection, and rows for the list panel.
// builderApp carries one of these as a.files; it is not redeclared here.
type listState struct {
	Filter   string
	Selected string
	Files    []listFile
}

func defaultListState() listState {
	return listState{
		Filter:   "all",
		Selected: "app.go",
		Files: []listFile{
			{Name: "main.go", Kind: "go", Size: "3.2 KB"},
			{Name: "app.go", Kind: "go", Size: "4.1 KB"},
			{Name: "view.go", Kind: "go", Size: "5.8 KB"},
			{Name: "screen_list.go", Kind: "go", Size: "2.9 KB"},
			{Name: "README.md", Kind: "docs", Size: "1.4 KB"},
			{Name: "GUIDE.md", Kind: "docs", Size: "2.2 KB"},
			{Name: "config.json", Kind: "data", Size: "0.8 KB"},
			{Name: "theme.json", Kind: "data", Size: "1.1 KB"},
		},
	}
}

// visible returns the files matching the current filter. The "all" filter
// shows everything; "go" shows Go sources; "docs" shows docs and data files.
// Any other non-empty filter falls back to a case-insensitive substring
// match against name and kind.
func (s *listState) visible() []listFile {
	if s == nil {
		return nil
	}
	filter := strings.ToLower(strings.TrimSpace(s.Filter))
	if filter == "" || filter == "all" {
		out := make([]listFile, len(s.Files))
		copy(out, s.Files)
		return out
	}
	matched := make([]listFile, 0, len(s.Files))
	for _, f := range s.Files {
		name := strings.ToLower(f.Name)
		kind := strings.ToLower(f.Kind)
		switch filter {
		case "go":
			if strings.HasSuffix(name, ".go") || kind == "go" {
				matched = append(matched, f)
			}
		case "docs", "doc", "data":
			if strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".json") ||
				kind == "docs" || kind == "doc" || kind == "data" {
				matched = append(matched, f)
			}
		default:
			if strings.Contains(name, filter) || strings.Contains(kind, filter) {
				matched = append(matched, f)
			}
		}
	}
	return matched
}

func listFilterButton(label, value, active string) *ui.Element {
	bg := "#2b3950"
	if active == value || (active == "" && value == "all") {
		bg = "#355596"
	}
	return ui.Button(label).Attr("data-action", "list:filter:"+value).
		BackgroundColor(bg).Color("#f1f5fd").
		CSS("border", "1px solid #40546f").
		Padding("9px 12px").BorderRadius("8px").FontWeight("bold").FontSize("12px")
}

func (a *builderApp) listPanel() *ui.Element {
	files := a.files.visible()
	rows := make([]*ui.Element, 0, len(files))
	for _, f := range files {
		bg := "#26364e"
		if f.Name == a.files.Selected {
			bg = "#355596"
		}
		rows = append(rows, ui.Button("").Attr("data-action", "list:select:"+f.Name).
			CSS("text-align", "left").BackgroundColor(bg).Color("#ecf2fb").
			CSS("border", "1px solid #3e526e").
			Padding("12px 14px").BorderRadius("8px").
			Children(
				ui.Text(f.Name).FontWeight("bold"),
				ui.P().Text(f.Kind+"  ·  "+f.Size).Margin("4px 0 0").
					FontSize("11px").Color("#a0b6d0"),
			))
	}

	detail := "Select a file to see its details."
	for _, f := range a.files.Files {
		if f.Name == a.files.Selected {
			detail = "Selected " + f.Name + " — " + f.Kind + ", " + f.Size + "."
			break
		}
	}
	if a.files.Selected != "" && detail == "Select a file to see its details." {
		detail = "Selected " + a.files.Selected + " is hidden by the current filter."
	}

	return ui.Section().CSS("flex", "1").CSS("min-width", "0px").CSS("min-height", "0px").
		Display("flex").FlexDirection("column").Gap("12px").
		Children(
			ui.Div().Children(
				ui.Text("File list").FontSize("24px").FontWeight("bold").Color("#f4f7ff"),
				ui.P().Text("Rows are built in Go with Children(rows...) — no template range.").
					Margin("7px 0 0").Color("#a9bbd1").FontSize("12px"),
			),
			ui.Div().Display("flex").Gap("8px").
				Children(
					listFilterButton("All", "all", a.files.Filter),
					listFilterButton("Go", "go", a.files.Filter),
					listFilterButton("Docs", "docs", a.files.Filter),
				),
			ui.Div().CSS("max-width", "600px").CSS("width", "100%").
				Children(
					ui.Div().Display("flex").FlexDirection("column").Gap("8px").
						Children(rows...),
					ui.Div().Padding("12px 15px").BackgroundColor("#202d42").
						BorderRadius("10px").CSS("margin-top", "12px").
						Children(
							ui.Text(detail).FontWeight("bold"),
							ui.P().Text("Showing "+strconv.Itoa(len(files))+" of "+strconv.Itoa(len(a.files.Files))+" files. Click a row to select it.").
								Margin("8px 0 0").FontSize("12px").Color("#abbcd2"),
						),
				),
		)
}

// handleList handles "list:*" actions, updating a.files and a.notice.
// It reports true when the action was consumed.
func (a *builderApp) handleList(action string) bool {
	rest, ok := strings.CutPrefix(action, "list:")
	if !ok {
		return false
	}
	if value, ok := strings.CutPrefix(rest, "filter:"); ok {
		switch normalized := strings.ToLower(strings.TrimSpace(value)); normalized {
		case "", "all":
			a.files.Filter = "all"
			a.notice = "Showing all files."
		case "go":
			a.files.Filter = "go"
			a.notice = "Filtered to Go sources."
		case "docs", "doc", "data":
			a.files.Filter = "docs"
			a.notice = "Filtered to docs and data files."
		default:
			a.files.Filter = normalized
			a.notice = "Filtered to " + normalized + "."
		}
		return true
	}
	if name, ok := strings.CutPrefix(rest, "select:"); ok {
		for _, f := range a.files.Files {
			if f.Name == name {
				a.files.Selected = name
				a.notice = "Selected " + name + "."
				return true
			}
		}
		a.notice = "Unknown file " + name + "."
		return true
	}
	return false
}
