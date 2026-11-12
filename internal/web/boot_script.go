package web

// bootScript runs in the application document itself. There is no iframe.
// It posts field values and clicks to the page and patches the document
// when the server redraws. The state poll waits 100 milliseconds.
const bootScript = `"use strict";
(function () {
  var queue = Promise.resolve();
  var lastGen = null;
  var domScheduled = false;
  // Milliseconds between GET /state polls.
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
      if (resp && resp.ok) { return loadDom(); }
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

  function inBody(el) {
    return !!el && !!document.body && el !== document.body && el !== document.documentElement &&
      document.body.contains(el);
  }

  function ancestorReplaced(el, replaced) {
    var parent = el.parentElement;
    while (parent) {
      if (parent.id && replaced[parent.id]) { return true; }
      parent = parent.parentElement;
    }
    return false;
  }

  function placeNew(fresh) {
    if (!document.body) { return; }
    var parent = fresh.parentElement;
    var mount = document.body;
    while (parent) {
      if (parent.id) {
        var found = document.getElementById(parent.id);
        if (inBody(found)) { mount = found; break; }
      }
      parent = parent.parentElement;
    }
    if (mount === document.documentElement) { mount = document.body; }
    mount.appendChild(fresh.cloneNode(true));
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

  function patchBody(parsed) {
    if (!document.body || !parsed.body) { return; }
    var replaced = {};
    var nodes = parsed.body.querySelectorAll("[id]");
    var i;
    for (i = 0; i < nodes.length; i++) {
      var fresh = nodes[i];
      var id = fresh.id;
      if (!id || ancestorReplaced(fresh, replaced)) { continue; }
      var cur = document.getElementById(id);
      if (!inBody(cur)) {
        if (cur) { continue; }
        placeNew(fresh);
        replaced[id] = true;
        continue;
      }
      if (cur === document.documentElement || cur === document.body) { continue; }
      if (cur.outerHTML === fresh.outerHTML) { continue; }
      // Keep the focused field, and do not swap an ancestor that contains it.
      if (holdsActiveField(cur)) { continue; }
      cur.replaceWith(fresh.cloneNode(true));
      replaced[id] = true;
    }
    var live = Array.prototype.slice.call(document.body.querySelectorAll("[id]"));
    for (i = 0; i < live.length; i++) {
      var old = live[i];
      if (!old.id || replaced[old.id]) { continue; }
      if (parsed.getElementById(old.id)) { continue; }
      if (holdsActiveField(old)) { continue; }
      old.remove();
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
    if (!parsed || !parsed.body) {
      window.scrollTo(scrollX, scrollY);
      return;
    }
    syncStyles(data.theme, parsed);
    patchBody(parsed);
    window.scrollTo(scrollX, scrollY);
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

  function poll() {
    fetch("/state", { method: "GET", cache: "no-store", credentials: "same-origin" })
      .then(function (resp) {
        if (!resp || !resp.ok) { return null; }
        return resp.json();
      })
      .then(function (state) {
        if (!state || typeof state.generation !== "number") { return; }
        if (state.generation === lastGen) { return; }
        scheduleDom();
      })
      .catch(function () {})
      .then(function () { setTimeout(poll, pollMs); });
  }

  document.addEventListener("input", onInput);
  document.addEventListener("change", onChange);
  document.addEventListener("click", onClick);
  document.addEventListener("focusin", onFocusIn);
  document.addEventListener("focusout", onFocusOut);
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", poll);
  } else {
    poll();
  }
})();
`
