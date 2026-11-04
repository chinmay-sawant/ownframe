package main

// hoverCSS adds selectable-row hover, selected outline, sidebar active,
// and toast elevation styles for MailPulse.
const hoverCSS = `
.mp-selectable { cursor: pointer; border-radius: 10px; transition: background .12s; }
.mp-selectable:hover { background: #f7f7f4; }
.mp-selected { outline: 2px solid #8b5cf6; outline-offset: -2px; background: #faf8ff; }
.of-sidebar-item { transition: background .12s; }
#mp-sidebar .of-sidebar-item--active { background: #ececec; font-weight: 600; }
.mp-toast { z-index: 50; box-shadow: 0 8px 24px rgba(0, 0, 0, 0.25); }
`
