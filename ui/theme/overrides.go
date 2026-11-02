package theme

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// Overrides contains CSS custom properties. Empty overrides inherit the current
// theme. Names include the -- prefix; applications may add their own names.
type Overrides map[string]string

// Apply scopes overrides to an element and its descendants. A descendant's own
// overrides take precedence. Values are validated when the tree is rendered.
func (o Overrides) Apply(element *ui.Element) *ui.Element {
	for name, value := range o {
		element.CustomProperty(name, value)
	}
	return element
}

// Stylesheet returns root-scoped properties for Config.Theme or Page.SetTheme.
// Append it to a built-in theme to override application-wide defaults.
func (o Overrides) Stylesheet() (string, error) {
	// Reuse the builder's declaration validation before emitting a stylesheet.
	if _, err := ui.Render(o.Apply(ui.Div())); err != nil {
		return "", err
	}
	names := make([]string, 0, len(o))
	for name := range o {
		names = append(names, name)
	}
	sort.Strings(names)
	var sheet strings.Builder
	sheet.WriteString(":root{")
	for _, name := range names {
		sheet.WriteString(name + ":" + o[name] + ";")
	}
	sheet.WriteByte('}')
	return sheet.String(), nil
}

// LoadOverrides reads a JSON object of custom properties. fs.FS supports both
// os.DirFS directories and embedded assets. The file is validated before use.
func LoadOverrides(files fs.FS, name string) (Overrides, error) {
	data, err := fs.ReadFile(files, name)
	if err != nil {
		return nil, fmt.Errorf("theme: read overrides: %w", err)
	}
	var overrides Overrides
	if err := json.Unmarshal(data, &overrides); err != nil {
		return nil, fmt.Errorf("theme: decode overrides: %w", err)
	}
	if overrides == nil {
		return nil, fmt.Errorf("theme: overrides must be a JSON object")
	}
	if _, err := overrides.Stylesheet(); err != nil {
		return nil, fmt.Errorf("theme: invalid overrides: %w", err)
	}
	return overrides, nil
}
