package main

import (
	"testing"
)

func TestParseWSLRoot(t *testing.T) {
	tests := []struct {
		name    string
		root    string
		distro  string
		wslPath string
		ok      bool
	}{
		{
			name:    "wsl dollar basic",
			root:    `\\wsl$\Ubuntu\home\user`,
			distro:  "Ubuntu",
			wslPath: "/home/user",
			ok:      true,
		},
		{
			name:    "wsl localhost basic",
			root:    `\\wsl.localhost\Ubuntu\home\user`,
			distro:  "Ubuntu",
			wslPath: "/home/user",
			ok:      true,
		},
		{
			name:    "forward slash form",
			root:    `//wsl$/Ubuntu/home/user`,
			distro:  "Ubuntu",
			wslPath: "/home/user",
			ok:      true,
		},
		{
			name:    "forward slash localhost",
			root:    `//wsl.localhost/Ubuntu/home/user`,
			distro:  "Ubuntu",
			wslPath: "/home/user",
			ok:      true,
		},
		{
			name:    "host case insensitive dollar",
			root:    `\\WSL$\Ubuntu\home\user`,
			distro:  "Ubuntu",
			wslPath: "/home/user",
			ok:      true,
		},
		{
			name:    "host case insensitive localhost",
			root:    `\\WSL.LOCALHOST\Ubuntu\home\user`,
			distro:  "Ubuntu",
			wslPath: "/home/user",
			ok:      true,
		},
		{
			name:    "mixed separators",
			root:    `\\wsl$\Ubuntu/home\user/project`,
			distro:  "Ubuntu",
			wslPath: "/home/user/project",
			ok:      true,
		},
		{
			name:    "nested path",
			root:    `\\wsl$\Debian\a\b\c\d`,
			distro:  "Debian",
			wslPath: "/a/b/c/d",
			ok:      true,
		},
		{
			name:    "spaces in path",
			root:    `\\wsl$\Ubuntu\home\my user\my docs`,
			distro:  "Ubuntu",
			wslPath: "/home/my user/my docs",
			ok:      true,
		},
		{
			name:    "trailing backslash",
			root:    `\\wsl$\Ubuntu\home\user\`,
			distro:  "Ubuntu",
			wslPath: "/home/user",
			ok:      true,
		},
		{
			name:    "trailing forward slash",
			root:    `//wsl$/Ubuntu/home/user/`,
			distro:  "Ubuntu",
			wslPath: "/home/user",
			ok:      true,
		},
		{
			name:    "bare distro backslash",
			root:    `\\wsl$\Ubuntu`,
			distro:  "Ubuntu",
			wslPath: "/",
			ok:      true,
		},
		{
			name:    "bare distro localhost",
			root:    `\\wsl.localhost\Ubuntu`,
			distro:  "Ubuntu",
			wslPath: "/",
			ok:      true,
		},
		{
			name:    "bare distro forward slash",
			root:    `//wsl$/Ubuntu`,
			distro:  "Ubuntu",
			wslPath: "/",
			ok:      true,
		},
		{
			name:    "bare distro trailing sep",
			root:    `\\wsl$\Ubuntu\`,
			distro:  "Ubuntu",
			wslPath: "/",
			ok:      true,
		},
		{
			name: "plain UNC share",
			root: `\\server\share\dir`,
			ok:   false,
		},
		{
			name: "missing distro",
			root: `\\wsl$\`,
			ok:   false,
		},
		{
			name: "host only",
			root: `\\wsl$`,
			ok:   false,
		},
		{
			name: "single leading slash",
			root: `\wsl$\Ubuntu\home`,
			ok:   false,
		},
		{
			name: "local path",
			root: `C:\Users\name`,
			ok:   false,
		},
		{
			name: "posix path",
			root: `/home/user`,
			ok:   false,
		},
		{
			name: "empty",
			root: ``,
			ok:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			distro, wslPath, ok := parseWSLRoot(tc.root)
			if ok != tc.ok {
				t.Fatalf("parseWSLRoot(%q) ok = %v, want %v", tc.root, ok, tc.ok)
			}
			if !tc.ok {
				return
			}
			if distro != tc.distro {
				t.Errorf("parseWSLRoot(%q) distro = %q, want %q", tc.root, distro, tc.distro)
			}
			if wslPath != tc.wslPath {
				t.Errorf("parseWSLRoot(%q) wslPath = %q, want %q", tc.root, wslPath, tc.wslPath)
			}
		})
	}
}

func TestWindowsToWSLPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
		ok   bool
	}{
		{
			name: "upper backslash",
			path: `C:\Users\name`,
			want: "/mnt/c/Users/name",
			ok:   true,
		},
		{
			name: "lower backslash",
			path: `c:\Users\name`,
			want: "/mnt/c/Users/name",
			ok:   true,
		},
		{
			name: "upper forward slash",
			path: `C:/Users/name`,
			want: "/mnt/c/Users/name",
			ok:   true,
		},
		{
			name: "lower forward slash",
			path: `d:/work/proj`,
			want: "/mnt/d/work/proj",
			ok:   true,
		},
		{
			name: "other drive",
			path: `D:\data`,
			want: "/mnt/d/data",
			ok:   true,
		},
		{
			name: "drive root backslash",
			path: `C:\`,
			want: "/mnt/c",
			ok:   true,
		},
		{
			name: "drive root forward slash",
			path: `C:/`,
			want: "/mnt/c",
			ok:   true,
		},
		{
			name: "trailing separator trimmed",
			path: `C:\Users\name\`,
			want: "/mnt/c/Users/name",
			ok:   true,
		},
		{
			name: "spaces",
			path: `C:\My Docs\file name.txt`,
			want: "/mnt/c/My Docs/file name.txt",
			ok:   true,
		},
		{
			name: "relative path",
			path: `Users\name`,
			ok:   false,
		},
		{
			name: "dot relative",
			path: `.\foo`,
			ok:   false,
		},
		{
			name: "drive relative no separator",
			path: `C:foo`,
			ok:   false,
		},
		{
			name: "bare drive",
			path: `C:`,
			ok:   false,
		},
		{
			name: "plain UNC share",
			path: `\\server\share\dir`,
			ok:   false,
		},
		{
			name: "wsl unc is not a drive",
			path: `\\wsl$\Ubuntu\home\user`,
			ok:   false,
		},
		{
			name: "posix path",
			path: `/home/user`,
			ok:   false,
		},
		{
			name: "empty",
			path: ``,
			ok:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := windowsToWSLPath(tc.path)
			if ok != tc.ok {
				t.Fatalf("windowsToWSLPath(%q) ok = %v, want %v", tc.path, ok, tc.ok)
			}
			if !tc.ok {
				return
			}
			if got != tc.want {
				t.Errorf("windowsToWSLPath(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}
