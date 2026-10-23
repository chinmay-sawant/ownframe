package page

import "github.com/chinmay-sawant/ownframe/internal/page/source"

// srcFiles reads template and theme files. installSourceReader swaps it
// in tests. The gochecknoglobals allowlist names this file.
var srcFiles source.Reader = source.OS{}
