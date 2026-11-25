package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

// A minimal OpenXML reader and writer. It keeps values, formulas, number
// formats, fonts, fills, per-edge borders, alignment, widths and heights,
// hidden rows and columns, merges, freeze panes, tab colors, conditional
// formats, data validation and cell comments.

const (
	xlNSMain = "http://schemas.openxmlformats.org/spreadsheetml/2006/main"
	xlNSRel  = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
)

func xlxEsc(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

func xlNumFmtCode(st Style) string {
	dec := func(def int) string {
		n := xlDecimals(st, def)
		if n == 0 {
			return ""
		}
		return "." + strings.Repeat("0", n)
	}
	switch st.NumFmt {
	case "", "General":
		if st.Decimals > 0 {
			return "0" + dec(0)
		}
		return "General"
	case "Number":
		return "0" + dec(xlDefaultDecimals)
	case "Currency":
		return "\"$\"#,##0" + dec(xlDefaultDecimals)
	case "Accounting":
		return "_(\"$\"* #,##0" + dec(xlDefaultDecimals) + "_)"
	case "Percent":
		return "0" + dec(0) + "%"
	case "Scientific":
		return "0" + dec(xlDefaultDecimals) + "E+00"
	case "Date":
		return "yyyy-mm-dd"
	case "LongDate":
		return "ddd, mmm d, yyyy"
	case "Time":
		return "hh:mm:ss"
	case "Text":
		return "@"
	}
	return st.NumFmt
}

func xlStyleFromCode(code string) (numFmt string, dec int) {
	decOf := func() int {
		_, frac, ok := strings.Cut(code, ".")
		if !ok {
			return 0
		}
		n := 0
		for n < len(frac) && frac[n] == '0' {
			n++
		}
		return n
	}
	d := func(def int) int {
		n := decOf()
		switch n {
		case def:
			return 0
		case 0:
			return -1
		}
		return n
	}
	switch {
	case code == "" || strings.EqualFold(code, "General"):
		return "", 0
	case code == "@":
		return "Text", 0
	case code == "yyyy-mm-dd" || code == "m/d/yyyy" || code == "mm-dd-yy" || code == "m/d/yy":
		return "Date", 0
	case code == "ddd, mmm d, yyyy":
		return "LongDate", 0
	case code == "hh:mm:ss" || code == "h:mm:ss":
		return "Time", 0
	case strings.Contains(code, "%"):
		return "Percent", d(0)
	case strings.Contains(code, "E+"):
		return "Scientific", d(xlDefaultDecimals)
	case strings.Contains(code, "_(") && strings.Contains(code, "$"):
		return "Accounting", d(xlDefaultDecimals)
	case strings.Contains(code, "$"):
		return "Currency", d(xlDefaultDecimals)
	case xlIsDateCode(code):
		return code, 0
	case strings.Trim(code, "0#,.") == "":
		return "Number", d(xlDefaultDecimals)
	}
	return code, 0
}

func xlBuiltinFmts() map[int]string {
	return map[int]string{
		0: "General", 1: "0", 2: "0.00", 3: "#,##0", 4: "#,##0.00", 9: "0%", 10: "0.00%", 11: "0.00E+00",
		14: "m/d/yyyy", 15: "d-mmm-yy", 16: "d-mmm", 17: "mmm-yy", 18: "h:mm AM/PM", 19: "h:mm:ss AM/PM",
		20: "h:mm", 21: "h:mm:ss", 22: "m/d/yyyy h:mm", 49: "@",
	}
}

const (
	xlColCharPx     = 7    // pixels per character-width unit in OpenXML column widths
	xlPxToPt        = 0.75 // CSS pixels to points for row heights
	xlVMLShapeBase  = 1025 // first legacy VML shape id used for comment notes
	xlNoteCols      = 3    // comment note box width in columns
	xlNoteRows      = 4    // comment note box height in rows
	xlCustomFmtBase = 164  // first custom number-format id in styles.xml
	xlRoundHalf     = 0.5
	xlARGBLen       = 8
	xlRGBLen        = 6
)

func xlColorToARGB(css string) string {
	if len(css) == 7 && css[0] == '#' {
		return "FF" + strings.ToUpper(css[1:])
	}
	return ""
}

func xlARGBToColor(argb string) string {
	if len(argb) == xlARGBLen {
		return "#" + strings.ToUpper(argb[2:])
	}
	if len(argb) == xlRGBLen {
		return "#" + strings.ToUpper(argb)
	}
	return ""
}

// xlFontNames maps Style.FontFamily to the OpenXML font name and back.
func xlFontNames() map[string]string {
	return map[string]string{"": "Calibri", "sans": "Arial", "mono": "Courier New", "serif": "Times New Roman"}
}

func xlFamilyOf(name string) string {
	l := strings.ToLower(name)
	switch {
	case strings.Contains(l, "courier"), strings.Contains(l, "consolas"), strings.Contains(l, "mono"), strings.Contains(l, "menlo"):
		return "mono"
	case strings.Contains(l, "times"), strings.Contains(l, "georgia"), strings.Contains(l, "cambria"), strings.Contains(l, "serif") && !strings.Contains(l, "sans"), strings.Contains(l, "garamond"):
		return "serif"
	case strings.Contains(l, "arial"), strings.Contains(l, "helvetica"), strings.Contains(l, "verdana"), strings.Contains(l, "segoe"), strings.Contains(l, "sans"):
		return "sans"
	}
	return ""
}

func xlVAlignOut() map[string]string {
	return map[string]string{"top": "top", "middle": "center", "bottom": "bottom"}
}

func xlVAlignIn() map[string]string {
	return map[string]string{"top": "top", "center": "middle", "bottom": "bottom"}
}

// xlCondOps maps cell-compare rule kinds to OpenXML operators.
func xlCondOps() map[string]string {
	return map[string]string{"gt": "greaterThan", "lt": "lessThan", "eq": "equal", "between": "between"}
}

// xlxWriter collects the shared style tables while sheets are written.
type xlxWriter struct {
	xfs    []Style
	xfIdx  map[Style]int
	dxfs   [][2]string // fg, bg
	dxfIdx map[[2]string]int
}

func (x *xlxWriter) xf(st Style) int {
	key := st.migrated()
	if i, ok := x.xfIdx[key]; ok {
		return i
	}
	x.xfs = append(x.xfs, key)
	x.xfIdx[key] = len(x.xfs)
	return len(x.xfs)
}

func (x *xlxWriter) dxf(fg, bg string) int {
	k := [2]string{fg, bg}
	if i, ok := x.dxfIdx[k]; ok {
		return i
	}
	x.dxfs = append(x.dxfs, k)
	x.dxfIdx[k] = len(x.dxfs) - 1
	return len(x.dxfs) - 1
}

// cfXML renders the conditional formatting elements of a sheet.
func (x *xlxWriter) cfXML(s *Sheet) string {
	var b strings.Builder
	prio := 0
	for _, e := range s.conds {
		if e.off {
			continue
		}
		prio++
		tl := e.rng.Min.A1()
		r := e.rule
		var inner string
		pre := xlPresetOf(r.Preset, "red")
		dxf := func() string {
			return fmt.Sprintf(` dxfId="%d"`, x.dxf(xlColorToARGB(pre.fg), xlColorToARGB(pre.bg)))
		}
		lit := func(v string) string {
			if _, ok := xlCondNum(v); ok {
				return strings.TrimSpace(v)
			}
			return `"` + strings.ReplaceAll(v, `"`, `""`) + `"`
		}
		switch r.Kind {
		case "gt", "lt", "eq":
			inner = fmt.Sprintf(`<cfRule type="cellIs"%s priority="%d" operator="%s"><formula>%s</formula></cfRule>`,
				dxf(), prio, xlCondOps()[r.Kind], xlxEsc(lit(r.A)))
		case "between":
			inner = fmt.Sprintf(`<cfRule type="cellIs"%s priority="%d" operator="between"><formula>%s</formula><formula>%s</formula></cfRule>`,
				dxf(), prio, xlxEsc(lit(r.A)), xlxEsc(lit(r.B)))
		case "text":
			inner = fmt.Sprintf(`<cfRule type="containsText"%s priority="%d" operator="containsText" text="%s"><formula>NOT(ISERROR(SEARCH(%s,%s)))</formula></cfRule>`,
				dxf(), prio, xlxEsc(r.A), xlxEsc(`"`+strings.ReplaceAll(r.A, `"`, `""`)+`"`), tl)
		case "dupes":
			inner = fmt.Sprintf(`<cfRule type="duplicateValues"%s priority="%d"/>`, dxf(), prio)
		case "top", "bottom":
			n, err := strconv.Atoi(strings.TrimSpace(r.A))
			if err != nil || n < 1 {
				n = 10
			}
			bottom := ""
			if r.Kind == "bottom" {
				bottom = ` bottom="1"`
			}
			inner = fmt.Sprintf(`<cfRule type="top10"%s priority="%d"%s rank="%d"/>`, dxf(), prio, bottom, n)
		case "above":
			inner = fmt.Sprintf(`<cfRule type="aboveAverage"%s priority="%d"/>`, dxf(), prio)
		case "below":
			inner = fmt.Sprintf(`<cfRule type="aboveAverage"%s priority="%d" aboveAverage="0"/>`, dxf(), prio)
		case "databar":
			inner = fmt.Sprintf(`<cfRule type="dataBar" priority="%d"><dataBar><cfvo type="min"/><cfvo type="max"/><color rgb="%s"/></dataBar></cfRule>`,
				prio, xlColorToARGB(xlPresetOfBar(r.Preset)))
		case "scale3":
			inner = fmt.Sprintf(`<cfRule type="colorScale" priority="%d"><colorScale><cfvo type="min"/><cfvo type="percentile" val="50"/><cfvo type="max"/>`+
				`<color rgb="FFF8696B"/><color rgb="FFFFEB84"/><color rgb="FF63BE7B"/></colorScale></cfRule>`, prio)
		default:
			prio--
			continue
		}
		fmt.Fprintf(&b, `<conditionalFormatting sqref="%s">%s</conditionalFormatting>`, e.rng.String(), inner)
	}
	return b.String()
}

// dvXML renders the dataValidations element of a sheet.
func dvXML(s *Sheet) string {
	if len(s.valid) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<dataValidations count="%d">`, len(s.valid))
	for _, v := range s.valid {
		switch v.kind {
		case "list":
			fmt.Fprintf(&b, `<dataValidation type="list" allowBlank="1" showErrorMessage="1" sqref="%s"><formula1>%s</formula1></dataValidation>`,
				v.rng.String(), xlxEsc(`"`+strings.Join(v.list, ",")+`"`))
		default:
			typ := v.kind
			if typ == "length" {
				typ = "textLength"
			}
			fmt.Fprintf(&b, `<dataValidation type="%s" operator="between" allowBlank="1" showErrorMessage="1" sqref="%s"><formula1>%s</formula1><formula2>%s</formula2></dataValidation>`,
				typ, v.rng.String(), strconv.FormatFloat(v.min, 'g', -1, 64), strconv.FormatFloat(v.max, 'g', -1, 64))
		}
	}
	b.WriteString(`</dataValidations>`)
	return b.String()
}

// sheetXML renders one worksheet. legacy is the relationship id of the VML
// drawing that carries the comment shapes ("" when the sheet has no comments).
func (x *xlxWriter) sheetXML(s *Sheet, legacy string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	fmt.Fprintf(&b, `<worksheet xmlns="%s" xmlns:r="%s">`, xlNSMain, xlNSRel)
	if c := xlColorToARGB(s.TabColor); c != "" {
		fmt.Fprintf(&b, `<sheetPr><tabColor rgb="%s"/></sheetPr>`, c)
	}
	if s.freezeR > 0 || s.freezeC > 0 {
		state := "bottomRight"
		switch {
		case s.freezeC == 0:
			state = "bottomLeft"
		case s.freezeR == 0:
			state = "topRight"
		}
		fmt.Fprintf(&b, `<sheetViews><sheetView workbookViewId="0"><pane xSplit="%d" ySplit="%d" topLeftCell="%s" activePane="%s" state="frozen"/></sheetView></sheetViews>`,
			s.freezeC, s.freezeR, CellRef{s.freezeR, s.freezeC}.A1(), state)
	}
	colSet := map[int]bool{}
	for c := range s.colW {
		colSet[c] = true
	}
	for c := range s.hiddenCols {
		colSet[c] = true
	}
	if len(colSet) > 0 {
		b.WriteString("<cols>")
		for _, c := range sortedKeys(colSet) {
			hidden := ""
			if s.hiddenCols[c] {
				hidden = ` hidden="1"`
			}
			fmt.Fprintf(&b, `<col min="%d" max="%d" width="%.2f" customWidth="1"%s/>`, c+1, c+1, float64(s.ColWidth(c))/xlColCharPx, hidden)
		}
		b.WriteString("</cols>")
	}
	b.WriteString("<sheetData>")
	rowsSet := map[int]bool{}
	cellSet := map[CellRef]bool{}
	for ref := range s.cells {
		rowsSet[ref.Row], cellSet[ref] = true, true
	}
	for ref := range s.styles {
		rowsSet[ref.Row], cellSet[ref] = true, true
	}
	for r := range s.rowH {
		rowsSet[r] = true
	}
	for r := range s.hiddenRows {
		rowsSet[r] = true
	}
	byRow := map[int][]CellRef{}
	for _, ref := range sortedRefs(cellSet) {
		byRow[ref.Row] = append(byRow[ref.Row], ref)
	}
	for _, r := range sortedKeys(rowsSet) {
		attrs := ""
		if h, ok := s.rowH[r]; ok {
			attrs += fmt.Sprintf(` ht="%.2f" customHeight="1"`, float64(h)*xlPxToPt)
		}
		if s.hiddenRows[r] {
			attrs += ` hidden="1"`
		}
		fmt.Fprintf(&b, `<row r="%d"%s>`, r+1, attrs)
		for _, ref := range byRow[r] {
			b.WriteString(x.cellXML(s, ref))
		}
		b.WriteString("</row>")
	}
	b.WriteString("</sheetData>")
	if len(s.merges) > 0 {
		fmt.Fprintf(&b, `<mergeCells count="%d">`, len(s.merges))
		for _, m := range s.merges {
			fmt.Fprintf(&b, `<mergeCell ref="%s"/>`, m.String())
		}
		b.WriteString("</mergeCells>")
	}
	b.WriteString(x.cfXML(s))
	b.WriteString(dvXML(s))
	if legacy != "" {
		fmt.Fprintf(&b, `<legacyDrawing r:id="%s"/>`, legacy)
	}
	b.WriteString("</worksheet>")
	return b.String()
}

func (x *xlxWriter) cellXML(s *Sheet, ref CellRef) string {
	st := s.styles[ref]
	attrs := fmt.Sprintf(` r="%s"`, ref.A1())
	if st != (Style{}) {
		attrs += fmt.Sprintf(` s="%d"`, x.xf(st))
	}
	c := s.cells[ref]
	if c == nil {
		return fmt.Sprintf(`<c%s/>`, attrs)
	}
	typ, val := "", ""
	switch c.val.Kind {
	case KindNumber:
		val = strconv.FormatFloat(c.val.Num, 'g', -1, 64)
	case KindString:
		typ, val = "str", c.val.Str
	case KindBool:
		typ, val = "b", "0"
		if c.val.Bool {
			val = "1"
		}
	case KindError:
		typ, val = "e", c.val.Err
	}
	if len(c.raw) > 1 && c.raw[0] == '=' {
		if typ != "" {
			attrs += ` t="` + typ + `"`
		}
		return fmt.Sprintf(`<c%s><f>%s</f><v>%s</v></c>`, attrs, xlxEsc(c.raw[1:]), xlxEsc(val))
	}
	switch typ {
	case "str":
		return fmt.Sprintf(`<c%s t="inlineStr"><is><t xml:space="preserve">%s</t></is></c>`, attrs, xlxEsc(val))
	case "":
		if val == "" {
			return fmt.Sprintf(`<c%s/>`, attrs)
		}
		return fmt.Sprintf(`<c%s><v>%s</v></c>`, attrs, val)
	}
	return fmt.Sprintf(`<c%s t="%s"><v>%s</v></c>`, attrs, typ, xlxEsc(val))
}

func commentsXML(s *Sheet) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`+"\n"+`<comments xmlns="%s"><authors><author>ownframe</author></authors><commentList>`, xlNSMain)
	for _, ref := range sortedRefs(s.comments) {
		fmt.Fprintf(&b, `<comment ref="%s" authorId="0"><text><r><t xml:space="preserve">%s</t></r></text></comment>`, ref.A1(), xlxEsc(s.comments[ref]))
	}
	b.WriteString(`</commentList></comments>`)
	return b.String()
}

func vmlXML(s *Sheet) string {
	var b strings.Builder
	b.WriteString(`<xml xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office" xmlns:x="urn:schemas-microsoft-com:office:excel">` +
		`<o:shapelayout v:ext="edit"><o:idmap v:ext="edit" data="1"/></o:shapelayout>` +
		`<v:shapetype id="_x0000_t202" coordsize="21600,21600" o:spt="202" path="m,l,21600r21600,l21600,xe"><v:stroke joinstyle="miter"/><v:path gradientshapeok="t" o:connecttype="rect"/></v:shapetype>`)
	for i, ref := range sortedRefs(s.comments) {
		fmt.Fprintf(&b, `<v:shape id="_x0000_s%d" type="#_x0000_t202" style="position:absolute;margin-left:80pt;margin-top:2pt;width:108pt;height:60pt;z-index:%d;visibility:hidden" fillcolor="#ffffe1" o:insetmode="auto">`+
			`<v:fill color2="#ffffe1"/><v:shadow on="t" color="black" obscured="t"/><v:path o:connecttype="none"/><v:textbox style="mso-direction-alt:auto"><div style="text-align:left"></div></v:textbox>`+
			`<x:ClientData ObjectType="Note"><x:MoveWithCells/><x:SizeWithCells/><x:Anchor>%d, 15, %d, 2, %d, 15, %d, 4</x:Anchor><x:AutoFill>False</x:AutoFill><x:Row>%d</x:Row><x:Column>%d</x:Column></x:ClientData></v:shape>`,
			xlVMLShapeBase+i, i+1, ref.Col+1, ref.Row, ref.Col+xlNoteCols, ref.Row+xlNoteRows, ref.Row, ref.Col)
	}
	b.WriteString(`</xml>`)
	return b.String()
}

// stylesXML renders styles.xml from the collected tables.
func (x *xlxWriter) stylesXML() string {
	var fontsX, fillsX, bordersX, xfX, fmtsX, dxfsX strings.Builder
	fontIdx, fillIdx, borderIdx, fmtIdx := map[string]int{}, map[string]int{}, map[Edges]int{}, map[string]int{}
	fonts, fills, borders, fmtCodes := 1, 2, 1, 0
	fontsX.WriteString(`<font><sz val="11"/><name val="Calibri"/></font>`)
	fillsX.WriteString(`<fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill>`)
	bordersX.WriteString(`<border><left/><right/><top/><bottom/><diagonal/></border>`)
	edge := func(tag, style string) string {
		if style == "" {
			return "<" + tag + "/>"
		}
		return fmt.Sprintf(`<%s style="%s"><color auto="1"/></%s>`, tag, style, tag)
	}
	for _, st := range x.xfs {
		var f strings.Builder
		if st.Bold {
			f.WriteString("<b/>")
		}
		if st.Italic {
			f.WriteString("<i/>")
		}
		if st.Strike {
			f.WriteString("<strike/>")
		}
		if st.Underline {
			f.WriteString("<u/>")
		}
		sz := st.FontSize
		if sz == 0 {
			sz = 11
		}
		fmt.Fprintf(&f, `<sz val="%d"/>`, sz)
		if c := xlColorToARGB(st.FG); c != "" {
			fmt.Fprintf(&f, `<color rgb="%s"/>`, c)
		}
		name := xlFontNames()[st.FontFamily]
		if name == "" {
			name = "Calibri"
		}
		fmt.Fprintf(&f, `<name val="%s"/>`, name)
		fkey := f.String()
		fi, ok := fontIdx[fkey]
		if !ok {
			fi = fonts
			fontIdx[fkey] = fi
			fonts++
			fontsX.WriteString("<font>" + fkey + "</font>")
		}
		bgi := 0
		if c := xlColorToARGB(st.BG); c != "" {
			bgi, ok = fillIdx[c]
			if !ok {
				bgi = fills
				fillIdx[c] = bgi
				fills++
				fmt.Fprintf(&fillsX, `<fill><patternFill patternType="solid"><fgColor rgb="%s"/><bgColor indexed="64"/></patternFill></fill>`, c)
			}
		}
		bi := 0
		if st.Edges.Any() {
			bi, ok = borderIdx[st.Edges]
			if !ok {
				bi = borders
				borderIdx[st.Edges] = bi
				borders++
				fmt.Fprintf(&bordersX, `<border>%s%s%s%s<diagonal/></border>`,
					edge("left", st.Edges.Left), edge("right", st.Edges.Right), edge("top", st.Edges.Top), edge("bottom", st.Edges.Bottom))
			}
		}
		nf := 0
		if code := xlNumFmtCode(st); code != "General" {
			nf, ok = fmtIdx[code]
			if !ok {
				nf = xlCustomFmtBase + fmtCodes
				fmtCodes++
				fmtIdx[code] = nf
				fmt.Fprintf(&fmtsX, `<numFmt numFmtId="%d" formatCode="%s"/>`, nf, xlxEsc(code))
			}
		}
		align := ""
		if st.Align != "" || st.Wrap || st.VAlign != "" || st.Indent > 0 {
			align = "<alignment"
			h := st.Align
			if h == "" && st.Indent > 0 {
				h = "left"
			}
			if h != "" {
				align += fmt.Sprintf(` horizontal="%s"`, h)
			}
			if v := xlVAlignOut()[st.VAlign]; v != "" {
				align += fmt.Sprintf(` vertical="%s"`, v)
			}
			if st.Wrap {
				align += ` wrapText="1"`
			}
			if st.Indent > 0 {
				align += fmt.Sprintf(` indent="%d"`, st.Indent)
			}
			align += "/>"
		}
		fmt.Fprintf(&xfX, `<xf numFmtId="%d" fontId="%d" fillId="%d" borderId="%d" applyNumberFormat="1" applyFont="1" applyFill="1" applyBorder="1" applyAlignment="1">%s</xf>`,
			nf, fi, bgi, bi, align)
	}
	for _, d := range x.dxfs {
		fmt.Fprintf(&dxfsX, `<dxf><font><color rgb="%s"/></font><fill><patternFill><bgColor rgb="%s"/></patternFill></fill></dxf>`, d[0], d[1])
	}
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`+"\n"+
		`<styleSheet xmlns="%s"><numFmts count="%d">%s</numFmts><fonts count="%d">%s</fonts><fills count="%d">%s</fills>`+
		`<borders count="%d">%s</borders>`+
		`<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>`+
		`<cellXfs count="%d"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>%s</cellXfs>`+
		`<dxfs count="%d">%s</dxfs></styleSheet>`,
		xlNSMain, fmtCodes, fmtsX.String(), fonts, fontsX.String(), fills, fillsX.String(), borders, bordersX.String(),
		len(x.xfs)+1, xfX.String(), len(x.dxfs), dxfsX.String())
}

// SaveXLSX writes an OpenXML workbook.
func (w *Workbook) SaveXLSX(file string) error {
	x := &xlxWriter{xfIdx: map[Style]int{}, dxfIdx: map[[2]string]int{}}
	type part struct{ name, body string }
	var parts []part
	var wbX, relX, ctX strings.Builder
	fmt.Fprintf(&wbX, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`+"\n"+`<workbook xmlns="%s" xmlns:r="%s"><bookViews><workbookView activeTab="%d"/></bookViews><sheets>`, xlNSMain, xlNSRel, w.Active)
	fmt.Fprintf(&relX, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`+"\n"+`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	ctX.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n" +
		`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` +
		`<Default Extension="xml" ContentType="application/xml"/><Default Extension="vml" ContentType="application/vnd.openxmlformats-officedocument.vmlDrawing"/>` +
		`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>` +
		`<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>`)
	for i, s := range w.Sheets {
		n := i + 1
		legacy := ""
		if len(s.comments) > 0 {
			legacy = "rId2"
			parts = append(parts,
				part{fmt.Sprintf("xl/comments%d.xml", n), commentsXML(s)},
				part{fmt.Sprintf("xl/drawings/vmlDrawing%d.vml", n), vmlXML(s)},
				part{fmt.Sprintf("xl/worksheets/_rels/sheet%d.xml.rels", n),
					`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n" +
						`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
						fmt.Sprintf(`<Relationship Id="rId1" Type="%s/comments" Target="../comments%d.xml"/>`, xlNSRel, n) +
						fmt.Sprintf(`<Relationship Id="rId2" Type="%s/vmlDrawing" Target="../drawings/vmlDrawing%d.vml"/>`, xlNSRel, n) +
						`</Relationships>`})
			fmt.Fprintf(&ctX, `<Override PartName="/xl/comments%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.comments+xml"/>`, n)
		}
		parts = append(parts, part{fmt.Sprintf("xl/worksheets/sheet%d.xml", n), x.sheetXML(s, legacy)})
		fmt.Fprintf(&wbX, `<sheet name="%s" sheetId="%d" r:id="rId%d"/>`, xlxEsc(s.Name), n, n)
		fmt.Fprintf(&relX, `<Relationship Id="rId%d" Type="%s/worksheet" Target="worksheets/sheet%d.xml"/>`, n, xlNSRel, n)
		fmt.Fprintf(&ctX, `<Override PartName="/xl/worksheets/sheet%d.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>`, n)
	}
	fmt.Fprintf(&relX, `<Relationship Id="rId%d" Type="%s/styles" Target="styles.xml"/></Relationships>`, len(w.Sheets)+1, xlNSRel)
	wbX.WriteString(`</sheets></workbook>`)
	ctX.WriteString(`</Types>`)
	rootRels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n" +
		`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="` + xlNSRel + `/officeDocument" Target="xl/workbook.xml"/></Relationships>`
	parts = append([]part{
		{"[Content_Types].xml", ctX.String()}, {"_rels/.rels", rootRels},
		{"xl/workbook.xml", wbX.String()}, {"xl/_rels/workbook.xml.rels", relX.String()}, {"xl/styles.xml", x.stylesXML()},
	}, parts...)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, p := range parts {
		f, err := zw.Create(p.name)
		if err != nil {
			return err
		}
		if _, err := f.Write([]byte(p.body)); err != nil {
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return os.WriteFile(file, buf.Bytes(), savedFileMode)
}

// ---- reading ----

type xlxRels struct {
	Rel []struct {
		ID     string `xml:"Id,attr"`
		Type   string `xml:"Type,attr"`
		Target string `xml:"Target,attr"`
	} `xml:"Relationship"`
}

type xlxBook struct {
	Sheets []struct {
		Name string `xml:"name,attr"`
		RID  string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
	} `xml:"sheets>sheet"`
	View struct {
		Active int `xml:"activeTab,attr"`
	} `xml:"bookViews>workbookView"`
}

type xlxText struct {
	T string `xml:"t"`
	R []struct {
		T string `xml:"t"`
	} `xml:"r"`
}

func (t xlxText) text() string {
	if len(t.R) == 0 {
		return t.T
	}
	var sb strings.Builder
	for _, r := range t.R {
		sb.WriteString(r.T)
	}
	return sb.String()
}

type xlxEdge struct {
	Style string `xml:"style,attr"`
}

type xlxStyles struct {
	NumFmts []struct {
		ID   int    `xml:"numFmtId,attr"`
		Code string `xml:"formatCode,attr"`
	} `xml:"numFmts>numFmt"`
	Fonts []struct {
		B      *struct{} `xml:"b"`
		I      *struct{} `xml:"i"`
		U      *struct{} `xml:"u"`
		Strike *struct{} `xml:"strike"`
		Sz     struct {
			Val float64 `xml:"val,attr"`
		} `xml:"sz"`
		Color struct {
			RGB string `xml:"rgb,attr"`
		} `xml:"color"`
		Name struct {
			Val string `xml:"val,attr"`
		} `xml:"name"`
	} `xml:"fonts>font"`
	Fills []struct {
		Pattern struct {
			Type string `xml:"patternType,attr"`
			FG   struct {
				RGB string `xml:"rgb,attr"`
			} `xml:"fgColor"`
		} `xml:"patternFill"`
	} `xml:"fills>fill"`
	Borders []struct {
		Left   xlxEdge `xml:"left"`
		Right  xlxEdge `xml:"right"`
		Top    xlxEdge `xml:"top"`
		Bottom xlxEdge `xml:"bottom"`
	} `xml:"borders>border"`
	Xfs []struct {
		NumFmt int `xml:"numFmtId,attr"`
		Font   int `xml:"fontId,attr"`
		Fill   int `xml:"fillId,attr"`
		Border int `xml:"borderId,attr"`
		Align  struct {
			H      string `xml:"horizontal,attr"`
			V      string `xml:"vertical,attr"`
			Wrap   string `xml:"wrapText,attr"`
			Indent int    `xml:"indent,attr"`
		} `xml:"alignment"`
	} `xml:"cellXfs>xf"`
	Dxfs []struct {
		FontColor struct {
			RGB string `xml:"rgb,attr"`
		} `xml:"font>color"`
		Fill struct {
			BG struct {
				RGB string `xml:"rgb,attr"`
			} `xml:"bgColor"`
			FG struct {
				RGB string `xml:"rgb,attr"`
			} `xml:"fgColor"`
		} `xml:"fill>patternFill"`
	} `xml:"dxfs>dxf"`
}

type xlxCFRule struct {
	Type     string   `xml:"type,attr"`
	Operator string   `xml:"operator,attr"`
	Text     string   `xml:"text,attr"`
	Bottom   string   `xml:"bottom,attr"`
	Above    string   `xml:"aboveAverage,attr"`
	Rank     string   `xml:"rank,attr"`
	DxfID    string   `xml:"dxfId,attr"`
	Formula  []string `xml:"formula"`
	DataBar  *struct {
		Color struct {
			RGB string `xml:"rgb,attr"`
		} `xml:"color"`
	} `xml:"dataBar"`
	ColorScale *struct{} `xml:"colorScale"`
}

type xlxSheet struct {
	TabColor struct {
		RGB string `xml:"rgb,attr"`
	} `xml:"sheetPr>tabColor"`
	Pane struct {
		X string `xml:"xSplit,attr"`
		Y string `xml:"ySplit,attr"`
	} `xml:"sheetViews>sheetView>pane"`
	Cols []struct {
		Min    int     `xml:"min,attr"`
		Max    int     `xml:"max,attr"`
		Width  float64 `xml:"width,attr"`
		Hidden string  `xml:"hidden,attr"`
	} `xml:"cols>col"`
	Rows []struct {
		R      int     `xml:"r,attr"`
		Ht     float64 `xml:"ht,attr"`
		Hidden string  `xml:"hidden,attr"`
		C      []struct {
			R  string  `xml:"r,attr"`
			S  int     `xml:"s,attr"`
			T  string  `xml:"t,attr"`
			F  string  `xml:"f"`
			V  string  `xml:"v"`
			IS xlxText `xml:"is"`
		} `xml:"c"`
	} `xml:"sheetData>row"`
	Merges []struct {
		Ref string `xml:"ref,attr"`
	} `xml:"mergeCells>mergeCell"`
	CF []struct {
		Sqref string      `xml:"sqref,attr"`
		Rules []xlxCFRule `xml:"cfRule"`
	} `xml:"conditionalFormatting"`
	DV []struct {
		Type     string `xml:"type,attr"`
		Operator string `xml:"operator,attr"`
		Sqref    string `xml:"sqref,attr"`
		F1       string `xml:"formula1"`
		F2       string `xml:"formula2"`
	} `xml:"dataValidations>dataValidation"`
}

func xlxTrue(v string) bool { return v == "1" || v == "true" }

// xlxCondFromRule converts a cfRule back into a CondRule. dxfPreset names the
// preset whose colors match the rule's dxf.
func xlxCondFromRule(r xlxCFRule, dxfPreset string) (CondRule, bool) {
	unq := func(s string) string {
		s = strings.TrimSpace(s)
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			s = strings.ReplaceAll(s[1:len(s)-1], `""`, `"`)
		}
		return s
	}
	arg := func(i int) string {
		if i < len(r.Formula) {
			return unq(r.Formula[i])
		}
		return ""
	}
	switch r.Type {
	case "cellIs":
		for kind, op := range xlCondOps() {
			if op == r.Operator {
				return CondRule{Kind: kind, A: arg(0), B: arg(1), Preset: dxfPreset}, true
			}
		}
	case "containsText":
		return CondRule{Kind: "text", A: r.Text, Preset: dxfPreset}, true
	case "duplicateValues":
		return CondRule{Kind: "dupes", Preset: dxfPreset}, true
	case "top10":
		kind := "top"
		if xlxTrue(r.Bottom) {
			kind = "bottom"
		}
		return CondRule{Kind: kind, A: r.Rank, Preset: dxfPreset}, true
	case "aboveAverage":
		kind := "above"
		if r.Above == "0" || r.Above == "false" {
			kind = "below"
		}
		return CondRule{Kind: kind, Preset: dxfPreset}, true
	case "dataBar":
		preset := "blue"
		if r.DataBar != nil {
			bar := xlARGBToColor(r.DataBar.Color.RGB)
			for name, p := range xlPresets() {
				if strings.EqualFold(p.bar, bar) {
					preset = name
				}
			}
		}
		return CondRule{Kind: "databar", Preset: preset}, true
	case "colorScale":
		return CondRule{Kind: "scale3"}, true
	}
	return CondRule{}, false
}

// LoadXLSX reads a workbook written by SaveXLSX or a typical spreadsheet app.
func LoadXLSX(file string) (*Workbook, error) {
	zr, err := zip.OpenReader(file)
	if err != nil {
		return nil, err
	}
	defer func() { _ = zr.Close() }()
	files := map[string]*zip.File{}
	for _, f := range zr.File {
		files[f.Name] = f
	}
	read := func(name string, v any) error {
		f := files[name]
		if f == nil {
			return fmt.Errorf("xlsx: missing %s", name)
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() { _ = rc.Close() }()
		return xml.NewDecoder(rc).Decode(v)
	}
	var book xlxBook
	if err := read("xl/workbook.xml", &book); err != nil {
		return nil, err
	}
	var rels xlxRels
	if err := read("xl/_rels/workbook.xml.rels", &rels); err != nil {
		return nil, err
	}
	targets := map[string]string{}
	for _, r := range rels.Rel {
		t := r.Target
		if strings.HasPrefix(t, "/") {
			targets[r.ID] = strings.TrimPrefix(t, "/")
		} else {
			targets[r.ID] = path.Join("xl", t)
		}
	}
	var shared []string
	if files["xl/sharedStrings.xml"] != nil {
		var ss struct {
			SI []xlxText `xml:"si"`
		}
		if err := read("xl/sharedStrings.xml", &ss); err != nil {
			return nil, err
		}
		for _, si := range ss.SI {
			shared = append(shared, si.text())
		}
	}
	var styles xlxStyles
	if files["xl/styles.xml"] != nil {
		if err := read("xl/styles.xml", &styles); err != nil {
			return nil, err
		}
	}
	custom := map[int]string{}
	for _, nf := range styles.NumFmts {
		custom[nf.ID] = nf.Code
	}
	edgeOf := func(e xlxEdge) string {
		switch e.Style {
		case "":
			return ""
		case "thick":
			return EdgeThick
		case "double":
			return EdgeDouble
		}
		return EdgeThin
	}
	styleOf := func(i int) Style {
		var st Style
		if i <= 0 || i >= len(styles.Xfs) {
			return st
		}
		xf := styles.Xfs[i]
		if xf.Font >= 0 && xf.Font < len(styles.Fonts) {
			f := styles.Fonts[xf.Font]
			st.Bold, st.Italic, st.Underline, st.Strike = f.B != nil, f.I != nil, f.U != nil, f.Strike != nil
			st.FG = xlARGBToColor(f.Color.RGB)
			if sz := int(f.Sz.Val); sz != 0 && sz != 11 {
				st.FontSize = sz
			}
			st.FontFamily = xlFamilyOf(f.Name.Val)
		}
		if xf.Fill >= 0 && xf.Fill < len(styles.Fills) && styles.Fills[xf.Fill].Pattern.Type == "solid" {
			st.BG = xlARGBToColor(styles.Fills[xf.Fill].Pattern.FG.RGB)
		}
		if xf.Border > 0 && xf.Border < len(styles.Borders) {
			b := styles.Borders[xf.Border]
			st.Edges = Edges{Left: edgeOf(b.Left), Right: edgeOf(b.Right), Top: edgeOf(b.Top), Bottom: edgeOf(b.Bottom)}
		}
		code, ok := custom[xf.NumFmt]
		if !ok {
			code = xlBuiltinFmts()[xf.NumFmt]
		}
		st.NumFmt, st.Decimals = xlStyleFromCode(code)
		st.Align = xf.Align.H
		if st.Align == "general" {
			st.Align = ""
		}
		st.VAlign = xlVAlignIn()[xf.Align.V]
		st.Indent = clampIndent(xf.Align.Indent)
		st.Wrap = xlxTrue(xf.Align.Wrap)
		return st
	}
	dxfPreset := func(id string) string {
		i, err := strconv.Atoi(id)
		if err != nil || i < 0 || i >= len(styles.Dxfs) {
			return "red"
		}
		d := styles.Dxfs[i]
		bg := xlARGBToColor(d.Fill.BG.RGB)
		if bg == "" {
			bg = xlARGBToColor(d.Fill.FG.RGB)
		}
		for name, p := range xlPresets() {
			if strings.EqualFold(p.bg, bg) {
				return name
			}
		}
		return "red"
	}
	w := &Workbook{Active: book.View.Active}
	for _, sh := range book.Sheets {
		s := newSheet(w, sh.Name)
		w.Sheets = append(w.Sheets, s)
		sheetPath := targets[sh.RID]
		var x xlxSheet
		if err := read(sheetPath, &x); err != nil {
			return nil, err
		}
		s.TabColor = xlARGBToColor(x.TabColor.RGB)
		for _, c := range x.Cols {
			for col := c.Min; col <= c.Max && col <= MaxCols; col++ {
				s.colW[col-1] = int(c.Width*xlColCharPx + xlRoundHalf)
				if xlxTrue(c.Hidden) {
					s.hiddenCols[col-1] = true
				}
			}
		}
		for _, r := range x.Rows {
			if r.Ht > 0 && r.R >= 1 {
				s.rowH[r.R-1] = int(r.Ht/xlPxToPt + xlRoundHalf)
			}
			if xlxTrue(r.Hidden) && r.R >= 1 {
				s.hiddenRows[r.R-1] = true
			}
			for _, c := range r.C {
				ref, ok := ParseA1(c.R)
				if !ok {
					continue
				}
				if st := styleOf(c.S); st != (Style{}) {
					s.styles[ref] = st
				}
				var raw string
				switch {
				case c.F != "":
					raw = "=" + c.F
				case c.T == "s":
					if i, err := strconv.Atoi(strings.TrimSpace(c.V)); err == nil && i >= 0 && i < len(shared) {
						raw = xlValueRaw(xlStrV(shared[i]))
					}
				case c.T == "inlineStr":
					raw = xlValueRaw(xlStrV(c.IS.text()))
				case c.T == "str":
					raw = xlValueRaw(xlStrV(c.V))
				case c.T == "b":
					raw = "FALSE"
					if strings.TrimSpace(c.V) == "1" {
						raw = "TRUE"
					}
				case c.T == "e":
					raw = "'" + c.V
				default:
					raw = strings.TrimSpace(c.V)
				}
				s.setRawQuiet(ref, raw)
			}
		}
		for _, m := range x.Merges {
			if r, ok := ParseRange(m.Ref); ok {
				s.merges = append(s.merges, r)
			}
		}
		s.freezeC, _ = strconv.Atoi(x.Pane.X)
		s.freezeR, _ = strconv.Atoi(x.Pane.Y)
		for _, cf := range x.CF {
			for _, rule := range cf.Rules {
				cr, ok := xlxCondFromRule(rule, dxfPreset(rule.DxfID))
				if !ok {
					continue
				}
				for _, part := range strings.Fields(cf.Sqref) {
					if r, ok := ParseRange(part); ok {
						s.AddCondFormat(r, cr)
					}
				}
			}
		}
		for _, dv := range x.DV {
			for _, part := range strings.Fields(dv.Sqref) {
				r, ok := ParseRange(part)
				if !ok {
					continue
				}
				switch dv.Type {
				case "list":
					f := strings.TrimSpace(dv.F1)
					if len(f) >= 2 && f[0] == '"' && f[len(f)-1] == '"' {
						var list []string
						for _, o := range strings.Split(f[1:len(f)-1], ",") {
							list = append(list, strings.TrimSpace(o))
						}
						s.SetValidation(r, list)
					}
				case "whole", "decimal", "textLength":
					if dv.Operator != "" && dv.Operator != "between" {
						continue
					}
					lo, _ := strconv.ParseFloat(strings.TrimSpace(dv.F1), 64)
					hi, _ := strconv.ParseFloat(strings.TrimSpace(dv.F2), 64)
					kind := dv.Type
					if kind == "textLength" {
						kind = "length"
					}
					s.SetValidationBounds(r, kind, lo, hi)
				}
			}
		}
		// comments live in a part named by the sheet's relationships
		relPath := path.Join(path.Dir(sheetPath), "_rels", path.Base(sheetPath)+".rels")
		if files[relPath] != nil {
			var sr xlxRels
			if err := read(relPath, &sr); err == nil {
				for _, r := range sr.Rel {
					if !strings.HasSuffix(r.Type, "/comments") {
						continue
					}
					cpath := path.Join(path.Dir(sheetPath), r.Target)
					if strings.HasPrefix(r.Target, "/") {
						cpath = strings.TrimPrefix(r.Target, "/")
					}
					var cm struct {
						List []struct {
							Ref  string  `xml:"ref,attr"`
							Text xlxText `xml:"text"`
						} `xml:"commentList>comment"`
					}
					if err := read(cpath, &cm); err != nil {
						continue
					}
					for _, c := range cm.List {
						if ref, ok := ParseA1(c.Ref); ok {
							s.SetComment(ref.Row, ref.Col, c.Text.text())
						}
					}
				}
			}
		}
	}
	if len(w.Sheets) == 0 {
		return nil, fmt.Errorf("xlsx: no sheets")
	}
	w.Active = max(0, min(w.Active, len(w.Sheets)-1))
	w.reloadAll()
	return w, nil
}
