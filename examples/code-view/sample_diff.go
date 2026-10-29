package main

import _ "embed"

// sampleDiff is a checked-in unified diff for the bundled workspace. It is
// the default Source Control view so the example has a rich diff without a
// repository. With -root, live git replaces it.
//
//go:embed fixtures/sample.diff
var sampleDiff []byte

// sampleStagedCount is the number of bundled sample files shown as staged
// so the fixture demonstrates both the staged and unstaged sections.
const sampleStagedCount = 2

// sampleChangeSet parses the bundled diff. The first two files show as
// staged so the sample demonstrates the staged and unstaged sections.
func sampleChangeSet() changeSet {
	files := parseDiff(string(sampleDiff))
	staged := make([]changedFile, 0, len(files))
	unstaged := make([]changedFile, 0, len(files))

	for i, file := range files {
		if i < sampleStagedCount {
			staged = append(staged, file)
		} else {
			unstaged = append(unstaged, file)
		}
	}

	return changeSet{
		Files:    files,
		Staged:   staged,
		Unstaged: unstaged,
		Branch:   "feat/probe-retries",
		Sample:   true,
		Note:     "Bundled sample. Run with -root to read a real repository.",
	}
}
