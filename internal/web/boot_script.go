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
      return loadDom().then(pullState);
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
    syncAudio(state);
    var gen = (typeof state.domGeneration === "number") ? state.domGeneration : state.generation;
    if (typeof gen !== "number") { return; }
    if (gen !== lastGen) {
      scheduleDom();
      return;
    }
    if (state.patches && state.patches.length) {
      if (!applyPatches(state.patches)) { scheduleDom(); }
    }
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

  function holdsActiveField(el) {
    var active = document.activeElement;
    if (!isActiveField(active) || !el || typeof el.contains !== "function") { return false; }
    return el.contains(active);
  }

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
    if (isActiveField(cur)) { return; }
    morphChildren(cur, next);
    syncControl(cur, next);
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
      lastAudio = gen;
      el.src = "/audio?v=" + gen;
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
    enqueue(function () {
      return post("/change", "id=" + encodeURIComponent(id) + "&value=" + encodeURIComponent(value));
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
    if (el.tagName === "TEXTAREA" || el.tagName === "SELECT") { return; }
    if (el.tagName === "INPUT") {
      var kind = (el.type || "").toLowerCase();
      if (kind !== "submit" && kind !== "button") { return; }
    }
    if (typeof el.closest !== "function") { return; }
    var target = el.closest("[data-action], button, input[type=submit], a[href]");
    if (!target) { return; }
    ev.preventDefault();
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
    } else {
      return;
    }
    enqueue(function () { return post("/click", body); });
  }

  // Focus only marks the target. A focus post redraws the field and eats keys.
  function onFocusIn(ev) {
    var el = eventElement(ev);
    if (!el || typeof el.setAttribute !== "function") { return; }
    el.setAttribute("data-ownframe-focus", "1");
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
      .then(function (state) { applyState(state); })
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
  document.addEventListener("click", onClick);
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
