package snip

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	osclip "github.com/chinmay-sawant/ownframe/internal/clipboard"
)

// save writes the cropped preview as snip-YYYYMMDD-HHMMSS.png. A
// failed write keeps the preview and names the error in the status.
func (a *App) save() {
	if len(a.crop) == 0 {
		a.view.Status = "nothing to save"
		return
	}
	dir := a.saveDir
	if dir == "" {
		var err error
		if dir, err = defaultSaveDir(); err != nil {
			a.view.Status = "save failed: " + err.Error()
			return
		}
	}
	if err := os.MkdirAll(dir, snipDirPerm); err != nil {
		a.view.Status = "save failed: " + err.Error()
		return
	}
	path := filepath.Join(dir, "snip-"+time.Now().Format("20060102-150405")+".png")
	if err := os.WriteFile(path, a.crop, snipFilePerm); err != nil {
		a.view.Status = "save failed: " + err.Error()
		return
	}
	a.view.Status = "saved " + path
}

// defaultSaveDir is UserConfigDir/ownframe/snipping-tool.
func defaultSaveDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		return "", errors.New("snip: no config dir")
	}
	return filepath.Join(base, "ownframe", "snipping-tool"), nil
}

// copy sends the cropped PNG to the OS clipboard. A web preview has
// none, so copy reports the size in the status line instead.
func (a *App) copy() {
	if len(a.crop) == 0 {
		a.view.Status = "nothing to copy"
		return
	}
	if a.web {
		a.view.Status = "capture " + a.view.Size + "; no clipboard in web preview"
		return
	}
	osclip.Write(string(a.crop))
	a.view.Status = "copied " + a.view.Size
}
