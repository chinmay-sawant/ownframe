package chadcn

// Light is the ChadCN light palette. The values follow the neutral reference
// captures. Components read these variables with fallbacks, so they also
// render without this sheet.
const Light = `html,body{background-color:var(--of-background,#ffffff);color:var(--of-text,#0a0a0a)}
:root {
 --of-background:#ffffff; --of-surface:#ffffff; --of-control:#f4f4f5;
 --of-card:#ffffff; --of-popover:#ffffff; --of-secondary:#f4f4f5;
 --of-text:#0a0a0a; --of-muted:#737373; --of-muted-foreground:#737373;
 --of-primary:#171717; --of-on-primary:#fafafa; --of-primary-hover:#262626;
 --of-secondary-foreground:#171717; --of-accent:#f5f5f5; --of-on-accent:#171717;
 --of-danger:#dc2626; --of-on-danger:#ffffff; --of-destructive:#dc2626;
 --of-on-destructive:#ffffff; --of-border:#e5e5e5; --of-input:#e5e5e5;
 --of-selected:#f5f5f5; --of-ring:#a1a1aa; --of-overlay:#00000052; --of-shadow:#00000018;
 --of-radius:10px; --of-radius-sm:6px; --of-radius-lg:14px;
 --of-spacing:16px; --of-scrollbar:#a1a1aa;
}`

// Dark is the ChadCN dark palette. Geometry matches Light.
const Dark = `html,body{background-color:var(--of-background,#0a0a0a);color:var(--of-text,#fafafa)}
:root {
 --of-background:#0a0a0a; --of-surface:#171717; --of-control:#262626;
 --of-card:#171717; --of-popover:#171717; --of-secondary:#262626;
 --of-text:#fafafa; --of-muted:#a3a3a3; --of-muted-foreground:#a3a3a3;
 --of-primary:#fafafa; --of-on-primary:#171717; --of-primary-hover:#e5e5e5;
 --of-secondary-foreground:#fafafa; --of-accent:#262626; --of-on-accent:#fafafa;
 --of-danger:#ef4444; --of-on-danger:#ffffff; --of-destructive:#ef4444;
 --of-on-destructive:#ffffff; --of-border:#262626; --of-input:#262626;
 --of-selected:#262626; --of-ring:#525252; --of-overlay:#000000b3; --of-shadow:#00000038;
 --of-radius:10px; --of-radius-sm:6px; --of-radius-lg:14px;
 --of-spacing:16px; --of-scrollbar:#525252;
}`
