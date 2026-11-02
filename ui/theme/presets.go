// Package theme provides OwnFrame palettes and typed semantic design tokens.
package theme

// Light is the default light palette. Components also work without it
// through CSS variable fallbacks. Append a :root rule to override tokens.
const Light = `html,body{background-color:var(--of-background,#f5f7fb)}
:root {
 --of-background:#f5f7fb; --of-surface:#ffffff; --of-control:#eef2f8;
 --of-text:#172b4d; --of-muted:#5c6d85; --of-primary:#315fca;
 --of-on-primary:#ffffff; --of-danger:#b3363f; --of-on-danger:#ffffff; --of-border:#dce4f0; --of-selected:#e5edff;
 --of-radius:8px; --of-spacing:16px; --of-scrollbar:#899ab1;
}
.of-scrollbar-thumb:focus-visible{outline:2px solid var(--of-primary);outline-offset:1px}`

// Dark changes paint while keeping the light theme's geometry.
const Dark = `html,body{background-color:var(--of-background,#111827)}
:root {
 --of-background:#111827; --of-surface:#1c2738; --of-control:#2b3950;
 --of-text:#e4ecf8; --of-muted:#9aabc3; --of-primary:#a6c1ff;
 --of-on-primary:#172b4d; --of-danger:#b3363f; --of-on-danger:#ffffff; --of-border:#39465a; --of-selected:#304867;
 --of-radius:8px; --of-spacing:16px; --of-scrollbar:#71849e;
}
.of-scrollbar-thumb:focus-visible{outline:2px solid var(--of-primary);outline-offset:1px}`
