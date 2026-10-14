package replay

import (
	"context"
	"fmt"
	"os"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

func checkPhotoMaskGPU() error {
	files := map[string]string{
		"mask.png":           "../../examples/css-masking/assets/alpha-half.png",
		"assets/flowers.jpg": "../../examples/css-masking/assets/flowers.jpg",
		"assets/coast.jpg":   "../../examples/css-masking/assets/coast.jpg",
	}
	images := make(map[string][]byte, len(files))
	for src, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		images[src] = data
	}
	state := render.State{Images: func(src string) ([]byte, error) { return images[src], nil }}
	masked, err := render.DisplayListState(context.Background(), photoMaskHTML(true, true), 200, 140, state)
	if err != nil {
		return err
	}
	if !render.Replayable(masked) {
		return fmt.Errorf("photo mask should replay")
	}
	plain, err := render.DisplayListState(context.Background(), photoMaskHTML(true, false), 200, 140, state)
	if err != nil {
		return err
	}
	backdrop, err := render.DisplayListState(context.Background(), photoMaskHTML(false, false), 200, 140, state)
	if err != nil {
		return err
	}
	got, front, back := gpuDisplayPixels(masked), gpuDisplayPixels(plain), gpuDisplayPixels(backdrop)
	if err := compareGPUPixel(got, back, 50, 50, "transparent photo mask"); err != nil {
		return err
	}
	if err := compareGPUPixel(got, front, 150, 50, "opaque photo mask"); err != nil {
		return err
	}
	return compareGPUPixel(got, back, 5, 5, "photo outside mask")
}
