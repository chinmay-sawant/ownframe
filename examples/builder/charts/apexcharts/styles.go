package apexcharts

import (
	"embed"
	"io/fs"
	"sort"
	"strings"
)

//go:embed css/*.css
var stylesFS embed.FS

// Stylesheet is the paint for every chart in this set. Append it to a theme:
//
//	Theme: theme.Dark + apexcharts.Stylesheet
//
//nolint:gochecknoglobals // exported read-only stylesheet; embedding needs package state
var Stylesheet = loadStylesheet()

func loadStylesheet() string {
	entries, err := fs.ReadDir(stylesFS, "css")
	if err != nil {
		panic("apexcharts: read css directory: " + err.Error())
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".css") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	var sheet strings.Builder
	for _, name := range names {
		data, err := fs.ReadFile(stylesFS, "css/"+name)
		if err != nil {
			panic("apexcharts: read " + name + ": " + err.Error())
		}
		sheet.Write(data)
		sheet.WriteByte('\n')
	}
	return sheet.String()
}
