package snip

import (
	"context"
	"errors"
	"os/exec"
)

// linuxFullscreen tries import, gnome-screenshot, scrot, then grim.
func linuxFullscreen(ctx context.Context) ([]byte, error) {
	var errs []error
	if _, err := exec.LookPath("import"); err != nil {
		errs = append(errs, err)
	} else if raw, err := runOut(ctx, "import", "-window", "root", "png:-"); err != nil {
		errs = append(errs, err)
	} else {
		return raw, nil
	}
	tools := [][]string{{"gnome-screenshot", "-f"}, {"scrot"}, {"grim"}}
	for _, tool := range tools {
		if _, err := exec.LookPath(tool[0]); err != nil {
			errs = append(errs, err)
			continue
		}
		if raw, err := shotToFile(ctx, tool[0], tool[1:]...); err != nil {
			errs = append(errs, err)
		} else {
			return raw, nil
		}
	}
	return nil, errors.Join(errs...)
}
