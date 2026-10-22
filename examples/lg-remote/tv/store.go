package tv

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Store is the paired TV remembered on disk.
type Store struct {
	Host  string   `json:"host"`
	Key   string   `json:"key"`
	Model string   `json:"model"`
	MACs  []string `json:"macs"`
}

// Load reads a saved TV. A missing file is an empty store.
func Load(path string) Store {
	b, err := os.ReadFile(path) //nolint:gosec // reason: path is the caller-selected store file under the app config dir, not remote input.
	if err != nil {
		return Store{MACs: []string{}}
	}

	var s Store
	if json.Unmarshal(b, &s) != nil {
		return Store{MACs: []string{}}
	}

	if s.MACs == nil {
		s.MACs = []string{}
	}

	return s
}

// Save writes the store. The directory is created if needed.
func Save(path string, s Store) error {
	if s.MACs == nil {
		s.MACs = []string{}
	}

	if err := os.MkdirAll(filepath.Dir(path), storeDirPerm); err != nil {
		return err
	}

	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, storeFilePerm)
}
