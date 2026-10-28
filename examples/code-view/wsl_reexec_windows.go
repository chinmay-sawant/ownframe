//go:build windows

package main

import (
	"errors"
	"flag"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

// wslEnabled backs the -wsl flag: allow the automatic handoff into WSL when
// the workspace root lives on a WSL network path.
var wslEnabled bool //nolint:gochecknoglobals // -wsl flag backing store, read once at startup

// wslBin backs the -wsl-bin flag: the Linux companion binary path inside
// WSL. Empty means the companion sits next to this executable.
var wslBin string //nolint:gochecknoglobals // -wsl-bin flag backing store, read once at startup

// wslExeSuffix is the Windows executable extension stripped from this
// binary's path to locate the Linux companion sitting next to it.
const wslExeSuffix = ".exe"

// wslUnknownExit is the fallback exit code when the WSL child was started
// but waiting on it reported an error carrying no exit code.
const wslUnknownExit = 1

// addWSLFlags registers the WSL handoff flags: -wsl (default true) allows a
// Windows launch to hand itself to the Linux companion inside WSL, and
// -wsl-bin names that companion binary inside WSL.
func addWSLFlags(fs *flag.FlagSet) {
	fs.BoolVar(&wslEnabled, "wsl", true, "allow automatic handoff into WSL when the workspace root is a WSL network path")
	fs.StringVar(&wslBin, "wsl-bin", "", "Linux companion binary path inside WSL; empty means the companion sits next to this executable")
}

// maybeReexecInWSL hands a Windows launch to the Linux companion inside WSL
// and exits with the child's exit code, reporting true. It triggers only
// when the workspace root (or, for a relative root, the current directory)
// is a WSL network path per parseWSLRoot; -wsl-bin alone never forces a
// handoff. Any failure before the child starts returns false so the app runs
// natively instead of stranding the user. The re-exec itself cannot be
// unit-tested headless; it needs a Windows host with WSL.
func maybeReexecInWSL() bool {
	if !wslEnabled {
		return false
	}
	if !wslAvailable() {
		return false
	}

	rawRoot, hasRoot := wslRootFromArgs(os.Args[1:])
	if !hasRoot {
		rawRoot = "."
	}

	var distro, linuxRoot string
	if parsedDistro, parsedPath, ok := parseWSLRoot(rawRoot); ok {
		distro, linuxRoot = parsedDistro, parsedPath
	} else {
		// An explicit absolute root chose native Windows; only a
		// relative root falls back to the current directory.
		if filepath.IsAbs(rawRoot) {
			return false
		}
		cwd, err := os.Getwd()
		if err != nil {
			return false
		}
		parsedDistro, parsedBase, ok := parseWSLRoot(cwd)
		if !ok {
			return false
		}
		distro, linuxRoot = parsedDistro, joinWSLRoot(parsedBase, rawRoot)
	}

	bin := wslBin
	if bin == "" {
		exe, err := os.Executable()
		if err != nil {
			return false
		}
		if strings.HasSuffix(strings.ToLower(exe), wslExeSuffix) {
			exe = exe[:len(exe)-len(wslExeSuffix)]
		}
		translated, ok := windowsToWSLPath(exe)
		if !ok {
			return false
		}
		bin = translated
	}

	launcher, err := exec.LookPath("wsl.exe")
	if err != nil {
		return false
	}
	childArgs := append([]string{"-d", distro, "--", bin}, wslRebuildArgs(os.Args[1:], linuxRoot)...)
	cmd := exec.Command(launcher, childArgs...) //nolint:gosec // launcher resolved by LookPath; distro, binary, and args are the operator's own launch
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	hideConsoleWindow(cmd)

	// Start separates launch failure (fall back to native) from a child
	// that ran: once started, propagate its exit code and never return, so
	// the app cannot run twice.
	if err := cmd.Start(); err != nil {
		return false
	}
	if err := cmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(wslUnknownExit)
	}
	os.Exit(0)

	return true
}

// wslRootFromArgs scans argv for -root/-root=/--root/--root= and reports the
// workspace root the child must open. ok is false when no root flag is
// present. Scanning stops at -- or the first positional, mirroring
// flag.Parse, so a -root after a positional is not treated as effective.
func wslRootFromArgs(argv []string) (root string, ok bool) {
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		if arg == "--" || arg == "-" || !strings.HasPrefix(arg, "-") {
			return "", false
		}
		if arg == "-root" || arg == "--root" {
			if i+1 >= len(argv) {
				return "", false
			}
			return argv[i+1], true
		}
		if value, found := strings.CutPrefix(arg, "-root="); found {
			return value, true
		}
		if value, found := strings.CutPrefix(arg, "--root="); found {
			return value, true
		}
	}

	return "", false
}

// joinWSLRoot resolves a relative -root against the WSL base directory the
// current directory translated to. "." keeps the base; anything else is
// appended with Windows separators converted and ".." cleaned lexically.
func joinWSLRoot(base, rel string) string {
	converted := strings.ReplaceAll(rel, `\`, "/")
	if converted == "" {
		return base
	}

	return path.Join(base, converted)
}

// wslRebuildArgs rebuilds argv for the Linux companion: -wsl and -wsl-bin
// exist only on the Windows side, so both dash spellings in = and space
// form are dropped, and -root is pointed at the translated WSL path. A -wsl
// in space form only consumes the next argument when it parses as a bool,
// so a bare -wsl never eats the following flag. Arguments after -- or the
// first positional pass through untouched. When no root flag is present, an
// explicit absolute -root is appended, because the child starts in its own
// working directory.
func wslRebuildArgs(argv []string, wslPath string) []string {
	out := make([]string, 0, len(argv)+2)
	seenRoot := false
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		if arg == "--" || arg == "-" || !strings.HasPrefix(arg, "-") {
			out = append(out, argv[i:]...)

			break
		}
		switch {
		case arg == "-wsl" || arg == "--wsl":
			if i+1 < len(argv) {
				if _, err := strconv.ParseBool(argv[i+1]); err == nil {
					i++
				}
			}
		case arg == "-wsl-bin" || arg == "--wsl-bin":
			i++
		case strings.HasPrefix(arg, "-wsl=") || strings.HasPrefix(arg, "--wsl=") ||
			strings.HasPrefix(arg, "-wsl-bin=") || strings.HasPrefix(arg, "--wsl-bin="):
			// Equals form: the value rides in this same argument.
		case arg == "-root" || arg == "--root":
			out = append(out, arg, wslPath)
			seenRoot = true
			i++
		case strings.HasPrefix(arg, "-root="):
			out = append(out, "-root="+wslPath)
			seenRoot = true
		case strings.HasPrefix(arg, "--root="):
			out = append(out, "--root="+wslPath)
			seenRoot = true
		default:
			out = append(out, arg)
		}
	}
	if !seenRoot {
		out = append(out, "-root", wslPath)
	}

	return out
}
