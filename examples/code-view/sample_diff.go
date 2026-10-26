package main

import _ "embed"

// sampleDiff is a checked-in unified diff for the bundled workspace. It is
// the default Source Control view so the example has a rich diff without a
// repository. With -root, live git replaces it.
//
//go:embed fixtures/sample.diff
var sampleDiff []byte

// sampleChangeSet parses the bundled diff.
func sampleChangeSet() changeSet {
	return changeSet{
		Files:  parseDiff(string(sampleDiff)),
		Branch: "feat/probe-retries",
		Sample: true,
		Note:   "Bundled sample. Run with -root to read a real repository.",
	}
}
