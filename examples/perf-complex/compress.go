package main

import (
	"compress/gzip"
	"encoding/json"
	"os"
)

func writeLayout(path string, v any) error {
	// The caller builds path from the output dir, the allow-listed mode,
	// and a fixed suffix.
	f, err := os.Create(path) //nolint:gosec // reason: path is out dir plus allow-listed mode plus a fixed suffix, assembled by writeDump.
	if err != nil {
		return err
	}
	z := gzip.NewWriter(f)
	err = json.NewEncoder(z).Encode(v)
	ze := z.Close()
	fe := f.Close()
	if err != nil {
		return err
	}
	if ze != nil {
		return ze
	}
	return fe
}
