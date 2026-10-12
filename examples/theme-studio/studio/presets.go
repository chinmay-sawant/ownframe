package studio

// Preset is one named theme stylesheet. Every sheet sets the eight custom
// properties the preview reads, under :root, and nothing else.
type Preset struct{ Name, CSS string }

// Presets returns the built-in sheets in strip order. Light is first and is
// the default.
func Presets() []Preset {
	return []Preset{
		{Name: "light", CSS: lightCSS},
		{Name: "dark", CSS: darkCSS},
		{Name: "amber", CSS: amberCSS},
		{Name: "paper", CSS: paperCSS},
	}
}

// defaultCSS is the sheet New seeds into the theme file.
func defaultCSS() string { return Presets()[0].CSS }

const lightCSS = `:root {
  --bg: #f7f5f0;
  --surface: #ffffff;
  --fg: #26221d;
  --muted: #6f675e;
  --accent: #b03a2e;
  --accent-fg: #ffffff;
  --border: #d9d3c7;
  --radius: 8px;
}`

const darkCSS = `:root {
  --bg: #101216;
  --surface: #1a1e26;
  --fg: #e8eaf0;
  --muted: #969ca8;
  --accent: #7aa2f7;
  --accent-fg: #101216;
  --border: #2b313d;
  --radius: 8px;
}`

const amberCSS = `:root {
  --bg: #0b0b0c;
  --surface: #17150f;
  --fg: #f2e7cf;
  --muted: #a89a7c;
  --accent: #ffb000;
  --accent-fg: #1a1200;
  --border: #35301f;
  --radius: 8px;
}`

const paperCSS = `:root {
  --bg: #e8e2d2;
  --surface: #f6f1e3;
  --fg: #2c261c;
  --muted: #6f6551;
  --accent: #8a4b16;
  --accent-fg: #fff8ea;
  --border: #cdc4ac;
  --radius: 8px;
}`
