// Package theme provides OwnFrame palettes and typed semantic design tokens.
package theme

// Light is the default light palette. Components also work without it
// through CSS variable fallbacks. Append a :root rule to override tokens.
const Light = `html,body{background-color:var(--of-background,#f5f7fb)}
:root {
 --of-background:#f5f7fb; --of-surface:#ffffff; --of-control:#eef2f8;
 --of-text:#172b4d; --of-muted:#73829c; --of-primary:#4677ed;
 --of-on-primary:#ffffff; --of-border:#dce4f0; --of-selected:#e5edff;
 --of-radius:8px; --of-spacing:16px; --of-scrollbar:#899ab1;
}`

// Dark changes paint while keeping the light theme's geometry.
const Dark = `html,body{background-color:var(--of-background,#111827)}
:root {
 --of-background:#111827; --of-surface:#1c2738; --of-control:#2b3950;
 --of-text:#e4ecf8; --of-muted:#9aabc3; --of-primary:#4e7ff0;
 --of-on-primary:#ffffff; --of-border:#39465a; --of-selected:#304867;
 --of-radius:8px; --of-spacing:16px; --of-scrollbar:#71849e;
}`
