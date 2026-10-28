package web

const shellHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>ownframe</title>
<style>
html, body { margin: 0; padding: 0; }
body { background: #111; color: #ddd; font: 14px/1.4 system-ui, sans-serif; }
#frame { display: block; max-width: 100%; height: auto; outline: none; cursor: default; }
#frame:focus { outline: 2px solid #4a9eff; outline-offset: -2px; }
.fallback { margin: 8px; }
.fallback input { width: 12em; }
</style>
</head>
<body>
<p class="fallback">This page shows the screen and sends pointer, key, and wheel input to it.</p>
<img id="frame" tabindex="0" src="/frame.png" usemap="#screen" alt="screen" data-generation="{{.Generation}}"{{if .Width}} width="{{.Width}}" height="{{.Height}}"{{end}}>
<map name="screen">
{{range .Areas}}<area shape="rect" coords="{{.Coords}}" href="{{.Href}}" alt="{{.Alt}}" data-x="{{.X}}" data-y="{{.Y}}">
{{end}}</map>
<form class="fallback" method="post" action="/type">
<label>Text <input type="text" name="text"></label>
<button type="submit">Type</button>
</form>
<form class="fallback" method="post" action="/backspace">
<button type="submit">Backspace</button>
</form>
<script>
(function () {
  "use strict";
  var frame = document.getElementById("frame");
  if (!frame) { return; }

  var pollMs = 120;
  var generation = parseInt(frame.getAttribute("data-generation") || "0", 10) || 0;

  function post(path, body) {
    return fetch(path, {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded;charset=UTF-8" },
      body: body,
      credentials: "same-origin"
    }).catch(function () {});
  }

  function sendKey(key, down, shift) {
    var body = "key=" + encodeURIComponent(key) + "&down=" + (down ? "1" : "0");
    if (shift) { body += "&shift=1"; }
    post("/key", body);
  }

  function typeText(text) {
    post("/type", "text=" + encodeURIComponent(text));
  }

  function framePoint(ev) {
    var rect = frame.getBoundingClientRect();
    var width = frame.naturalWidth || frame.width || 1;
    var height = frame.naturalHeight || frame.height || 1;
    return [
      (ev.clientX - rect.left) * width / (rect.width || 1),
      (ev.clientY - rect.top) * height / (rect.height || 1)
    ];
  }

  function wheelUnits(delta, mode) {
    if (mode === 1) { return delta / 3; }
    if (mode === 2) { return delta; }
    return delta / 100;
  }

  var special = {
    "ArrowUp": "arrowup",
    "ArrowDown": "arrowdown",
    "ArrowLeft": "arrowleft",
    "ArrowRight": "arrowright",
    "PageUp": "pageup",
    "PageDown": "pagedown",
    "Home": "home",
    "End": "end",
    "Enter": "enter",
    "Escape": "escape",
    "Esc": "escape",
    "Tab": "tab",
    " ": "space",
    "Spacebar": "space",
    "Backspace": "backspace",
    "Delete": "delete",
    "Control": "control",
    "Shift": "shift",
    "Meta": "meta",
    "Alt": "alt"
  };
  var chorded = { arrowleft: true, arrowright: true, home: true, end: true };
  var modifier = { control: true, shift: true, meta: true, alt: true };

  function specialName(key) {
    return special[key] || "";
  }

  function chordName(name, shift, control, alt) {
    if (!chorded[name]) { return name; }
    if (shift) { name = "shift+" + name; }
    if (control) { name = "ctrl+" + name; }
    if (alt) { name = "alt+" + name; }
    return name;
  }

  // JS takes over hit testing. Without it the image map above still clicks.
  frame.removeAttribute("usemap");

  frame.addEventListener("click", function (ev) {
    ev.preventDefault();
    frame.focus({ preventScroll: true });
    var point = framePoint(ev);
    post("/click", "x=" + encodeURIComponent(point[0]) + "&y=" + encodeURIComponent(point[1]));
  });

  frame.addEventListener("wheel", function (ev) {
    ev.preventDefault();
    var point = framePoint(ev);
    post("/wheel",
      "x=" + encodeURIComponent(point[0]) +
      "&y=" + encodeURIComponent(point[1]) +
      "&dx=" + encodeURIComponent(-wheelUnits(ev.deltaX, ev.deltaMode)) +
      "&dy=" + encodeURIComponent(-wheelUnits(ev.deltaY, ev.deltaMode)));
  }, { passive: false });

  window.addEventListener("keydown", function (ev) {
    if (document.activeElement !== frame) { return; }
    var name = specialName(ev.key);
    if (!name) {
      if (ev.key.length === 1 && (ev.ctrlKey || ev.metaKey) && !ev.altKey) {
        ev.preventDefault();
        sendKey(ev.key.toLowerCase(), true, ev.shiftKey);
        return;
      }
      if (ev.key.length === 1 && !ev.ctrlKey && !ev.metaKey && !ev.altKey) {
        ev.preventDefault();
        typeText(ev.key);
      }
      return;
    }
    ev.preventDefault();
    if (modifier[name]) {
      sendKey(name, true, ev.shiftKey);
      return;
    }
    if (name === "tab") {
      sendKey("tab", true, ev.shiftKey);
      return;
    }
    if (name === "space") {
      typeText(" ");
    }
    if (name === "backspace" && !ev.ctrlKey && !ev.metaKey && !ev.altKey) {
      post("/backspace", "");
    }
    sendKey(chordName(name, ev.shiftKey, ev.ctrlKey || ev.metaKey, ev.altKey), true, ev.shiftKey);
  });

  window.addEventListener("keyup", function (ev) {
    if (document.activeElement !== frame) { return; }
    var name = specialName(ev.key);
    if (!name) {
      if (ev.key.length === 1 && (ev.ctrlKey || ev.metaKey) && !ev.altKey) {
        sendKey(ev.key.toLowerCase(), false, ev.shiftKey);
      }
      return;
    }
    ev.preventDefault();
    sendKey(chordName(name, ev.shiftKey, ev.ctrlKey || ev.metaKey, ev.altKey), false, ev.shiftKey);
  });

  function poll() {
    fetch("/state", { cache: "no-store", credentials: "same-origin" })
      .then(function (resp) { return resp.ok ? resp.json() : null; })
      .then(function (state) {
        if (!state) { return; }
        if (typeof state.generation === "number" && state.generation !== generation) {
          generation = state.generation;
          frame.src = "/frame.png?g=" + encodeURIComponent(String(generation));
        }
        if (state.width && state.height &&
            (frame.width !== state.width || frame.height !== state.height)) {
          frame.width = state.width;
          frame.height = state.height;
        }
      })
      .catch(function () {})
      .then(function () { setTimeout(poll, pollMs); });
  }

  setTimeout(poll, pollMs);
  frame.focus({ preventScroll: true });
})();
</script>
{{if .Reload}}<script>
(function () {
  var frame = document.getElementById("frame");
  setInterval(function () { frame.src = "/frame.png?t=" + Date.now(); }, 250);
})();
</script>{{end}}
</body>
</html>
`
