package page

import (
	"html/template"
	"io/fs"
	"os"

	"github.com/chinmay-sawant/ownframe/internal/page/source"
)

const (
	maxSourceSize = source.MaxBytes

	MaxSourceSize      = maxSourceSize
	MaxSourceBytes     = maxSourceSize
	MaxSourceFileSize  = maxSourceSize
	MaxSourceFileBytes = maxSourceSize
	SourceSizeLimit    = maxSourceSize
	SourceSizeMax      = maxSourceSize
	MaxSourceLength    = maxSourceSize
	MaxTemplateSize    = maxSourceSize
	MaxTemplateBytes   = maxSourceSize
	MaxSource          = maxSourceSize

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
		Limit:     maxSourceSize,
		BadSource: ErrBadSource,
	})
}

// sourceTheme returns the theme source and the FileInfo of its file.
// Setting Theme and ThemeFile together is an error.
func sourceTheme(cfg Config) (string, os.FileInfo, error) {
	return source.Load(srcFiles, source.Text{
		File:      cfg.ThemeFile,
		Inline:    cfg.Theme,
		Limit:     maxSourceSize,
		BadSource: ErrBadSource,
	})
}

// sourceTooLarge builds the wrapped oversize error. It wraps both
// ErrBadSource and ErrSourceTooLarge so errors.Is matches either.
func sourceTooLarge(path string, size int64) error {
	return source.Oversize(path, size, maxSourceSize, ErrBadSource)
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
