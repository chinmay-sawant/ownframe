package web

// bootScript runs in the application document itself. There is no iframe.
// It posts field values and clicks to the page and morphs the body in place
// when the server redraws. A non-zero audioGen on GET /state points the
// preview audio element at /audio. EventSource pushes state; the 100ms
// state poll runs only where EventSource is unavailable.
const bootScript = `"use strict";
(function () {
  var queue = Promise.resolve();
  var lastGen = null;
  var lastAudio = 0;
  var domScheduled = false;
  var events = null;
  var restoringFocus = false;
  // Milliseconds between GET /state polls. Used only when the browser has
  // no EventSource; with SSE the server pushes and the poll stays off.
  var pollMs = 100;

  function enqueue(task) {
    queue = queue.then(task).catch(function () {});
    return queue;
  }

  function post(path, body) {
    return fetch(path, {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: body,
      credentials: "same-origin"
    }).then(function (resp) {
      if (!(resp && resp.ok)) { return; }
      return loadDom().then(function () {
        var el = ensureAudio();
        if (el) { syncSeek(el); }
      }).then(pullState);
    }).catch(function () {});
  }

  function loadDom() {
    return fetch("/dom", {
      method: "GET",
      cache: "no-store",
      credentials: "same-origin"
    }).then(function (resp) {
      if (!resp || !resp.ok) { return null; }
      return resp.json();
    }).then(function (data) {
      if (!data) { return; }
      if (typeof data.generation === "number") { lastGen = data.generation; }
      if (typeof data.html !== "string" || data.html === "" || data.fallback === true) { return; }
      patchDocument(data);
    }).catch(function () {});
  }

  function applyPatches(ops) {
    if (!ops || !ops.length) { return true; }
    var failed = false;
    var i;
    for (i = 0; i < ops.length; i++) {
      try {
        var op = ops[i];
        if (!op || typeof op.id !== "string" || !op.id) { failed = true; continue; }
        var el = document.getElementById(op.id);
        if (!el) { failed = true; continue; }
        if (isActiveField(el) || holdsActiveField(el)) { continue; }
        if (typeof op.style === "string") {
          el.setAttribute("style", op.style);
        } else if (op.geometry && typeof op.geometry === "object") {
          if (typeof op.geometry.left === "string") { el.style.left = op.geometry.left; }
          if (typeof op.geometry.top === "string") { el.style.top = op.geometry.top; }
          if (typeof op.geometry.width === "string") { el.style.width = op.geometry.width; }
          if (typeof op.geometry.height === "string") { el.style.height = op.geometry.height; }
        } else if (typeof op.text === "string") {
          if (el.textContent !== op.text) { el.textContent = op.text; }
        } else {
          failed = true;
        }
      } catch (err) {
        failed = true;
      }
    }
    return !failed;
  }

  function applyState(state) {
    if (!state) { return; }
    var gen = (typeof state.domGeneration === "number") ? state.domGeneration : state.generation;
    if (typeof gen !== "number") { return; }
    if (gen !== lastGen) {
      scheduleDom();
      syncAudio(state);
      return;
    }
    if (state.patches && state.patches.length) {
      if (!applyPatches(state.patches)) { scheduleDom(); }
    }
    syncAudio(state);
  }

  function scheduleDom() {
    if (domScheduled) { return; }
    domScheduled = true;
    enqueue(function () {
      domScheduled = false;
      return loadDom();
    });
  }

  function eventElement(ev) {
    var node = ev.target;
    if (!node) { return null; }
    if (node.nodeType === 1) { return node; }
    return node.parentElement || null;
  }

  function isActiveField(el) {
    return !!el && el === document.activeElement &&
      (el.tagName === "INPUT" || el.tagName === "TEXTAREA");
  }

  function hasCaret(el) {
    if (!el) { return false; }
    if (el.tagName === "TEXTAREA") { return true; }
    if (el.tagName !== "INPUT") { return false; }
    var kind = (el.type || "").toLowerCase();
    return kind === "" || kind === "text" || kind === "password" ||
      kind === "search" || kind === "tel" || kind === "url";
  }

  function postCaret(el) {
    if (!hasCaret(el) || !el.id) { return; }
    var caret;
    try {
      if (typeof el.selectionStart !== "number") { return; }
      caret = runeOffset(el.value, el.selectionStart);
    } catch (err) { return; }
    var body = "id=" + encodeURIComponent(el.id) + "&caret=" + encodeURIComponent(caret);
    enqueue(function () { return post("/change", body); });
  }

  function runeOffset(value, utf16Offset) {
    return Array.from(value.slice(0, utf16Offset)).length;
  }

  function utf16Offset(value, runeOffset) {
    return Array.from(value).slice(0, runeOffset).join("").length;
  }

  function holdsActiveField(el) {
    var active = document.activeElement;
    if (!isActiveField(active) || !el || typeof el.contains !== "function") { return false; }
    return el.contains(active);
  }

  var specialKeys = {
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
    " ": "space",
    "Spacebar": "space",
    "Backspace": "backspace",
    "Delete": "delete",
    "Control": "control",
    "Shift": "shift",
    "Meta": "meta",
    "Alt": "alt"
  };

  function pageKey(ev) {
    var name = specialKeys[ev.key] || "";
    if (!name && ev.key && ev.key.length === 1 && ev.key.charCodeAt(0) > 32 && ev.key.charCodeAt(0) < 127) {
      name = ev.key.toLowerCase();
    }
    if (!name) { return ""; }
    if (name === "arrowleft" || name === "arrowright" || name === "home" || name === "end") {
      if (ev.shiftKey) { name = "shift+" + name; }
      if (ev.ctrlKey || ev.metaKey) { name = "ctrl+" + name; }
      if (ev.altKey) { name = "alt+" + name; }
    }
    return name;
  }

  function wheelUnits(delta, mode) {
    if (mode === 1) { return delta / 3; }
    if (mode === 2) { return delta; }
    return delta / 100;
  }

  function hasNativeScroll(el, dx, dy) {
    var node = el;
    while (node && node !== document.body && node !== document.documentElement) {
      if (node.nodeType === 1) {
        var style = window.getComputedStyle(node);
        if (dx !== 0 && (style.overflowX === "auto" || style.overflowX === "scroll") &&
            node.scrollWidth > node.clientWidth &&
            ((dx < 0 && node.scrollLeft > 0) ||
             (dx > 0 && node.scrollLeft + node.clientWidth < node.scrollWidth))) {
          return true;
        }
        if (dy !== 0 && (style.overflowY === "auto" || style.overflowY === "scroll") &&
            node.scrollHeight > node.clientHeight &&
            ((dy < 0 && node.scrollTop > 0) ||
             (dy > 0 && node.scrollTop + node.clientHeight < node.scrollHeight))) {
          return true;
        }
      }
      node = node.parentElement;
    }
    var root = document.scrollingElement || document.documentElement;
    if (root) {
      if (dy < 0 && root.scrollTop > 0) { return true; }
      if (dy > 0 && root.scrollTop + root.clientHeight < root.scrollHeight - 1) { return true; }
      if (dx < 0 && root.scrollLeft > 0) { return true; }
      if (dx > 0 && root.scrollLeft + root.clientWidth < root.scrollWidth - 1) { return true; }
    }
    return false;
  }

  function onWheel(ev) {
    var el = eventElement(ev);
    if (el && typeof el.closest === "function" &&
        el.closest("input, textarea, select, [contenteditable=true]")) { return; }
    if (hasNativeScroll(el, ev.deltaX, ev.deltaY)) { return; }
    if (ev.cancelable) { ev.preventDefault(); }
    var body = "x=" + encodeURIComponent(ev.pageX) +
      "&y=" + encodeURIComponent(ev.pageY) +
      "&dx=" + encodeURIComponent(-wheelUnits(ev.deltaX, ev.deltaMode)) +
      "&dy=" + encodeURIComponent(-wheelUnits(ev.deltaY, ev.deltaMode));
    enqueue(function () { return post("/wheel", body); });
  }

  function isNativeKeyTarget(el) {
    if (!el) { return false; }
    var tag = el.tagName;
    if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT" || el.isContentEditable) { return true; }
    return typeof el.closest === "function" && !!el.closest("button, a[href], [role=button]");
  }

  function onKey(ev, down) {
    var target = eventElement(ev);
    if (isNativeKeyTarget(target)) { return; }
    var name = pageKey(ev);
    if (!name || name === "tab") { return; }
    if (ev.cancelable) { ev.preventDefault(); }
    var body = "key=" + encodeURIComponent(name) + "&down=" + (down ? "1" : "0");
    if (ev.shiftKey) { body += "&shift=1"; }
    enqueue(function () { return post("/key", body); });
  }

  function onKeyDown(ev) { onKey(ev, true); }
  function onKeyUp(ev) { onKey(ev, false); }

  function keepFocusAttr(el) {
    return !!el && el === document.activeElement;
  }

  // syncStyles copies stylesheet text from the new head, then keeps
  // style#of-theme last when the JSON theme field is a non-empty string.
  function syncStyles(theme, parsed) {
    var head = document.head;
    if (!head) { return; }
    var incoming = [];
    if (parsed.head) {
      var found = parsed.head.querySelectorAll("style");
      var n;
      for (n = 0; n < found.length; n++) {
        if (found[n].id === "of-theme") { continue; }
        incoming.push(found[n].textContent);
      }
    }
    var have = [];
    var current = head.querySelectorAll("style");
    for (n = 0; n < current.length; n++) {
      if (current[n].id === "of-theme") { continue; }
      have.push(current[n]);
    }
    for (n = 0; n < incoming.length; n++) {
      if (!have[n]) {
        have[n] = document.createElement("style");
        head.appendChild(have[n]);
      }
      if (have[n].textContent !== incoming[n]) { have[n].textContent = incoming[n]; }
    }
    for (n = have.length - 1; n >= incoming.length; n--) { have[n].remove(); }
    var themeEl = document.getElementById("of-theme");
    if (themeEl && themeEl.tagName !== "STYLE") { themeEl = null; }
    if (typeof theme === "string" && theme !== "") {
      if (!themeEl) {
        themeEl = document.createElement("style");
        themeEl.id = "of-theme";
      }
      if (themeEl.textContent !== theme) { themeEl.textContent = theme; }
      head.appendChild(themeEl);
      return;
    }
    if (themeEl && themeEl.parentNode === head) { themeEl.remove(); }
  }

  function sameKind(cur, next) {
    if (!cur || !next || cur.nodeType !== next.nodeType) { return false; }
    if (cur.nodeType === 1 && cur.tagName !== next.tagName) { return false; }
    return true;
  }

  function syncAttrs(cur, next) {
    var active = keepFocusAttr(cur);
    var keepValue = isActiveField(cur);
    var seen = Object.create(null);
    var attrs = next.attributes;
    var i;
    for (i = 0; i < attrs.length; i++) {
      var attr = attrs[i];
      var name = attr.name;
      seen[name] = true;
      if (active && name === "data-ownframe-focus") { continue; }
      if (keepValue && name === "value") { continue; }
      if (cur.getAttribute(name) !== attr.value) { cur.setAttribute(name, attr.value); }
    }
    var own = cur.attributes;
    for (i = own.length - 1; i >= 0; i--) {
      var have = own[i].name;
      if (seen[have]) { continue; }
      if (active && have === "data-ownframe-focus") { continue; }
      if (keepValue && have === "value") { continue; }
      cur.removeAttribute(have);
    }
  }

  // Unfocused controls adopt the parsed value. Replacing the node would
  // drop it; setting the attribute alone leaves a dirty field unchanged.
  function syncControl(cur, next) {
    var tag = cur.tagName;
    if (tag === "INPUT") {
      var kind = (cur.type || "").toLowerCase();
      if (kind === "checkbox" || kind === "radio") {
        if (cur.checked !== next.checked) { cur.checked = next.checked; }
        return;
      }
      if (kind === "file") { return; }
      if (cur.value !== next.value) { cur.value = next.value; }
      return;
    }
    if (tag === "TEXTAREA") {
      if (cur.value !== next.value) { cur.value = next.value; }
      return;
    }
    if (tag === "SELECT") {
      if (cur.multiple) { return; }
      if (cur.value !== next.value) { cur.value = next.value; }
      return;
    }
    if (tag === "OPTION" && cur.selected !== next.selected) { cur.selected = next.selected; }
  }

  function syncSelection(cur, next) {
    if (cur.tagName !== "INPUT" && cur.tagName !== "TEXTAREA") { return; }
    if (next.hasAttribute("data-ownframe-focus") && !isActiveField(cur)) {
      restoringFocus = true;
      try { cur.focus({ preventScroll: true }); }
      catch (err) { try { cur.focus(); } catch (ignored) {} }
      restoringFocus = false;
    }
    if (!isActiveField(cur)) { return; }
    var start = parseInt(next.getAttribute("data-ownframe-selection-start"), 10);
    var end = parseInt(next.getAttribute("data-ownframe-selection-end"), 10);
    if (!isFinite(start) || !isFinite(end)) { return; }
    start = utf16Offset(cur.value, Math.max(0, start));
    end = utf16Offset(cur.value, Math.max(0, end));
    try {
      if (cur.selectionStart !== start || cur.selectionEnd !== end) { cur.setSelectionRange(start, end); }
    } catch (err) {}
  }

  function atOrAfter(parent, node, index) {
    var kids = parent.childNodes;
    var n;
    for (n = index; n < kids.length; n++) {
      if (kids[n] === node) { return true; }
    }
    return false;
  }

  function insertAt(parent, node, index) {
    var before = parent.childNodes[index] || null;
    if (before) { parent.insertBefore(node, before); }
    else { parent.appendChild(node); }
  }

  // Moving a focused field detaches it for a moment and Chrome blurs it.
  // Put the caret back on the same node with the same selection.
  function moveChild(parent, node, index) {
    var kids = parent.childNodes;
    if (kids[index] === node) { return; }
    var active = document.activeElement;
    var keep = null;
    var sel = null;
    if (isActiveField(active) && (active === node || (node.nodeType === 1 && node.contains(active)))) {
      keep = active;
      try {
        sel = {start: active.selectionStart, end: active.selectionEnd, dir: active.selectionDirection};
      } catch (err) {
        sel = null;
      }
    }
    var from = -1;
    var n;
    for (n = 0; n < kids.length; n++) {
      if (kids[n] === node) { from = n; break; }
    }
    if (from < 0) {
      insertAt(parent, node, index);
    } else {
      var at = index;
      if (from < at) { at++; }
      var before = parent.childNodes[at] || null;
      if (before) { parent.insertBefore(node, before); }
      else { parent.appendChild(node); }
    }
    if (!keep || document.activeElement === keep) { return; }
    keep.focus();
    if (!sel) { return; }
    try { keep.setSelectionRange(sel.start, sel.end, sel.dir); }
    catch (err) {}
  }

  function canPositionMatch(node, want, pending) {
    if (!sameKind(node, want)) { return false; }
    if (node.nodeType === 1 && node.id && pending[node.id]) { return false; }
    return true;
  }

  function forget(match, want, byId, pending) {
    if (match.nodeType === 1 && match.id) {
      delete byId[match.id];
      delete pending[match.id];
    }
    if (want.nodeType === 1 && want.id) {
      delete byId[want.id];
      delete pending[want.id];
    }
  }

  function morphNode(cur, next) {
    if (!cur || !next) { return; }
    if (cur.nodeType === 3 || cur.nodeType === 4) {
      if (cur.data !== next.data) { cur.replaceWith(document.createTextNode(next.data)); }
      return;
    }
    if (cur.nodeType === 8) {
      if (cur.data !== next.data) { cur.replaceWith(document.createComment(next.data)); }
      return;
    }
    if (cur.nodeType !== 1 || next.nodeType !== 1 || cur.tagName !== next.tagName) {
      if (isActiveField(cur) || holdsActiveField(cur)) { return; }
      cur.replaceWith(next.cloneNode(true));
      return;
    }
    syncAttrs(cur, next);
    if (isActiveField(cur)) {
      syncControl(cur, next);
      syncSelection(cur, next);
      return;
    }
    morphChildren(cur, next);
    syncControl(cur, next);
    syncSelection(cur, next);
  }

  // Match an element id anywhere in this sibling list, then by position.
  // A node with no match is inserted at that index, not on document.body.
  function morphChildren(curParent, nextParent) {
    var nextKids = Array.prototype.slice.call(nextParent.childNodes);
    var pending = Object.create(null);
    var byId = Object.create(null);
    var i;
    var live = curParent.childNodes;
    for (i = 0; i < live.length; i++) {
      var have = live[i];
      if (have.nodeType === 1 && have.id && !byId[have.id]) { byId[have.id] = have; }
    }
    for (i = 0; i < nextKids.length; i++) {
      var marked = nextKids[i];
      if (marked.nodeType === 1 && marked.id) { pending[marked.id] = true; }
    }

    var index = 0;
    for (i = 0; i < nextKids.length; i++) {
      var want = nextKids[i];
      var match = null;
      if (want.nodeType === 1 && want.id) {
        var found = byId[want.id];
        if (found && found.parentNode === curParent && atOrAfter(curParent, found, index)) {
          match = found;
        }
      }
      if (!match) {
        var at = curParent.childNodes[index] || null;
        if (canPositionMatch(at, want, pending)) { match = at; }
      }
      if (!match) {
        insertAt(curParent, want.cloneNode(true), index);
        index++;
        continue;
      }
      if (!sameKind(match, want)) {
        if (isActiveField(match) || holdsActiveField(match)) {
          insertAt(curParent, want.cloneNode(true), index);
          index++;
          continue;
        }
        moveChild(curParent, match, index);
        match.replaceWith(want.cloneNode(true));
        forget(match, want, byId, pending);
        index++;
        continue;
      }
      forget(match, want, byId, pending);
      moveChild(curParent, match, index);
      morphNode(match, want);
      index++;
    }
    dropExtra(curParent, index);
  }

  function dropExtra(parent, index) {
    while (parent.childNodes[index]) {
      var extra = parent.childNodes[index];
      if (isActiveField(extra) || (extra.nodeType === 1 && holdsActiveField(extra))) {
        index++;
        continue;
      }
      extra.remove();
    }
  }

  function patchDocument(data) {
    var scrollX = window.scrollX;
    var scrollY = window.scrollY;
    var parsed = null;
    try {
      parsed = new DOMParser().parseFromString(data.html, "text/html");
    } catch (err) {
      parsed = null;
    }
    if (!parsed || !parsed.body || !document.body) {
      window.scrollTo(scrollX, scrollY);
      return;
    }
    syncStyles(data.theme, parsed);
    ensureAudio();
    morphNode(document.body, parsed.body);
    keepAudio();
    window.scrollTo(scrollX, scrollY);
  }

  // The player lives on the head so a body morph cannot drop it.
  // audioGen 0 means there is no clip.
  function ensureAudio() {
    var nodes = document.querySelectorAll("audio#ownframe-audio");
    var el = nodes.length ? nodes[0] : null;
    var i;
    for (i = 1; i < nodes.length; i++) { nodes[i].remove(); }
    if (!el) {
      el = document.createElement("audio");
      el.id = "ownframe-audio";
    }
    el.controls = false;
    el.removeAttribute("controls");
    var head = document.head || document.documentElement;
    if (head && el.parentNode !== head) { head.appendChild(el); }
    return el;
  }

  function playEl(el) {
    if (!el || typeof el.play !== "function") { return; }
    try {
      var pending = el.play();
      if (pending && typeof pending.catch === "function") {
        pending.catch(function () {});
      }
    } catch (err) {}
  }

  function sameClip(el, gen) {
    var rel = "/audio?v=" + gen;
    if ((el.getAttribute("src") || "") === rel) { return true; }
    try {
      return el.src === new URL(rel, window.location.href).href;
    } catch (err) {
      return false;
    }
  }

  var lastPlaying = false;
  var lastVolume = 1;

  function clampVol(v) {
    if (typeof v !== "number" || !isFinite(v)) { return lastVolume; }
    if (v < 0) { return 0; }
    if (v > 1) { return 1; }
    return v;
  }

  // fillFraction reads the committed seek position from the live document.
  // It returns -1 when the fill is absent or unparseable.
  function fillFraction() {
    try {
      var fill = document.getElementById("seek-fill");
      if (!fill) { return -1; }
      var style = fill.getAttribute("style") || "";
      var m = style.match(/width\s*:\s*([\d.]+)\s*%/);
      if (!m) { return -1; }
      var fraction = parseFloat(m[1]) / 100;
      if (!isFinite(fraction) || fraction < 0 || fraction > 1) { return -1; }
      return fraction;
    } catch (err) {
      return -1;
    }
  }

  function trackDuration() {
    try {
      var seek = document.getElementById("seek");
      var raw = seek && seek.getAttribute("data-duration");
      if (!raw) { return 0; }
      var parts = raw.split(":");
      if (parts.length !== 2 || !/^\d+$/.test(parts[0]) || !/^\d{1,2}$/.test(parts[1])) { return 0; }
      var minutes = parseInt(parts[0], 10);
      var seconds = parseInt(parts[1], 10);
      if (seconds > 59) { return 0; }
      return minutes * 60 + seconds;
    } catch (err) {
      return 0;
    }
  }

  // syncSeek moves the browser element to the committed fill width when it
  // disagrees by more than a second. It runs after every click round trip
  // (the fill just morphed) and on every clip reload, so the element never
  // lags the document. Paused state is untouched: callers handle pause.
  // Some MP3 streams do not report a duration to HTMLMediaElement. In that
  // case, use the current track length printed on the seek element.
  function syncSeek(el, fresh) {
    if (!el) { return; }
    var apply = function () {
      try {
        var dur = el.duration;
        if (!isFinite(dur) || dur <= 0) { dur = trackDuration(); }
        if (!isFinite(dur) || dur <= 0) { return; }
        var fraction = fillFraction();
        if (fraction < 0) { return; }
        var target = fraction * dur;
        try {
          if (Math.abs(el.currentTime - target) > 1) { el.currentTime = target; }
        } catch (err) {}
      } catch (err) {}
    };
    if (!fresh) {
      try {
        if (el.readyState >= 1) { apply(); return; }
      } catch (err) {}
    }
    try { el.addEventListener("loadedmetadata", apply, { once: true }); } catch (err) {}
  }

  // applyTransport follows the page. A paused page pauses the element.
  // Volume is 0 to 1, so a bar at 10% is 0.1 here. A new clip does not
  // start until the page says it is playing.
  function applyTransport(state) {
    var el = ensureAudio();
    if (!el || !state) { return; }
    if (typeof state.audioVolume === "number") {
      lastVolume = clampVol(state.audioVolume);
    }
    el.volume = lastVolume;
    var playing = state.audioPlaying === true;
    lastPlaying = playing;
    var gen = typeof state.audioGen === "number" && isFinite(state.audioGen) ? state.audioGen : 0;
    if (gen && gen !== lastAudio) {
      var hadClip = !!lastAudio;
      lastAudio = gen;
      el.src = "/audio?v=" + gen;
      if (hadClip) { syncSeek(el, true); }
    }
    if (!playing || !lastAudio) {
      el.pause();
      return;
    }
    if (el.paused) { playEl(el); }
  }

  function syncAudio(state) { applyTransport(state); }

  function keepAudio() {
    var el = ensureAudio();
    if (!el) { return; }
    el.volume = lastVolume;
    if (lastAudio && !sameClip(el, lastAudio)) { el.src = "/audio?v=" + lastAudio; }
    if (!lastPlaying) { el.pause(); }
  }

  function onInput(ev) {
    var el = eventElement(ev);
    if (!el || !el.id) { return; }
    if (el.tagName !== "INPUT" && el.tagName !== "TEXTAREA") { return; }
    if (el.tagName === "INPUT") {
      var kind = (el.type || "").toLowerCase();
      if (kind === "checkbox" || kind === "radio" || kind === "button" || kind === "submit") { return; }
    }
    var id = el.id;
    var value = el.value;
    var caret = null;
    try {
      if (typeof el.selectionStart === "number") { caret = runeOffset(value, el.selectionStart); }
    } catch (err) {}
    enqueue(function () {
      var body = "id=" + encodeURIComponent(id) + "&value=" + encodeURIComponent(value);
      if (caret !== null) { body += "&caret=" + encodeURIComponent(caret); }
      return post("/change", body);
    });
  }

  function onChange(ev) {
    var el = eventElement(ev);
    if (!el || !el.id) { return; }
    var id = el.id;
    if (el.tagName === "SELECT") {
      var value = el.value;
      enqueue(function () {
        return post("/change", "id=" + encodeURIComponent(id) + "&value=" + encodeURIComponent(value));
      });
      return;
    }
    if (el.tagName !== "INPUT") { return; }
    var kind = (el.type || "").toLowerCase();
    if (kind !== "checkbox" && kind !== "radio") { return; }
    var checked = el.checked ? "1" : "0";
    enqueue(function () {
      return post("/change", "id=" + encodeURIComponent(id) + "&checked=" + checked);
    });
  }

  function onClick(ev) {
    var player = ensureAudio();
    if (player) {
      player.volume = lastVolume;
      // Inside the click so the browser allows playback. pullState pauses
      // again when the page is not playing.
      if (player.src) { playEl(player); }
    }
    var el = eventElement(ev);
    if (!el) { return; }
    if (hasCaret(el)) {
      postCaret(el);
      return;
    }
    if (el.tagName === "TEXTAREA" || el.tagName === "SELECT") { return; }
    if (typeof el.closest === "function" && el.closest("label")) { return; }
    if (el.tagName === "INPUT") {
      var kind = (el.type || "").toLowerCase();
      if (kind !== "submit" && kind !== "button") { return; }
    }
    if (typeof el.closest !== "function") { return; }
    var target = el.closest("[data-action], button, input[type=submit], a[href], [role=button]");
    if (!target) { target = el.closest("[id]"); }
    if (!target) { return; }
    var action = target.getAttribute("data-action");
    var body = "";
    if (target.id) {
      body = "id=" + encodeURIComponent(target.id);
      var rect = target.getBoundingClientRect();
      if (rect.width > 0 && rect.height > 0) {
        body += "&ox=" + encodeURIComponent(ev.clientX - rect.left) +
          "&oy=" + encodeURIComponent(ev.clientY - rect.top) +
          "&w=" + encodeURIComponent(rect.width) +
          "&h=" + encodeURIComponent(rect.height);
      }
    } else if (action) {
      body = "action=" + encodeURIComponent(action);
    } else if (target.tagName === "A") {
      return;
    } else {
      body = "x=" + encodeURIComponent(ev.pageX) + "&y=" + encodeURIComponent(ev.pageY);
    }
    if (action || target.tagName === "BUTTON" || target.tagName === "A" ||
        (target.tagName === "INPUT" && (target.type || "").toLowerCase() === "submit")) {
      ev.preventDefault();
    }
    enqueue(function () { return post("/click", body); });
  }

  function onMouseDown(ev) {
    var el = eventElement(ev);
    if (!el || typeof el.closest !== "function" || el.closest("label")) { return; }
    if (el.closest("input, textarea, select, button, a[href]")) { return; }
    if (!el.closest("[data-action], [id]")) { return; }
    if (ev.cancelable) { ev.preventDefault(); }
  }

  // Focus only marks the target. A focus post redraws the field and eats keys.
  function onFocusIn(ev) {
    var el = eventElement(ev);
    if (!el || typeof el.setAttribute !== "function") { return; }
    el.setAttribute("data-ownframe-focus", "1");
    if (restoringFocus || !isActiveField(el) || !el.id) { return; }
    var body = "id=" + encodeURIComponent(el.id);
    enqueue(function () { return post("/change", body); });
  }

  function onFocusOut(ev) {
    var el = eventElement(ev);
    if (!el || typeof el.removeAttribute !== "function") { return; }
    el.removeAttribute("data-ownframe-focus");
  }

  function pullState() {
    return fetch("/state", { method: "GET", cache: "no-store", credentials: "same-origin" })
      .then(function (resp) {
        if (!resp || !resp.ok) { return null; }
        return resp.json();
      })
      .then(function (state) { if (state) { applyTransport(state); } });
  }

  function poll() {
    fetch("/state", { method: "GET", cache: "no-store", credentials: "same-origin" })
      .then(function (resp) {
        if (!resp || !resp.ok) { return null; }
        return resp.json();
      })
      .then(function (state) { enqueue(function () { applyState(state); }); })
      .catch(function () {})
      .then(function () { setTimeout(poll, pollMs); });
  }

  function connectEvents() {
    if (events || typeof EventSource === "undefined") { return false; }
    try {
      events = new EventSource("/events");
    } catch (err) {
      events = null;
      return false;
    }
    events.onmessage = function (ev) {
      var state = null;
      try { state = JSON.parse(ev.data); } catch (err) { return; }
      enqueue(function () { applyState(state); });
    };
    events.onerror = function () {
      try { events.close(); } catch (err) {}
      events = null;
      setTimeout(function () { if (!events) { connectEvents(); } }, 1000);
    };
    if (typeof window !== "undefined" && typeof window.addEventListener === "function") {
      window.addEventListener("pagehide", function () {
        if (events) { try { events.close(); } catch (err) {} }
        events = null;
      });
    }
    return true;
  }

  document.addEventListener("input", onInput);
  document.addEventListener("change", onChange);
  document.addEventListener("mousedown", onMouseDown);
  document.addEventListener("click", onClick);
  document.addEventListener("wheel", onWheel, { passive: false });
  document.addEventListener("keydown", onKeyDown);
  document.addEventListener("keyup", onKeyUp);
  document.addEventListener("focusin", onFocusIn);
  document.addEventListener("focusout", onFocusOut);
  function start() {
    ensureAudio();
    if (!connectEvents()) { poll(); }
  }
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", start);
  } else {
    start();
  }
})();
`
