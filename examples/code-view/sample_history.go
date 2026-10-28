package main

// sampleCommitSpec groups the checked-in sample diff into a small history.
type sampleCommitSpec struct {
	Hash    string
	Subject string
	Body    string
	Author  string
	Date    string
	Paths   []string
}

// sampleCommits is the fixed three-commit series for the bundled workspace.
func sampleCommits() []sampleCommitSpec {
	return []sampleCommitSpec{
		{
			Hash:    "0a891918d3fa29a434ac0eb41df0fd7068e71c9c",
			Subject: "Retry probes and persist their history",
			Body:    "Add a retry loop with a bounded client and keep the last day of probe results in the local store.",
			Author:  "Mira Chen",
			Date:    "2026-10-24",
			Paths:   []string{"backend/probe.go", "backend/store.go", "config/services.yaml"},
		},
		{
			Hash:    "621400e7efb80a9aaa3a261f2ea5c3f4d2be917f",
			Subject: "Document deployment and add accent themes",
			Body:    "Document the container flow and add the theme variables used by the dashboard.",
			Author:  "Devon Park",
			Date:    "2026-10-21",
			Paths:   []string{"docs/deployment.md", "web/theme.css"},
		},
		{
			Hash:    "094ad38196af2edd36ad8269825ae299b2b586ad",
			Subject: "Replace the legacy report",
			Body:    "Move the report to the Python client and remove the old shell wrapper.",
			Author:  "Sana Iqbal",
			Date:    "2026-10-17",
			Paths:   []string{"tools/legacy.py", "tools/report.py"},
		},
	}
}

// sampleCommitLog builds the sample rail and its preloaded file patches.
func sampleCommitLog() (commitLog, map[string]*changedFile) {
	changes := sampleChangeSet()
	files := make(map[string]*changedFile, len(changes.Files))

	for i := range changes.Files {
		files[changes.Files[i].Path] = &changes.Files[i]
	}

	details := make(map[string]*changedFile)
	specs := sampleCommits()
	commits := make([]gitCommit, 0, len(specs))

	for _, spec := range specs {
		commit := gitCommit{
			Hash:    spec.Hash,
			Short:   spec.Hash[:7],
			Subject: spec.Subject,
			Body:    spec.Body,
			Author:  spec.Author,
			Date:    spec.Date,
		}

		for _, path := range spec.Paths {
			file := files[path]
			if file == nil {
				continue
			}

			commit.Files = append(commit.Files, commitFile{
				Path:    file.Path,
				OldPath: file.OldPath,
				Status:  file.Status,
				Add:     file.Add,
				Del:     file.Del,
				Binary:  file.Binary,
			})
			commit.Add += file.Add
			commit.Del += file.Del
			details[commitDetailKey(spec.Hash, path)] = file
		}

		commits = append(commits, commit)
	}

	for i := range commits {
		if i+1 < len(commits) {
			commits[i].Parents = []string{commits[i+1].Hash}
		}
	}

	log := commitLog{
		Branch:  "feat/probe-retries",
		Head:    commits[0].Hash,
		Commits: commits,
		Note:    "Bundled sample history: a fixed three-commit series, not a live Git repository.",
		Sample:  true,
	}

	return log, details
}
