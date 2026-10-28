package main

import "fmt"

type theme struct {
	ID, Name, Description                                                string
	Background, Panel, Panel2, Panel3, Line, Text, Muted, Dim, Accent    string
	Keyword, Type, Function, String, Number, Comment, Operator, Variable string
	Selection                                                            string
}

func themeCatalog() []theme {
	return []theme{
		{"midnight", "Midnight", "The original Code Review dark theme", "#0e111a", "#0c0e14", "#10131b", "#151924", "#262d3a", "#d7dce8", "#a0aabc", "#788397", "#7aa2f7", "#c792ea", "#ffcb6b", "#82aaff", "#c3e88d", "#f78c6c", "#8490a5", "#89ddff", "#f07178", "#334c79"},
		{"dracula", "Dracula Night", "Purple and pink on charcoal", "#282a36", "#21222c", "#282a36", "#343746", "#454858", "#f8f8f2", "#c1c5d4", "#9aa3bd", "#bd93f9", "#ff79c6", "#8be9fd", "#50fa7b", "#f1fa8c", "#bd93f9", "#9ba6c3", "#ff79c6", "#ffb86c", "#555b78"},
		{"light", "Light", "A clean, familiar light editor", "#ffffff", "#f3f3f3", "#eaeaea", "#dedede", "#c5c5c5", "#242424", "#505050", "#626262", "#005fb8", "#af00db", "#267f99", "#795e26", "#a31515", "#098658", "#477a48", "#333333", "#001080", "#add6ff"},
		{"solarized", "Solarized Light", "Warm paper with restrained color", "#fdf6e3", "#eee8d5", "#f5efdd", "#e5dec9", "#c8c0ab", "#344f56", "#526b70", "#637a7c", "#006da3", "#6c3fa3", "#9a6a00", "#006da3", "#557b00", "#a33a22", "#647b73", "#007f82", "#b22b54", "#c6dce0"},
		{"nord", "Nord", "Cool slate and muted frost", "#242933", "#2e3440", "#343c4a", "#3b4555", "#4c566a", "#eceff4", "#c0c9d8", "#9caac0", "#88c0d0", "#b48ead", "#8fbcbb", "#88c0d0", "#a3be8c", "#d08770", "#a2aec2", "#81a1c1", "#bf8b8b", "#465d75"},
		{"contrast", "High Contrast", "Bright text and strong boundaries", "#000000", "#101010", "#171717", "#252525", "#858585", "#ffffff", "#eeeeee", "#bdbdbd", "#75caff", "#ff9dff", "#ffdf75", "#75caff", "#a5f79a", "#ffba86", "#bdbdbd", "#a6eaff", "#ff9a9a", "#24598a"},
	}
}

type themeRow struct {
	ID, Name, Description string
	Active                bool
}

func validTheme(id string) string {
	for _, t := range themeCatalog() {
		if t.ID == id {
			return id
		}
	}
	return "midnight"
}

func (a *App) themeRows() []themeRow {
	rows := make([]themeRow, 0, len(themeCatalog()))
	for _, t := range themeCatalog() {
		rows = append(rows, themeRow{t.ID, t.Name, t.Description, t.ID == validTheme(a.theme)})
	}
	return rows
}

func themeCSS(id string) string {
	id = validTheme(id)
	for _, t := range themeCatalog() {
		if t.ID != id {
			continue
		}
		semantic := "--green:#3ddc84; --red:#f0606a; --amber:#f0b75a; --purple:#c792ea;"
		if id == "light" || id == "solarized" {
			semantic = "--green:#176b35; --red:#b32930; --amber:#865400; --purple:#733d96;"
		}
		if id == "nord" {
			semantic = `--green:#a3be8c; --red:#e48a91; --amber:#ebcb8b; --purple:#b48ead;
			--accent-soft:rgba(136,192,208,0.14); --green-bg:rgba(163,190,140,0.10);
			--green-word:rgba(163,190,140,0.27); --green-active:rgba(163,190,140,0.18);
			--green-gutter:rgba(163,190,140,0.08); --red-bg:rgba(191,97,106,0.10);
			--red-word:rgba(191,97,106,0.27); --red-active:rgba(191,97,106,0.18);
			--red-gutter:rgba(191,97,106,0.08); --amber-soft:rgba(235,203,139,0.12);`
		}
		icons := ""
		if id == "light" || id == "solarized" {
			icons = `
.commit-tip-float, .inline-card { box-shadow:none !important; }
.lang-go { background:#d5eef2 !important; color:#075c70 !important; }
.lang-python { background:#f5e9c8 !important; color:#715300 !important; }
.lang-javascript { background:#f5edc3 !important; color:#675500 !important; }
.lang-markdown { background:#dce8f8 !important; color:#24578f !important; }
.lang-css { background:#eae0f7 !important; color:#663994 !important; }
.lang-json { background:#f7e7ce !important; color:#795000 !important; }
.lang-html { background:#fae0d9 !important; color:#9b3c25 !important; }
.lang-yaml { background:#d8efe8 !important; color:#146451 !important; }
.lang-shell { background:#e5eed9 !important; color:#466420 !important; }
.lang-plain { background:#e4e7eb !important; color:#485464 !important; }`
		}
		return fmt.Sprintf(`:root { %s --bg:%s !important; --panel:%s !important; --panel-2:%s !important; --panel-3:%s !important; --line:%s !important; --line-soft:%s !important; --text:%s !important; --muted:%s !important; --dim:%s !important; --icon:%s !important; --icon-strong:%s !important; --accent:%s !important; --syntax-keyword:%s !important; --syntax-type:%s !important; --syntax-function:%s !important; --syntax-string:%s !important; --syntax-number:%s !important; --syntax-comment:%s !important; --syntax-operator:%s !important; --syntax-variable:%s !important; --selection:%s !important; }`, semantic, t.Background, t.Panel, t.Panel2, t.Panel3, t.Line, t.Line, t.Text, t.Muted, t.Dim, t.Muted, t.Muted, t.Accent, t.Keyword, t.Type, t.Function, t.String, t.Number, t.Comment, t.Operator, t.Variable, t.Selection) + icons
	}
	return ""
}
