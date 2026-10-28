package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestIsBrowseDriveRoot covers drive-root detection: only bare Windows
// drive roots count. The Unix root and anything below a drive root must
// not trigger the drives level.
func TestIsBrowseDriveRoot(t *testing.T) {
	cases := []struct {
		root string
		want bool
	}{
		{"C:", true},
		{`C:\`, true},
		{"C:/", true},
		{`d:\`, true},
		{"D:", true},
		{"", false},
		{"/", false},
		{"/mnt/c", false},
		{`/mnt`, false},
		{`C:\foo`, false},
		{"C:/foo", false},
		{"C:foo", false},
		{"CC:", false},
		{`1:\`, false},
		{`\\server\share`, false},
	}

	for _, tc := range cases {
		if got := isBrowseDriveRoot(tc.root); got != tc.want {
			t.Errorf("isBrowseDriveRoot(%q) = %v, want %v", tc.root, got, tc.want)
		}
	}
}

// TestBrowseUpFromDriveRootYieldsDrives checks that Up on a drive root
// steps out to the virtual drives level (the empty sentinel).
func TestBrowseUpFromDriveRootYieldsDrives(t *testing.T) {
	for _, root := range []string{`D:\`, "D:/", "D:", `c:\`} {
		next, ok := browseUpTarget(root)
		if !ok {
			t.Fatalf("browseUpTarget(%q) = not ok, want ok", root)
		}
		if next != "" {
			t.Fatalf("browseUpTarget(%q) = %q, want empty drives sentinel", root, next)
		}
	}
}

// TestBrowseUpInsideDriveReturnsToRoot covers the descent back: Up inside
// a drive climbs to the drive root, and one more Up reaches drives.
func TestBrowseUpInsideDriveReturnsToRoot(t *testing.T) {
	cases := []struct {
		root string
		want string
	}{
		{`D:\photos\2024`, `D:\photos`},
		{`D:\photos`, `D:\`},
		{`D:\photos\`, `D:\`},
		{"D:/photos/2024", "D:/photos"},
		{"D:/photos", `D:\`},
	}

	for _, tc := range cases {
		next, ok := browseUpTarget(tc.root)
		if !ok {
			t.Fatalf("browseUpTarget(%q) = not ok, want ok", tc.root)
		}
		if next != tc.want {
			t.Errorf("browseUpTarget(%q) = %q, want %q", tc.root, next, tc.want)
		}
	}

	if next, ok := browseUpTarget(`D:\`); !ok || next != "" {
		t.Errorf("browseUpTarget(`D:\\`) = %q, %v; want empty sentinel, true", next, ok)
	}
}

// TestAvailableDrivesWithStub probes the drive enumeration with a fake
// filesystem so the test is deterministic on machines without A:\..Z:\.
func TestAvailableDrivesWithStub(t *testing.T) {
	exists := func(path string) bool {
		return path == `C:\` || path == `D:\`
	}

	rows := availableDrives(exists)
	if len(rows) != 2 {
		t.Fatalf("availableDrives = %d rows, want 2", len(rows))
	}
	if rows[0].Path != `C:\` || rows[1].Path != `D:\` {
		t.Fatalf("availableDrives paths = %q, %q; want C:\\ then D:\\", rows[0].Path, rows[1].Path)
	}
	for _, row := range rows {
		if row.Name != row.Path {
			t.Errorf("drive row name = %q, want path %q", row.Name, row.Path)
		}
	}

	all := availableDrives(func(string) bool { return true })
	if len(all) != 26 {
		t.Errorf("availableDrives(all) = %d rows, want 26", len(all))
	}

	none := availableDrives(func(string) bool { return false })
	if len(none) != 0 {
		t.Errorf("availableDrives(none) = %d rows, want 0", len(none))
	}
}

// TestBrowseDrivesLevelRoundTrip wires the level to the App: an empty
// root lists drives and shows This PC, and a listed drive path climbs
// back to drives via Up.
func TestBrowseDrivesLevelRoundTrip(t *testing.T) {
	app := &App{}
	if name := app.browseDisplayName(); name != "This PC" {
		t.Fatalf("drives display name = %q, want This PC", name)
	}

	rows := availableDrives(func(path string) bool { return path == `C:\` })
	if len(rows) != 1 {
		t.Fatalf("availableDrives = %d rows, want 1", len(rows))
	}

	// Entering a drive row browses into it through the normal path.
	app.browseRoot = rows[0].Path
	if name := app.browseDisplayName(); name == "" || name == "This PC" {
		t.Fatalf("drive display name = %q, want the drive root", name)
	}

	next, ok := browseUpTarget(app.browseRoot)
	if !ok || next != "" {
		t.Fatalf("browseUpTarget(%q) = %q, %v; want empty sentinel, true", app.browseRoot, next, ok)
	}

	app.browseRoot = next
	if name := app.browseDisplayName(); name != "This PC" {
		t.Fatalf("after Up display name = %q, want This PC", name)
	}
}

// TestBrowseEntriesListsSubdirs keeps the normal level intact: real
// directories still list immediate subdirectories sorted by name.
func TestBrowseEntriesListsSubdirs(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"beta", "alpha"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0o750); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := &App{browseRoot: root}
	rows := app.browseEntries()
	if len(rows) != 2 {
		t.Fatalf("browseEntries = %d rows, want 2", len(rows))
	}
	if rows[0].Name != "alpha" || rows[1].Name != "beta" {
		t.Fatalf("browseEntries names = %q, %q; want alpha, beta", rows[0].Name, rows[1].Name)
	}
}

// TestBrowseUpUnixRootUnchanged locks the Unix behavior: / and the empty
// drives level itself go nowhere, and normal parents still resolve.
func TestBrowseUpUnixRootUnchanged(t *testing.T) {
	if next, ok := browseUpTarget("/"); ok || next != "/" {
		t.Errorf("browseUpTarget(/) = %q, %v; want /, false", next, ok)
	}
	if next, ok := browseUpTarget(""); ok || next != "" {
		t.Errorf("browseUpTarget(empty) = %q, %v; want empty, false", next, ok)
	}

	next, ok := browseUpTarget("/home/user")
	if !ok || next != "/home" {
		t.Errorf("browseUpTarget(/home/user) = %q, %v; want /home, true", next, ok)
	}

	next, ok = browseUpTarget("/mnt/c")
	if !ok || next != "/mnt" {
		t.Errorf("browseUpTarget(/mnt/c) = %q, %v; want /mnt, true", next, ok)
	}
}
