package displaypatch

import "os"

// DebugPath returns the configured debug output path, if any.
func DebugPath() string {
	return os.Getenv("OWNFRAME_DISPLAY_PATCH_DEBUG")
}

// Enabled reports whether patch debugging is on.
func (d *DebugLog) EnabledEnv() bool {
	return os.Getenv("OWNFRAME_DISPLAY_PATCH_DEBUG") != ""
}
