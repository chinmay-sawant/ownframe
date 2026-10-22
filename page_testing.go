package ownframe

// Testing is for test files. Example programs call Page methods directly.
// The methods here simulate input or read host probes that an app does not call.
type Testing struct{ page *Page }

// Testing returns helpers that belong in test files.
func (p *Page) Testing() Testing { return Testing{page: p} }
