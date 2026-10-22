package console

const dayTheme = `
:root {
  --bg: #e8e2d4;
  --card: #f7f3ea;
  --ink: #1c1915;
  --accent: #8a4b12;
  --rail: #2c241c;
  --muted: #6b6560;
  --line: #c8c2b4;
}
`

const nightTheme = `
:root {
  --bg: #1a1714;
  --card: #2a241e;
  --ink: #f0e6d8;
  --accent: #c47a3a;
  --rail: #120f0c;
  --muted: #a3988c;
  --line: #3a322a;
}
`

func (a *App) toggleNight() {
	a.view.Night = !a.view.Night
	css := dayTheme
	if a.view.Night {
		css = nightTheme
	}

	if err := a.page.SetTheme(css); err != nil {
		a.view.Notice = "Theme did not apply."

		return
	}

	a.view.Notice = "Theme updated."
}
