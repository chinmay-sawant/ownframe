package print

import (
	"fmt"
	"os"
)

// debugf writes to stderr when OWNFRAME_PRINT_DEBUG is not empty.
func debugf(format string, args ...any) {
	if os.Getenv("OWNFRAME_PRINT_DEBUG") == "" && os.Getenv("GPUI_PRINT_DEBUG") == "" {
		return
	}

	fmt.Fprintf(os.Stderr, "ownframe print: "+format+"\n", args...)
}
