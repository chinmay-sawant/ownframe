package main

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe"
)

const (
	totalRows    = 1000
	windowWidth  = 840
	windowHeight = 640
)

const cullWindowHTML = `<!DOCTYPE html>
<html>
<head>
<style>
html, body {
  margin: 0;
  padding: 0;
  background: #f6f4ef;
  color: #1c1917;
  font-family: sans-serif;
}
.header {
  box-sizing: border-box;
  height: 76px;
  padding: 16px 24px;
  background: #1c1917;
  color: #fafaf9;
  border-bottom: 1px solid #292524;
}
.header-title {
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 4px;
}
.header-sub {
  font-size: 12px;
  color: #a8a29e;
}
.row {
  display: flex;
  box-sizing: border-box;
  height: 44px;
  align-items: center;
  gap: 16px;
  padding: 0 24px;
  border-bottom: 1px solid #e7e5e4;
  background: #ffffff;
}
.row-alt {
  background: #faf9f6;
}
.col-id {
  width: 72px;
  font-size: 12px;
  font-weight: bold;
  color: #78716c;
}
.col-name {
  width: 220px;
  font-size: 14px;
  font-weight: bold;
  color: #1c1917;
}
.col-detail {
  font-size: 13px;
  color: #57534e;
}
</style>
</head>
<body>
  <div class="header">
    <div class="header-title">Viewport Cull Window Demo (1,000 rows)</div>
    <div class="header-sub">Config.CullWindow skips layout and display ops for fixed-height off-screen rows while keeping full scroll height.</div>
  </div>
  {{range .Rows}}
  <div id="row-{{.ID}}" class="row{{if .Alt}} row-alt{{end}}">
    <span class="col-id">#{{.ID}}</span>
    <span class="col-name">{{.Name}}</span>
    <span class="col-detail">{{.Detail}}</span>
  </div>
  {{end}}
</body>
</html>`

type rowItem struct {
	ID     int
	Alt    bool
	Name   string
	Detail string
}

type viewData struct {
	Rows []rowItem
}

func makeRows(n int) []rowItem {
	rows := make([]rowItem, n)
	for i := range rows {
		id := i + 1
		s := strconv.Itoa(id)
		rows[i] = rowItem{
			ID:     id,
			Alt:    id%2 == 0,
			Name:   "Record " + s,
			Detail: "Fixed-height HTML/CSS flex row #" + s + " rendered with opt-in CullWindow",
		}
	}

	return rows
}

func newCullWindowPage(cull bool) (*ownframe.Page, error) {
	p, err := ownframe.New(ownframe.Config{
		Title:      "ownframe — cull-window (1,000 rows)",
		HTML:       cullWindowHTML,
		Width:      windowWidth,
		Height:     windowHeight,
		CullWindow: cull,
		Perf:       true,
	})
	if err != nil {
		return nil, err
	}

	p.SetData(viewData{Rows: makeRows(totalRows)})

	return p, nil
}
