package page

import (
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/source"
	"html/template"
	"io/fs"
	"os"
	"reflect"
)

const (
	// MaxSourceSize caps a page source in bytes.
	MaxSourceSize = source.MaxBytes

	// MaxDOMDepth caps parse-time nesting (GEN-29): layout recursion past
	// it is a fatal stack overflow that recover cannot catch.
	MaxDOMDepth = source.MaxDOMDepth

	// MaxDOMNodes caps the parse-time element count.
	MaxDOMNodes = source.MaxDOMNodes

	// MaxDataImageBytes caps one data:image/* URL at resolve time (GEN-30).
	// The cap applies to the encoded URL, so decode can never exceed it.
	MaxDataImageBytes = source.MaxDataImageBytes
)

var (
	// ErrSourceTooLarge reports a source file larger than MaxSourceSize.
	ErrSourceTooLarge = source.ErrTooLarge

	// ErrDOMTooDeep means the template nests past MaxDOMDepth.
	ErrDOMTooDeep = source.ErrDOMTooDeep

	// ErrTooManyNodes means the template holds past MaxDOMNodes elements.
	ErrTooManyNodes = source.ErrTooManyNodes

	// ErrImageTooLarge means a data:image/* URL passed the size cap.
	ErrImageTooLarge = source.ErrImageTooLarge

	// ErrBadAsset means an asset name escapes its directory.
	ErrBadAsset = source.ErrBadAsset
)

// sourceHTML returns the template source and the FileInfo of the file it
// came from, when there is one. Setting HTML and File together is an error.
func sourceHTML(cfg Config) (string, os.FileInfo, error) {
	return source.Load(srcFiles, source.Text{
		File:      cfg.File,
		Inline:    cfg.HTML,
		Limit:     MaxSourceSize,
		BadSource: ErrBadSource,
	})
}

// sourceTheme returns the theme source and the FileInfo of its file.
// Setting Theme and ThemeFile together is an error.
func sourceTheme(cfg Config) (string, os.FileInfo, error) {
	return source.Load(srcFiles, source.Text{
		File:      cfg.ThemeFile,
		Inline:    cfg.Theme,
		Limit:     MaxSourceSize,
		BadSource: ErrBadSource,
	})
}

// sourceTooLarge builds the wrapped oversize error. It wraps both
// ErrBadSource and ErrSourceTooLarge so errors.Is matches either.
func sourceTooLarge(path string, size int64) error {
	return source.Oversize(path, size, MaxSourceSize, ErrBadSource)
}

func checkDOMLimits(html string) error {
	return source.CheckDOM(html, MaxDOMDepth, MaxDOMNodes)
}

func domComment(s string, i int) (int, bool) { return source.SkipMeta(s, i) }

func domNameByte(c byte) bool { return source.NameByte(c) }

func isVoidTag(name string) bool { return source.Void(name) }

func checkImageSrc(src string) error {
	return source.ImageSrc(src, MaxDataImageBytes, ErrUnsupportedLinkScheme)
}

// AssetResolver serves img, font, and CSS bytes only from fsys.
// Names are slash paths rooted at the asset dir; absolute paths,
// backslashes, drive letters, null bytes, and dot-dot escapes fail.
func AssetResolver(fsys fs.FS) func(src string) ([]byte, error) {
	return source.Resolver(fsys)
}

// SafeFuncs returns helpers for page templates. Every helper returns
// a plain string so html/template keeps escaping it. No helper may
// return template.HTML, template.JS, template.URL, or template.CSS.
// A typed-safe value needs a reviewed constructor in the app package
// with a comment naming the safe inputs.
func SafeFuncs() template.FuncMap { return source.SafeFuncs() }

// Validate rejects illegal Config states early: HTML with File, Theme
// with ThemeFile, and unusable sizes. New calls it before reading
// files so conflicts fail at construction instead of first draw.
func (c Config) Validate() error {
	return source.Validate(source.Check{
		File:      c.File,
		HTML:      c.HTML,
		ThemeFile: c.ThemeFile,
		Theme:     c.Theme,
		Width:     c.Width,
		Height:    c.Height,
		MinWidth:  c.MinWidth,
		MinHeight: c.MinHeight,
		MaxWidth:  c.MaxWidth,
		MaxHeight: c.MaxHeight,
	}, ErrBadSource, ErrBadSize)
}

// mins returns the effective minimums; zero means 1.
func (c Config) mins() (int, int) { return source.Mins(c.MinWidth, c.MinHeight) }

type memSource struct{ mem *source.Mem }

func newMemSource() *memSource { return &memSource{mem: source.NewMem()} }

func (m *memSource) write(p string, b []byte) { m.mem.Write(p, b) }

func (m *memSource) Stat(p string) (os.FileInfo, error) { return m.mem.Stat(p) }

func (m *memSource) ReadFile(p string) ([]byte, error) { return m.mem.ReadFile(p) }

func installSourceReader(r source.Reader) {
	if r == nil {
		srcFiles = source.OS{}
		return
	}
	srcFiles = r
}

// ExecutedSource returns the last executed template source with live form
// values merged in. It is the same source the PNG path replays, so a browser
// shell can render the actual tags instead of a picture. Empty before the
// first Redraw.
func (p *Page) ExecutedSource() string {
	if p == nil {
		return ""
	}

	return p.source
}

// ThemeSource returns the raw theme stylesheet SetTheme stored, or "" when
// no theme is set.
func (p *Page) ThemeSource() string {
	if p == nil {
		return ""
	}

	return p.themeSrc
}

// SelectOption is one choice of a select control for the web preview.
type SelectOption = host.SelectOption

// PlaceCaret collapses the caret of the focused field to a rune offset.
// The web preview sends the browser caret with every edit so the server
// applies backspaces and typing where the browser put them. Out-of-range
// offsets clamp. It does nothing when id is not focused. It does not draw;
// the edit that follows it does.
func (p *Page) PlaceCaret(id string, caret int) {
	if p == nil {
		return
	}
	fs := formState(p)
	if fs == nil || fs.FocusID != id {
		return
	}
	c, ok := focusedEditable(p)
	if !ok {
		return
	}
	p.tickMutated.Store(true)
	fs.SetCaret(caret, runeLen(c.Value), false)
}

// SelectControls returns the option lists of the live select controls, keyed
// by control id. The executed source carries only the current label, so the
// browser shell needs this to fill real dropdowns. Nil when the page has no
// form yet.
func (p *Page) SelectControls() map[string][]SelectOption {
	if p == nil {
		return nil
	}
	fs := formState(p)
	if fs == nil {
		return nil
	}
	out := make(map[string][]SelectOption)
	for id, c := range fs.ByID {
		if c.Tag != "select" {
			continue
		}
		opts := make([]SelectOption, 0, len(c.Options))
		for _, o := range c.Options {
			opts = append(opts, SelectOption{
				Value:    o.Value,
				Label:    o.Label,
				Selected: o.Selected,
				Disabled: o.Disabled,
			})
		}
		out[id] = opts
	}

	return out
}

// srcFiles reads template and theme files. installSourceReader swaps it
// in tests. The gochecknoglobals allowlist names this file.
var srcFiles source.Reader = source.OS{}

// bindTarget returns the settable field name on the struct p.data points at.
// A missing or unexported field is false.
func bindTarget(p *Page, name string) (reflect.Value, bool) {
	if p == nil || name == "" || p.data == nil {
		return reflect.Value{}, false
	}

	v := reflect.ValueOf(p.data)
	if v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return reflect.Value{}, false
	}

	f := v.Elem().FieldByName(name)
	if !f.IsValid() || !f.CanSet() {
		return reflect.Value{}, false
	}

	return f, true
}

// bindKind is 'b' for a checkbox bool field, 's' for a string field, or 0.
// Text-like inputs, textareas, selects, and radios are 's'.
func bindKind(c Control) byte {
	if c.Type == "checkbox" {
		return 'b'
	}
	if c.Tag == "textarea" || c.Tag == "select" || c.Type == "radio" || textLike(c.Type) {
		return 's'
	}

	return 0
}

// bindRead copies the bound field into c and reports whether c is bound.
// A radio is checked when the string field holds the control value.
func bindRead(p *Page, c Control) (Control, bool) {
	f, ok := bindTarget(p, c.Bind)
	if !ok {
		return c, false
	}

	switch bindKind(c) {
	case 'b':
		if f.Kind() != reflect.Bool {
			return c, false
		}
		c.Checked = f.Bool()
	case 's':
		if f.Kind() != reflect.String {
			return c, false
		}
		if c.Type == "radio" {
			c.Checked = f.String() == c.Value
		} else {
			c.Value = f.String()
		}
	default:
		return c, false
	}

	return c, true
}

// bindWrite copies c into the bound field and reports whether c is bound.
// An unchecked radio writes nothing.
func bindWrite(p *Page, c Control) bool {
	// dataMu guards p.data, which bindTarget reads and the setters mutate,
	// and the formDirty flag set below. No caller holds it.
	if p == nil {
		return false
	}
	p.dataMu.Lock()
	defer p.dataMu.Unlock()

	f, ok := bindTarget(p, c.Bind)
	if !ok {
		return false
	}

	switch bindKind(c) {
	case 'b':
		if f.Kind() != reflect.Bool {
			return false
		}
		f.SetBool(c.Checked)
		p.formDirty = true
	case 's':
		if f.Kind() != reflect.String {
			return false
		}
		if c.Type != "radio" || c.Checked {
			f.SetString(c.Value)
			p.formDirty = true
		}
	default:
		return false
	}

	return true
}
