package storage

import (
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// memoryPath is the SQLite in-memory database name.
const memoryPath = ":memory:"

// fileDSN builds a file: URI that survives spaces, Unicode, and URI
// punctuation in the path. The URI form is used for every file so a path
// containing "?" or "#" is not read as a query or fragment.
func fileDSN(path string, windows bool) string {
	if path == memoryPath || strings.HasPrefix(path, "file:") {
		return path
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}

	host, slashPath := splitPath(filepath.ToSlash(abs), windows)

	return fileURI(host, slashPath)
}

// splitPath separates a UNC host from a slash path. On Windows a drive path
// gets a leading slash so it lands in the URI path, not the host.
func splitPath(p string, windows bool) (host, path string) {
	if strings.HasPrefix(p, "//") {
		rest := strings.TrimPrefix(p, "//")
		if i := strings.IndexByte(rest, '/'); i >= 0 {
			return rest[:i], rest[i:]
		}

		return rest, ""
	}

	if windows && isDrivePath(p) {
		return "", "/" + p
	}

	return "", p
}

// isDrivePath reports a Windows drive path such as C:/Users/x. It does not
// use filepath.VolumeName, which is host specific.
func isDrivePath(p string) bool {
	if len(p) < 3 || p[1] != ':' || p[2] != '/' {
		return false
	}

	c := p[0]

	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

// fileURI assembles the file: URI from its parts.
func fileURI(host, path string) string {
	return (&url.URL{Scheme: "file", Host: host, Path: path}).String()
}

// isWindows reports the host platform for the DSN shape.
func isWindows() bool { return runtime.GOOS == "windows" }

// mkdirAll creates a data directory.
func mkdirAll(dir string) error {
	return os.MkdirAll(dir, storeDirPerm)
}

// joinDir joins a data directory and a file name.
func joinDir(dir, name string) string {
	return filepath.Join(dir, name)
}
