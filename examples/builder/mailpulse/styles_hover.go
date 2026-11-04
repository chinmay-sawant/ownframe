package main

// hoverCSS adds selectable-row hover, selected outline, sidebar active,
// and toast stacking styles for MailPulse. The toast is flat: no shadow.
const hoverCSS = `
.mp-selectable { cursor: pointer; border-radius: 10px; transition: background .12s; }
.mp-selectable:hover { background: #f7f7f4; }
.mp-selected { outline: 2px solid #8b5cf6; outline-offset: -2px; background: #faf8ff; }
.of-sidebar-item { transition: background .12s; }
#mp-sidebar .of-sidebar-item--active { background: #ececec; font-weight: 600; }
.mp-toast { z-index: 200; box-shadow: none; }
.mp-bar:hover { outline: 2px solid #7c3aed; outline-offset: -2px; cursor: pointer; }
`
