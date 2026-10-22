package cat

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/chinmay-sawant/ownframe/examples/desktop-cat/assets"
)

// Expressions maps emotion names and individual pose names to image numbers.
func Expressions() map[string]int {
	files, _ := fs.Glob(assets.Cats, "cat_images/*.png")
	result := make(map[string]int)
	for i, file := range files {
		name := strings.TrimSuffix(filepath.Base(file)[3:], ".png")
		result[name] = i + 1
		emotion := strings.SplitN(name, "-", splitParts)[0]
		if _, found := result[emotion]; !found {
			result[emotion] = i + 1
		}
	}
	return result
}

func expressionIndex(name string) (int, error) {
	if index, ok := Expressions()[name]; ok {
		return index - 1, nil
	}
	return 0, fmt.Errorf("unknown expression %q; see GET /expressions", name)
}
