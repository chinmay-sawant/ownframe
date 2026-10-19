package csspreview

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCombinedFixtureDeclaresAll43CatalogProperties(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "css-preview", "styles.css"))
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	scan := bufio.NewScanner(strings.NewReader(string(data)))
	for scan.Scan() {
		line := strings.TrimSpace(scan.Text())
		for _, property := range catalogProperties {
			if strings.HasPrefix(line, property+":") {
				seen[property] = true
			}
		}
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	var missing []string
	for _, property := range catalogProperties {
		if !seen[property] {
			missing = append(missing, property)
		}
	}
	if len(catalogProperties) != 43 || len(missing) != 0 {
		t.Fatalf("catalog declarations: got %d names, missing %v", len(catalogProperties), missing)
	}
}
