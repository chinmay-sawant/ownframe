package main

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
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
	return widgetsBtn(label, "list:filter:"+value, value == active)
}

func (a *builderApp) listPanel() *ui.Element {
	files := a.files.visible()
	rows := make([]*ui.Element, 0, len(files))
	for _, f := range files {
		bg := theme.Control()
		if f.Name == a.files.Selected {
			bg = theme.Selected()
		}
		rows = append(rows, ui.Button("").Attr("data-action", "list:select:"+f.Name).
			TextAlign(ui.TextAlignLeft).BackgroundColor(bg).Color(theme.Text()).
			Border(theme.Border()).
			Padding(ui.Edges(ui.Px(12), ui.Px(14))).BorderRadius(ui.Px(8)).
			Children(
				ui.Text(f.Name).FontWeight(ui.FontWeightBold),
				ui.P().Text(f.Kind+"  ·  "+f.Size).Margin(ui.Edges(ui.Px(4), ui.Px(0), ui.Px(0))).
					FontSize(ui.Px(11)).Color(theme.Muted()),
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

	return ui.Section().Flex(ui.Number(1)).MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(12)).
		Children(
			ui.Div().Children(
				a.panelHeading("File list"),
				ui.P().Text("Rows are built in Go with Children(rows...) — no template range.").
					Margin(ui.Edges(ui.Px(7), ui.Px(0), ui.Px(0))).Color(theme.Muted()).FontSize(ui.Px(12)),
			),
			ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(8)).
				Children(
					listFilterButton("All", "all", a.files.Filter),
					listFilterButton("Go", "go", a.files.Filter),
					listFilterButton("Docs", "docs", a.files.Filter),
				),
			ui.Div().MaxWidth(ui.Px(600)).Width(ui.Percent(100)).
				Children(
					ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).
						Children(rows...),
					ui.Div().Padding(ui.Edges(ui.Px(12), ui.Px(15))).BackgroundColor(theme.Surface()).
						BorderRadius(ui.Px(10)).MarginTop(ui.Px(12)).
						Children(
							ui.Text(detail).FontWeight(ui.FontWeightBold),
							ui.P().Text("Showing "+strconv.Itoa(len(files))+" of "+strconv.Itoa(len(a.files.Files))+" files. Click a row to select it.").
								Margin(ui.Edges(ui.Px(8), ui.Px(0), ui.Px(0))).FontSize(ui.Px(12)).Color(theme.Muted()),
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
