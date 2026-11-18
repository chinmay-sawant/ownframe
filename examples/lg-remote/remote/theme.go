package remote

// darkTheme is the default. lightTheme is the other side of the toggle.
// Both only set the variables the template reads, so the layout stays put.
const darkTheme = `
:root {
  --bg: #0b0b12;
  --bg-top: #1c1830;
  --accent: #e6125d;
  --accent-hi: #ff3d7f;
  --glow: rgba(230,18,93,.45);
  --shade: rgba(0,0,0,.45);
  --edge: #3d3d4a;
  --key-hi: #30303a;
  --pad-glow: #2a2a3c;
  --ink: #f2f2f2;
  --muted: #9a9a9a;
  --key: #20202a;
  --line: #343438;
  --power: #c43838;
  --pad: #15151d;
  --selected: #3a3a44;
  --selected-ink: #f2f2f2;
  --pressed: #50505a;
  --focus: #8ab4ff;
}
`

const lightTheme = `
:root {
  --bg: #eeebe4;
  --bg-top: #ffffff;
  --accent: #c8104f;
  --accent-hi: #e8396f;
  --glow: rgba(200,16,79,.3);
  --shade: rgba(60,50,30,.22);
  --edge: #cfc8ba;
  --key-hi: #f6f3ec;
  --pad-glow: #ffffff;
  --ink: #1a1a1a;
  --muted: #6d675f;
  --key: #e7e2d8;
  --line: #d5d0c6;
  --power: #a32020;
  --pad: #ffffff;
  --selected: #d1c7b6;
  --selected-ink: #1a1a1a;
  --pressed: #c8bca7;
  --focus: #1a56db;
}
`
