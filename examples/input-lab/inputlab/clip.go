package inputlab

import (
	"context"
	"strings"

	osclip "github.com/chinmay-sawant/ownframe/internal/clipboard"
)

// clipClick runs the clipboard buttons.
func (a *App) clipClick(ctx context.Context, id string) error {
	switch id {
	case "c-copy":
		return a.copyButton(ctx)
	case "c-cut":
		return a.cutButton(ctx)
	case "c-paste":
		return a.pasteClipboard(ctx)
	case "c-selectall":
		return a.selectAllClip(ctx)
	case "c-pastego":
		return a.pasteGo(ctx)
	case "c-undo":
		return a.undoClip()
	case "c-redo":
		return a.redoClip()
	}
	return nil
}

func (a *App) copyButton(ctx context.Context) error {
	a.focusLast()
	text, ok, err := a.page.Copy(ctx)
	if err != nil {
		return err
	}
	if !ok || text == "" {
		a.setClip("nothing selected")
		return nil
	}
	osclip.Write(text)
	a.setClip("copied: " + text)
	return nil
}

func (a *App) cutButton(ctx context.Context) error {
	a.focusLast()
	text, ok, err := a.page.Cut(ctx)
	if err != nil {
		return err
	}
	if !ok || text == "" {
		a.setClip("nothing selected")
		return nil
	}
	osclip.Write(text)
	a.setClip("cut: " + text)
	return nil
}

// pasteGo inserts the fixed Go text in the last field.
func (a *App) pasteGo(ctx context.Context) error {
	if !a.focusLast() {
		return nil
	}
	if err := a.page.Paste(ctx, pasteFromGo); err != nil {
		return err
	}
	a.setClip("pasted: " + pasteFromGo)
	return nil
}

// pasteClipboard inserts OS text with newlines dropped.
func (a *App) pasteClipboard(ctx context.Context) error {
	if !a.focusLast() {
		a.setClip("nothing to paste into")
		return nil
	}
	raw, _ := osclip.Read()
	text := strings.ReplaceAll(raw, "\r", "")
	text = strings.ReplaceAll(text, "\n", "")
	if text == "" {
		a.setClip("clipboard is empty")
		return nil
	}
	if err := a.page.Paste(ctx, text); err != nil {
		return err
	}
	a.setClip("pasted: " + text)
	return nil
}

const pasteFromGo = "from Go"

func (a *App) setClip(text string) {
	a.view.Status = text
	a.page.SetData(&a.view)
}
