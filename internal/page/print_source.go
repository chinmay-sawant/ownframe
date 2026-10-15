package page

import (
	"context"
	"html"
	"strings"
)

// pdfSource returns the source a PDF would render: the last template output,
// or the current template when nothing has drawn yet, with the theme in a
// style element. blinkless does not write the file.
func (p *Page) pdfSource(ctx context.Context) (string, error) {
	if err := useContext(ctx); err != nil {
		return "", err
	}

	source := p.source
	if source == "" {
		var body strings.Builder
		if err := p.tpl.Execute(&body, p.data); err != nil {
			return "", err
		}

		source = p.syncForm(body.String())
	}

	return injectTheme(source, p.themeSrc), nil
}

// injectTheme puts theme in a style element in the head, before </head> when
// the source has one, and in front of the source otherwise.
func injectTheme(source, theme string) string {
	if strings.TrimSpace(theme) == "" {
		return source
	}

	style := "<style>" + html.EscapeString(theme) + "</style>"
	lower := strings.ToLower(source)

	if at := strings.Index(lower, "</head>"); at >= 0 {
		return source[:at] + style + source[at:]
	}

	if at := strings.Index(lower, "<head>"); at >= 0 {
		at += len("<head>")

		return source[:at] + style + source[at:]
	}

	return style + source
}
