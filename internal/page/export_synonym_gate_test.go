package page

import "testing"

func TestExportSynonymGate(t *testing.T) {
	got, err := exportedPageSynonyms()
	if err != nil {
		t.Fatal(err)
	}
	allow := synonymAllow()
	for name := range got {
		if !allow[name] {
			t.Errorf("new Page synonym %s; do not add aliases", name)
		}
	}
	for name := range allow {
		if !got[name] {
			t.Errorf("allowlisted synonym %s is gone; drop it", name)
		}
	}
}
