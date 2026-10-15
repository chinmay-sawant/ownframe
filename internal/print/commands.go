package print

// linuxCommands print directly, then open the file in a viewer.
var linuxCommands = []command{
	{name: "lp", args: lpFileArg},
	{name: "xdg-open", args: fileArg},
}

// darwinCommands open the print dialog, then fall back to the viewer.
var darwinCommands = []command{
	{name: "osascript", args: appleScriptPrint},
	{name: "open", args: openFileArg},
}

// windowsCommands hand the file to the shell print verb.
var windowsCommands = []command{
	{name: "powershell.exe", args: powershellPrint},
	{name: "pwsh.exe", args: powershellPrint},
}

func fileArg(path string) []string {
	return []string{"--", path}
}
