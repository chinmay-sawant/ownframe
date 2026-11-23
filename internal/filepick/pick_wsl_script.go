//go:build linux && !android

package filepick

import "strings"

// titleEnv carries the dialog title to PowerShell out of band so the script
// stays static and the title can never be parsed as PowerShell code.
const titleEnv = "OWNFRAME_DIALOG_TITLE"

// windowScript returns the static PowerShell script; it reads the title from
// the environment and never interpolates caller input.
func windowScript() string {
	return `[Console]::OutputEncoding = [System.Text.Encoding]::UTF8;` +
		`Add-Type -AssemblyName System.Windows.Forms;` +
		`$f = New-Object System.Windows.Forms.OpenFileDialog;` +
		`$f.Title = $env:` + titleEnv + `;` +
		`if ($f.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) ` +
		`{ [Console]::Out.Write($f.FileName) }`
}

// windowEnv returns env with the title set and forwarded across the WSL
// boundary (WSLENV, "/u" = Linux to Windows only).
func windowEnv(env []string, title string) []string {
	const (
		wslenv      = "WSLENV="
		extraFields = 2 // the title and WSLENV entries
	)
	out := make([]string, 0, len(env)+extraFields)
	forward := titleEnv + "/u"
	for _, kv := range env {
		switch {
		case strings.HasPrefix(kv, titleEnv+"="):
			continue
		case strings.HasPrefix(kv, wslenv):
			if cur := strings.TrimPrefix(kv, wslenv); cur != "" {
				forward = cur + ":" + forward
			}

			continue
		}
		out = append(out, kv)
	}

	return append(out, titleEnv+"="+title, wslenv+forward)
}
