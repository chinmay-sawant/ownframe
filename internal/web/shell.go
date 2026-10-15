package web

const shellHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>ownframe</title>
</head>
<body>
<p>This page only displays the picture. The screen is the image, not this HTML.</p>
<img id="frame" src="/frame.png" usemap="#screen" alt="screen"{{if .Width}} width="{{.Width}}" height="{{.Height}}"{{end}}>
<map name="screen">
{{range .Areas}}<area shape="rect" coords="{{.Coords}}" href="{{.Href}}" alt="{{.Alt}}" data-x="{{.X}}" data-y="{{.Y}}" onclick="return postClick(this)">
{{end}}</map>
<script>
function postClick(el) {
  var x = el.getAttribute("data-x");
  var y = el.getAttribute("data-y");
  fetch("/click?x=" + encodeURIComponent(x) + "&y=" + encodeURIComponent(y), {method: "POST"})
    .then(function () { window.location.reload(); });
  return false;
}
</script>
<form method="post" action="/type">
<label>Text <input type="text" name="text"></label>
<button type="submit">Type</button>
</form>
<form method="post" action="/backspace">
<button type="submit">Backspace</button>
</form>
{{if .Reload}}<script>
(function () {
  var frame = document.getElementById("frame");
  setInterval(function () { frame.src = "/frame.png?t=" + Date.now(); }, 250);
})();
</script>{{end}}
</body>
</html>
`
