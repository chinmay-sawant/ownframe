package transfer

import (
	"strings"
)

// PartialSuffix is appended to the destination name for the in-progress file.
const PartialSuffix = ".ownframe-part"

// PartialPath returns the partial-file path for a destination.
func PartialPath(dest string) string { return dest + PartialSuffix }

// reserved are Windows device names that cannot be used as file names.
var reserved = map[string]bool{
	"con": true, "prn": true, "aux": true, "nul": true,
	"com1": true, "com2": true, "com3": true, "com4": true,
	"com5": true, "com6": true, "com7": true, "com8": true, "com9": true,
	"lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true,
	"lpt5": true, "lpt6": true, "lpt7": true, "lpt8": true, "lpt9": true,
}

// SafeName derives a bare file name from an untrusted URL path base. The
// result has no separators or control characters, is never a Windows device
// name, and never ends in a dot or space. It falls back to "download".
func SafeName(raw string) string {
	name := pathBase(raw)
	name = strings.Map(safeRune, name)
	name = strings.TrimRight(name, ". ")
	name = trimWindows(name)
	if name == "" || name == "." || name == ".." {
		name = "download"
	}

	return clipName(name)
}

// pathBase takes the last path element of a URL-ish string. Query and
// fragment are dropped, and a backslash counts as a separator.
func pathBase(raw string) string {
	raw = strings.SplitN(raw, "?", splitParts)[0]
	raw = strings.SplitN(raw, "#", splitParts)[0]
	raw = strings.ReplaceAll(raw, "\\", "/")

	parts := strings.Split(raw, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}

	return ""
}

// safeRune drops control characters and replaces characters Windows
// forbids. Unicode letters pass through.
func safeRune(r rune) rune {
	if r < 0x20 || r == 0x7f {
		return -1
	}

	if strings.ContainsRune(`<>:"|?*/\`, r) {
		return '_'
	}

	return r
}
