package print

import "strings"

// linuxCommands print directly, then open the file in a viewer.
var linuxCommands = []command{
	{name: "lp", args: fileArg},
	{name: "xdg-open", args: fileArg},
}

// darwinCommands open the print dialog, then fall back to the viewer.
var darwinCommands = []command{
	{name: "osascript", args: appleScriptPrint},
	{name: "open", args: fileArg},
}

// windowsCommands hand the file to the shell print verb.
var windowsCommands = []command{
	{name: "powershell.exe", args: powershellPrint},
	{name: "pwsh.exe", args: powershellPrint},
}

func fileArg(path string) []string {
	return []string{path}
}

// appleScriptPrint asks Preview to print the file. The path travels as an
// osascript argument instead of being interpolated into the script, so a
// crafted (or newline-bearing) file name cannot break out of the string.
func appleScriptPrint(path string) []string {
	return []string{"-e", `on run argv`, "-e", `tell application "Preview" to print (POSIX file (item 1 of argv))`, "-e", `end run`, path}
}

// powershellPrint asks the shell to run the file's print verb.
func powershellPrint(path string) []string {
	quoted := strings.ReplaceAll(path, "'", "''")

	return []string{
		"-NoProfile",
		"-Command",
		"Start-Process -FilePath '" + quoted + "' -Verb Print",
	}
}
