package main

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"testing"
)

// patchCostTargets are the fixture elements the cost report names.
var patchCostTargets = []string{
	"play-all", "catalog", "catalog-grid", "catalog-row-01", "catalog-state-01",
	"catalog-target-webkit_animation", "start-property-webkit_animation",
	"metadata", "preview-metadata",
}

// patchCostProbeClass is the class the probe puts on an element. The state
// span keeps the class it carries, so the probe changes no size.
var patchCostProbeClass = map[string]string{
	"catalog-state-01": "catalog-state zz-probe",
}

// patchCostSource returns the executed fixture template. The catalog rows
// carry no measured values, so the page is the fixture with empty readouts.
func patchCostSource(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "css-preview", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	tpl := template.Must(template.New("").Parse(string(raw)))
	v := &view{Started: map[string]bool{}}
	v.Catalog = catalogRows(nil)
	v.Rows = v.Catalog
	var body bytes.Buffer
	if err := tpl.Execute(&body, v); err != nil {
		t.Fatal(err)
	}
	return body.Bytes()
}

// patchCostTheme returns the fixture stylesheet.
func patchCostTheme(t *testing.T) []byte {
	t.Helper()
	theme, err := os.ReadFile(filepath.Join("testdata", "css-preview", "styles.css"))
	if err != nil {
		t.Fatal(err)
	}
	return theme
}
