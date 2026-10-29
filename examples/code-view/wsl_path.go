package main

import (
	"strings"
)

// parseWSLRoot recognizes WSL network paths of the form
// \\wsl$\<distro>\... and \\wsl.localhost\<distro>\... and returns the
// distro name with the in-WSL POSIX path. The host part matches
// case-insensitively and either slash separator is accepted, including
// the //wsl$/... forward-slash form. Backslashes in the subpath are
// converted to forward slashes and a leading slash is ensured, so a bare
// \\wsl$\<distro> with no subpath maps to "/". Anything else reports
// ok=false.
func parseWSLRoot(root string) (distro, wslPath string, ok bool) {
	if len(root) < 2 || !isWSLSep(root[0]) || !isWSLSep(root[1]) {
		return "", "", false
	}
	rest := root[2:]
	hostEnd := indexWSLSep(rest)
	if hostEnd < 0 {
		return "", "", false
	}
	host := strings.ToLower(rest[:hostEnd])
	if host != `wsl$` && host != "wsl.localhost" {
		return "", "", false
	}
	afterHost := rest[hostEnd+1:]
	distroEnd := indexWSLSep(afterHost)
	if distroEnd < 0 {
		distro = afterHost
		if distro == "" {
			return "", "", false
		}
		return distro, "/", true
	}
	distro = afterHost[:distroEnd]
	if distro == "" {
		return "", "", false
	}
	rawSub := afterHost[distroEnd+1:]
	rawSub = strings.ReplaceAll(rawSub, `\`, "/")
	parts := strings.Split(rawSub, "/")
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		kept = append(kept, part)
	}
	if len(kept) == 0 {
		return distro, "/", true
	}
	return distro, "/" + strings.Join(kept, "/"), true
}

// drivePrefixLen is the shortest drive-letter path: `C:\`.
const drivePrefixLen = 3

// windowsToWSLPath maps drive-letter paths such as C:\x or C:/x (any
// letter, case-insensitive) to /mnt/<lower-letter>/x with forward slashes.
// Anything else, including UNC non-WSL paths, relative paths, and the empty
// string, reports false.
func windowsToWSLPath(p string) (string, bool) {
	if len(p) < drivePrefixLen {
		return "", false
	}
	drive := p[0]
	if !isDriveLetter(drive) {
		return "", false
	}
	if p[1] != ':' || !isWSLSep(p[2]) {
		return "", false
	}
	rest := strings.ReplaceAll(p[2:], `\`, "/")
	parts := strings.Split(rest, "/")
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		kept = append(kept, part)
	}
	mnt := "/mnt/" + strings.ToLower(string([]byte{drive}))
	if len(kept) == 0 {
		return mnt, true
	}
	return mnt + "/" + strings.Join(kept, "/"), true
}

func isWSLSep(c byte) bool {
	return c == '\\' || c == '/'
}

// wslMntToWindows maps a /mnt/<drive>/... path seen from Windows back to a
// drive-letter path such as C:\.... Anything else reports false.
func wslMntToWindows(p string) (string, bool) {
	lowered := strings.ToLower(p)
	if lowered != "/mnt" && !strings.HasPrefix(lowered, "/mnt/") {
		return "", false
	}

	rest := p[len("/mnt"):]
	rest = strings.TrimPrefix(rest, "/")
	if rest == "" {
		return "", false
	}

	drive := rest[0]
	if !isDriveLetter(drive) || (len(rest) > 1 && rest[1] != '/') {
		return "", false
	}

	tail := strings.TrimPrefix(rest[1:], "/")

	out := string([]byte{drive, ':'}) + `\`
	if tail != "" {
		out += strings.ReplaceAll(tail, "/", `\`)
	}

	return strings.ToUpper(out[:1]) + out[1:], true
}

func indexWSLSep(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' || s[i] == '/' {
			return i
		}
	}
	return -1
}
