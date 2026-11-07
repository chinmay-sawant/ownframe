package chadcn

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	attachmentMaxProgress  = 100
	attachmentMaxExtension = 4
)

// AttachmentOptions describes a file tile. Progress is a percentage from 0 to
// 100, and zero hides the bar. Action becomes the tile's data-action when set.
// Icon replaces the derived extension label. ID is an optional stable host id;
// the name, meta, and progress parts derive their ids from it so a same-shape
// meta edit patches instead of reparsing.
type AttachmentOptions struct {
	ID       string
	Name     string
	Meta     string
	Progress int
	Icon     *ui.Element
	Action   string
}

// Attachment builds a file tile: an icon slot, the file name and metadata, an
// optional upload bar, and a dismiss mark when the tile carries an action.
func Attachment(options AttachmentOptions) *ui.Element {
	actionClass := ""
	if options.Action != "" {
		actionClass = "of-attachment--action"
	}
	tile := ui.Div().Class(classes("of-attachment", actionClass))
	if options.ID != "" {
		tile.ID(options.ID)
	}
	if options.Action != "" {
		tile.Attr("data-action", options.Action)
	}

	icon := ui.Div().Class("of-attachment-icon")
	if id := partID(options.ID, "-icon"); id != "" {
		icon.ID(id)
	}
	if options.Icon != nil {
		icon.Children(options.Icon)
	} else {
		icon.Text(attachmentExtension(options.Name))
	}
	tile.Children(icon)

	body := ui.Div().Class("of-attachment-body")
	if options.Name != "" {
		name := ui.Span().Class("of-attachment-name").Text(options.Name)
		if id := partID(options.ID, "-name"); id != "" {
			name.ID(id)
		}
		body.Children(name)
	}
	if options.Meta != "" {
		meta := ui.Span().Class("of-attachment-meta").Text(options.Meta)
		if id := partID(options.ID, "-meta"); id != "" {
			meta.ID(id)
		}
		body.Children(meta)
	}
	if progress := min(max(options.Progress, 0), attachmentMaxProgress); progress > 0 {
		bar := ui.Div().Class("of-attachment-progress").Attr("role", "progressbar").
			Attr("aria-label", options.Name).Attr("aria-valuemin", "0").
			Attr("aria-valuemax", "100").Attr("aria-valuenow", strconv.Itoa(progress))
		if id := partID(options.ID, "-progress"); id != "" {
			bar.ID(id)
		}
		fill := ui.Div().Class("of-attachment-progress-fill").Width(ui.Percent(float64(progress)))
		if id := partID(options.ID, "-progress-fill"); id != "" {
			fill.ID(id)
		}
		body.Children(bar.Children(fill))
	}
	tile.Children(body)

	if options.Action != "" {
		tile.Children(ui.Span().Class("of-attachment-remove").Attr("aria-hidden", "true").Text("×"))
	}
	return tile
}

// attachmentExtension turns the file suffix into a short icon label.
func attachmentExtension(name string) string {
	base := name
	if index := strings.LastIndexAny(base, `/\`); index >= 0 {
		base = base[index+1:]
	}
	dot := strings.LastIndexByte(base, '.')
	if dot < 0 || dot == len(base)-1 {
		return "FILE"
	}
	extension := strings.ToUpper(base[dot+1:])
	if len(extension) > attachmentMaxExtension {
		return "FILE"
	}
	return extension
}
