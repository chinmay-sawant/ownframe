package window

import (
	"encoding/json"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/replay"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image"
	"image/color"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// devOpsHalf is the half-pixel reach of the 1 px outline; devOpsFull
// grows the union box by the full pixel.
const (
	devOpsHalf = 0.5
	devOpsFull = 1.0
)

// drawDevOps outlines every painted operation in paint order, colour by
// kind, culled to the page area the dock leaves visible. Noop and unknown
// operations carry no paint and get no outline. Only the strokes are culled;
// the row list still holds every operation, so picking is not affected.
// The outlines are one chrome region: the union of the strokes that were
// actually painted is noted once.
func (s *shell) drawDevOps(screen *ebiten.Image) {
	if s.display == nil {
		return
	}

	view := s.devOpsView()
	var union image.Rectangle

	for _, index := range s.display.Order {
		if index < 0 || index >= len(s.display.Ops) {
			continue
		}

		op := &s.display.Ops[index]
		ink, ok := devOpPaintInk(op)
		if !ok {
			continue
		}

		r := s.devScreen(devOpRect(op, s.display.PointsPerPixel))
		if !devRectsMeet(r, view) {
			continue
		}

		if s.dev.stroke != nil {
			s.dev.stroke(screen, r, ink)
		} else {
			vector.StrokeRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), 1, ink, false)
		}

		// The 1 px stroke reaches half a pixel past each edge.
		union = union.Union(frameRect(r.X-devOpsHalf, r.Y-devOpsHalf, r.W+devOpsFull, r.H+devOpsFull))
	}

	if !union.Empty() {
		s.frameNote(union)
	}
}

// devOpTextCap caps the quoted text runes in an Ops row.
const devOpTextCap = 40

// devOpRow builds one Ops-tab row for the operation at index i, the nth
// painted operation.
func devOpRow(op *layout.DisplayOp, i, n int, ink color.RGBA, ppt float64) devRow {
	x := devOpRect(op, ppt)
	sp := []devSpan{
		{text: fmt.Sprintf("#%d ", n), ink: devDim},
		{text: devOpsListName[op.Kind], ink: ink},
		{text: fmt.Sprintf(" x%g y%g %gx%g", x.X, x.Y, x.W, x.H), ink: devDim},
	}
	if op.Kind == layout.DisplayOpText || op.Kind == layout.DisplayOpBullet {
		t := op.Text
		if len([]rune(t)) > devOpTextCap {
			t = string([]rune(t)[:devOpTextCap]) + "..."
		}
		sp = append(sp, devSpan{text: fmt.Sprintf(" %q", t), ink: devStrInk})
	}

	return devRow{line: devLine{spans: sp}, act: devActOp, arg: i}
}

// devOpsView is the page area an outline may land in: the window minus the
// dock, so a stroke under the panel or off the window is skipped. The dock
// covering the whole window leaves it empty.
func (s *shell) devOpsView() devRect {
	if s.dev.mobile {
		return devRect{W: float64(s.screenW), H: s.dev.panel.Y}
	}
	if s.dev.panel.W <= 0 {
		return devRect{W: float64(s.screenW), H: float64(s.screenH)}
	}

	return devRect{W: s.dev.panel.X, H: float64(s.screenH)}
}

// devRectsMeet reports whether two screen rects overlap.
func devRectsMeet(a, b devRect) bool {
	return a.X < b.X+b.W && b.X < a.X+a.W && a.Y < b.Y+b.H && b.Y < a.Y+a.H
}

// devOpsListName is the per-kind label. A corner-shape path is one label
// whether the op is the fill or the border, matching its one kind.
var devOpsListName = [...]string{"", "fill", "stroke", "line", "text", "image", "", "text", "grid", "path"}

// devOpsBuildRows builds the cached Ops rows.
func (s *shell) devOpsBuildRows() []devRow {
	mark := "[ ]"
	if s.dev.ops {
		mark = "[x]"
	}
	r := []devRow{{line: devLine{spans: []devSpan{{text: mark, ink: devAccent}, {text: " Show paint outlines (o)", ink: devFg}}}, act: devActOpsToggle}}
	h := func(t string) { r = append(r, devRow{}, devRow{line: devText(t, devDim)}) }
	if s.display == nil {
		h("bitmap fallback")
		return r
	}

	h("DISPLAY")
	c := devCountOps(s.display)
	v := [7]int{c.Fill, c.Stroke, c.Line, c.Text, c.Image, c.Grid, c.Path}
	k := [...]layout.DisplayKind{1, 2, 3, 4, 5, 8, 9}
	for i, kind := range k {
		ink, _ := devOpInk(kind)
		lbl, num := devFg, devNumInk
		if v[i] == 0 {
			lbl, num = devDim, devDim
		}
		r = append(r, devRow{line: devLine{spans: []devSpan{devSwatch(ink), {text: " " + devOpsListName[kind], ink: lbl}, {text: fmt.Sprint(" ", v[i]), ink: num}}}})
	}
	r = append(r, devRow{line: devLine{spans: []devSpan{{text: "total", ink: devDim}, {text: fmt.Sprint(" ", c.total()), ink: devNumInk}}}})

	h("PAINT ORDER")
	n := 0
	for _, i := range s.display.Order {
		if i < 0 || i >= len(s.display.Ops) {
			continue
		}
		op := &s.display.Ops[i]
		ink, ok := devOpPaintInk(op)
		if !ok {
			continue
		}
		n++
		r = append(r, devOpRow(op, i, n, ink, s.display.PointsPerPixel))
	}
	if n == 0 {
		r = append(r, devRow{line: devText("(none)", devDim)})
	}
	return r
}

// devOpPickWidth is the selection outline width; the stroke reaches half
// past each edge.
const (
	devOpPickWidth  = 3.0
	devOpPickExpand = 1.5
)

// drawDevOpPick outlines the operation chosen in the Ops list in its kind
// colour, thicker than the ops view so it reads as the selection.
func (s *shell) drawDevOpPick(screen *ebiten.Image) {
	if s.display == nil || s.dev.opPick < 0 || s.dev.opPick >= len(s.display.Ops) {
		return
	}

	op := &s.display.Ops[s.dev.opPick]

	ink, ok := devOpPaintInk(op)
	if !ok {
		return
	}

	r := s.devScreen(devOpRect(op, s.display.PointsPerPixel))
	vector.StrokeRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), devOpPickWidth, ink, false)
	// The 3 px stroke reaches 1.5 px past each edge.
	s.frameNote(frameRect(r.X-devOpPickExpand, r.Y-devOpPickExpand, r.W+devOpPickWidth, r.H+devOpPickWidth))
}

// devOpInk returns the outline colour for one operation kind.
func devOpInk(kind layout.DisplayKind) (color.RGBA, bool) {
	switch kind {
	case layout.DisplayOpFillRect:
		return devFillInk, true
	case layout.DisplayOpStrokeRect:
		return devStrokeInk, true
	case layout.DisplayOpLine:
		return devLineInk, true
	case layout.DisplayOpText, layout.DisplayOpBullet:
		return devTextInk, true
	case layout.DisplayOpImage:
		return devImageInk, true
	case layout.DisplayOpGridRun:
		return devGridInk, true
	case layout.DisplayOpPath:
		return devFillInk, true
	}

	return color.RGBA{}, false
}

// devOpPaintInk is devOpInk for one operation, so a corner-shape border path
// shows the stroke colour the same kind's fill path shows the fill colour.
// Both share DisplayOpPath, which the count row can only label once.
func devOpPaintInk(op *layout.DisplayOp) (color.RGBA, bool) {
	if op.Kind == layout.DisplayOpPath {
		if corner, ok := op.CornerPathValue(); ok && corner != nil && corner.Stroke {
			return devStrokeInk, true
		}
	}

	return devOpInk(op.Kind)
}

// devOpRect is one operation's painted box in canvas CSS pixels: where the
// replay draws its ink, so an outline sits on the paint. replay.PaintBounds
// owns the rule, which covers the baseline of a text run, the stroke width of
// a border, the inward geometry of a mixed border line, and the segments of a
// table grid. An op that cannot be bounded, such as a transformed run, falls
// back to its nominal box converted from points.
func devOpRect(op *layout.DisplayOp, ppp float64) devRect {
	if ppp <= 0 {
		return devRect{}
	}

	box, ok := replay.PaintBounds(op, ppp)
	if !ok || box.Empty() {
		return devNominalRect(op, ppp)
	}

	return devRect{
		X: float64(box.Min.X), Y: float64(box.Min.Y),
		W: float64(box.Dx()), H: float64(box.Dy()),
	}
}

// devNominalRect converts an op's raw box from points to CSS pixels. A text
// op carries its baseline in Y, so its nominal box starts one line height
// above it.
func devNominalRect(op *layout.DisplayOp, ppp float64) devRect {
	x, y, w, h := op.X, op.Y, op.W, op.H

	if op.Kind == layout.DisplayOpText || op.Kind == layout.DisplayOpBullet {
		y = op.Y + op.InkDescent - op.H
	}

	return devRect{X: x / ppp, Y: y / ppp, W: w / ppp, H: h / ppp}
}

// devElementClip caps the element text runes in the panel.
const devElementClip = 120

// devElementRows returns the Elements tab: the picked target above its
// properties as pretty JSON. The pinned box wins over the hovered one; with
// neither, the tab explains how to pick.
func (s *shell) devElementRows() []devRow {
	box, ok := s.devBox()
	if !ok {
		return []devRow{
			{line: devPlain("Select an element")},
			{},
			{line: devText("Click a box in the page to pin it.", devDim)},
			{line: devText("Hover to preview.", devDim)},
		}
	}

	label, ink := "hovering", devDim
	if s.dev.havePin {
		label, ink = "pinned", devAccent
	}

	rows := []devRow{{line: devText(label+"  "+devBoxName(box), ink)}}
	rows = append(rows, devRow{}, devRow{line: devText("properties", devDim)}, devRow{})
	rows = append(rows, devJSONRows(s.devElementPayload(box), s.dev.collapsed)...)

	return rows
}

// devBox returns the picked box, pinned first.
func (s *shell) devBox() (layout.Box, bool) {
	if s.dev.havePin {
		return s.dev.pinned, true
	}

	if s.dev.haveHov {
		return s.dev.hovered, true
	}

	return layout.Box{}, false
}

// devElementPayload is the JSON payload for one box.
func (s *shell) devElementPayload(box layout.Box) devElementJSON {
	ops := 0
	if s.display != nil {
		ops = devOpsInBox(s.display, box)
	}

	return devElementJSON{
		Tag:    box.Tag,
		ID:     box.ID,
		Action: box.Action,
		Text:   devClip(box.Text, devElementClip),
		Rect: devElementRect{
			X: box.X, Y: box.Y, Width: box.W, Height: box.H,
		},
		Ops: ops,
	}
}

// devElementJSON is the picked element on the Elements tab.
type devElementJSON struct {
	Tag    string         `json:"tag"`
	ID     string         `json:"id"`
	Action string         `json:"action"`
	Text   string         `json:"text"`
	Rect   devElementRect `json:"rect"`
	Ops    int            `json:"ops"`
}

// devElementRect is an element border box in CSS pixels.
type devElementRect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// devPointer routes the pointer while the overlay is on. The dock owns its
// hits and its resize edge; the content pins the box under the cursor.
// Alt+click forwards press, caret, and click to the page.
func (s *shell) devPointer(x, y int, px, py float64) error {
	if err := s.devPointerHover(x, y, px, py); err != nil {
		return err
	}
	down := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	if handled, err := s.devMenuPointer(x, y,
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight),
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)); handled || err != nil {
		s.mouseDown = down
		return err
	}
	released := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	pressed := pressEdge(down, released, s.mouseDown)
	s.mouseDown = down

	if s.dev.resizing {
		if down {
			s.devResize(x)

			return nil
		}

		s.dev.resizing = false
	}
	// The vertical thumb overlaps the dock resize strip by 4 px. The thumb
	// wins there, so a scrollbar grab never resizes the dock for no reason.
	// An active resize keeps its gesture; a started thumb drag continues
	// here because the dock path bypasses pointerMouse.
	if s.pointerScrollbar(x, y, down, pressed) {
		return nil
	}
	if s.dev.forward && !pressed {
		return s.devForwardDrag(px, py, down, released)
	}

	if !s.dev.mobile && pressed && devOnEdge(s.dev.panel, x, y) {
		s.dev.resizing = true

		return nil
	}

	if devInRect(s.dev.panel, float64(x), float64(y)) {
		if pressed {
			s.devPanelPress(x, y)
		}

		return nil
	}

	if pressed {
		if err := s.devPick(px, py, ebiten.IsKeyPressed(ebiten.KeyAlt)); err != nil {
			return err
		}

		if released && s.dev.forward {
			s.dev.forward = false

			return s.releaseAt()
		}

		return nil
	}

	return nil
}

func (s *shell) devForwardDrag(x, y float64, down, released bool) error {
	if down {
		return s.dragAt(x, y)
	}
	if released {
		s.dev.forward = false
		return s.releaseAt()
	}
	return nil
}

// devMenuPointer gives the menu content clicks while keeping dock input local.
func (s *shell) devMenuPointer(x, y int, right, left bool) (bool, error) {
	if s.menu.open && left {
		return s.menuClick(x, y)
	}
	if right && !s.dev.resizing &&
		!devInRect(s.dev.panel, float64(x), float64(y)) &&
		(s.dev.mobile || !devOnEdge(s.dev.panel, x, y)) {
		s.openMenu(x, y)
		return true, nil
	}
	return false, nil
}

// devPointerHover keeps normal page hover and cursor feedback unless the
// pointer belongs to dock chrome or the explicitly enabled picker.
func (s *shell) devPointerHover(x, y int, px, py float64) error {
	if s.dev.resizing || !s.dev.mobile && devOnEdge(s.dev.panel, x, y) {
		// The thumb overlaps the edge strip. Show the thumb cursor there
		// to match the click, which scrolls instead of resizing.
		if axis, _, ok := s.scrollbarGrab(x, y); ok {
			s.applyThumbCursor(axis)
			return nil
		}
		s.setCursorShape(ebiten.CursorShapeEWResize)
		return nil
	}
	if devInRect(s.dev.panel, float64(x), float64(y)) {
		s.setCursorShape(ebiten.CursorShapeDefault)
		return nil
	}
	if s.devPickerOn() {
		s.setCursorShape(ebiten.CursorShapeDefault)
		s.devHover(px, py)
		return nil
	}
	return s.hoverCursor(x, y, px, py)
}

// devOpAt returns the topmost painted operation whose box holds the point
// and whose centre sits in the picked box, or -1 when none does. The point
// rule keeps a click on a title, an image, or a row's background on the
// operation the pointer actually touched.
func devOpAt(display *layout.Display, box layout.Box, x, y float64) int {
	if display == nil {
		return -1
	}

	found := -1

	for _, i := range display.Order {
		if i < 0 || i >= len(display.Ops) {
			continue
		}

		op := &display.Ops[i]
		if _, ok := devOpInk(op.Kind); !ok {
			continue
		}

		r := devOpRect(op, display.PointsPerPixel)
		if x < r.X || y < r.Y || x > r.X+r.W || y > r.Y+r.H {
			continue
		}

		cx, cy := r.X+r.W/devCenterDiv, r.Y+r.H/devCenterDiv
		if cx >= box.X-0.5 && cx <= box.X+box.W+0.5 &&
			cy >= box.Y-0.5 && cy <= box.Y+box.H+0.5 {
			found = i
		}
	}

	return found
}

// devPick handles a press in the content while the overlay is on. Alt
// forwards press, caret, and click to the page; otherwise the press acts on
// the control under the cursor, pinning its box in picker mode.
func (s *shell) devPick(px, py float64, alt bool) error {
	if alt || !s.devPickerOn() {
		s.dev.forward = true
		return s.pressAt(px, py)
	}

	box, ok := s.boxAt(px, py)
	if !ok {
		s.devDropPin()

		return nil
	}

	if s.dev.havePin && devSameBox(s.dev.pinned, box) {
		s.devDropPin()

		return nil
	}

	s.dev.pinned = box
	s.dev.havePin = true
	s.devShowBoxOps(box, px, py)

	return nil
}

// devPickerOn reports whether the element picker mode is active.
func (s *shell) devPickerOn() bool {
	return s.dev.on && s.dev.picker
}

// devSetPicker changes the explicit picker selection. Turning it off drops
// the hovered and pinned picks and the operation selected through the picker.
func (s *shell) devSetPicker(on bool) {
	s.dev.picker = on && s.dev.on

	if !s.dev.picker {
		s.dev.haveHov = false
		s.dev.havePin = false

		if s.dev.opPin {
			s.dev.haveOp = false
			s.dev.opPin = false
		}
	}
}

// drawDevPick outlines the hovered and pinned boxes, one colour each, with a
// label over each.
func (s *shell) drawDevPick(screen *ebiten.Image) {
	if !s.devPickerOn() {
		return
	}

	if s.dev.haveHov {
		s.devOutline(screen, s.dev.hovered, devHoverInk)
	}

	if s.dev.havePin {
		s.devOutline(screen, s.dev.pinned, devPinInk)
	}
}

func (s *shell) devPickerHit() {
	_, px := devHeaderRight(s.dev.panel)
	s.dev.hits = append(s.dev.hits, devHit{
		rect: devPickerRect(px, s.dev.panel.Y),
		act:  devActPicker,
	})
}

// devOpInBox returns the display index of the topmost painted operation
// whose centre lies inside box, or -1. The centre rule matches internal/frame
// and tolerates a stroke or an ink bound just past the element edge.
func devOpInBox(display *layout.Display, box layout.Box) int {
	if display == nil {
		return -1
	}

	found := -1

	for _, i := range display.Order {
		if i < 0 || i >= len(display.Ops) {
			continue
		}

		op := &display.Ops[i]
		if _, ok := devOpInk(op.Kind); !ok {
			continue
		}

		r := devOpRect(op, display.PointsPerPixel)
		cx, cy := r.X+r.W/devCenterDiv, r.Y+r.H/devCenterDiv
		if cx >= box.X-0.5 && cx <= box.X+box.W+0.5 &&
			cy >= box.Y-0.5 && cy <= box.Y+box.H+0.5 {
			found = i
		}
	}

	return found
}

// devShowBoxOps opens the Ops tab for a pinned box, selects the operation
// under the click when one sits there and the operation that fills the box
// otherwise, and scrolls its row into view.
func (s *shell) devShowBoxOps(box layout.Box, px, py float64) {
	s.dev.tab = devTabOps

	i := devOpAt(s.display, box, px, py)
	if i < 0 {
		i = devOpInBox(s.display, box)
	}

	if i < 0 {
		return
	}

	s.dev.opPick, s.dev.haveOp, s.dev.opPin = i, true, true
	s.devScrollToOp(i)
}

// devDropPin clears the pinned box and the operation the pick selected.
func (s *shell) devDropPin() {
	s.dev.havePin = false

	if s.dev.opPin {
		s.dev.haveOp = false
		s.dev.opPin = false
	}
}

// devScrollToOp moves the dock content so the row for an operation shows.
func (s *shell) devScrollToOp(i int) {
	c := s.dev.content
	if c.H <= 0 {
		return
	}

	lineH := devLineHeight()
	rows := s.devOpsRows()
	s.dev.rows = rows

	for n, row := range rows {
		if row.act != devActOp || row.arg != i {
			continue
		}

		top := float64(n) * lineH
		if top < s.dev.scroll {
			s.dev.scroll = top
		}

		if top+lineH > s.dev.scroll+c.H {
			s.dev.scroll = top + lineH - c.H
		}

		s.devClampScroll()

		return
	}
}

// devJSONRows renders v as pretty-printed JSON rows with syntax colours.
// Structs keep declaration order, maps render keys sorted, slices render in
// order, and pointers or interfaces render their contents. An object or
// array opens on a row carrying devActJSON and the node's dotted key path,
// so a click can flip collapsed[key] to fold the node. A folded node renders
// as {...} or [...] on one row.
func devJSONRows(v any, collapsed map[string]bool) []devRow {
	p := &devJSONPrinter{collapsed: collapsed}
	p.emit(p.build(reflect.ValueOf(v), "", ""), 0, devLine{}, "", true)

	return p.rows
}

// buildStruct renders exported fields in declaration order.
func (p *devJSONPrinter) buildStruct(v reflect.Value, key, path string) *devJSONNode {
	node := &devJSONNode{kind: 'o', key: key, path: path}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}

		name, opts := devJSONTag(f.Tag.Get("json"))
		if name == "-" {
			continue
		}

		if name == "" {
			name = f.Name
		}

		fv := v.Field(i)
		if opts == "omitempty" && devJSONEmpty(fv) {
			continue
		}

		node.kids = append(node.kids, p.build(fv, name, devJSONJoin(path, name)))
	}

	return node
}

// buildArray renders elements in order, named by index.
func (p *devJSONPrinter) buildArray(v reflect.Value, key, path string) *devJSONNode {
	node := &devJSONNode{kind: 'a', key: key, path: path}

	for i := 0; i < v.Len(); i++ {
		ik := strconv.Itoa(i)
		node.kids = append(node.kids, p.build(v.Index(i), ik, devJSONJoin(path, ik)))
	}

	return node
}

// devJSONPrinter accumulates the rows for one document.
type devJSONPrinter struct {
	rows      []devRow
	collapsed map[string]bool
}

// emit appends the rows for one node. depth is the nesting level, lead is
// the member prefix, tail is a trailing comma, and root keeps the top node
// unfoldable.
func (p *devJSONPrinter) emit(n *devJSONNode, depth int, lead devLine, tail string, root bool) {
	indent := strings.Repeat("  ", depth)

	if n.kind == 's' {
		p.rows = append(p.rows, devRow{line: devJSONLine(indent, lead, n.scalar, n.ink, tail)})

		return
	}

	open, close := "{", "}"
	if n.kind == 'a' {
		open, close = "[", "]"
	}

	if len(n.kids) == 0 {
		p.rows = append(p.rows, devRow{line: devJSONLine(indent, lead, open+close, devPunctInk, tail)})

		return
	}

	if !root && p.collapsed[n.path] {
		p.rows = append(p.rows, devRow{
			line: devJSONLine(indent, lead, open+"..."+close, devPunctInk, tail),
			act:  devActJSON,
			key:  n.path,
		})

		return
	}

	row := devRow{line: devJSONLine(indent, lead, open, devPunctInk, "")}
	if !root {
		row.act, row.key = devActJSON, n.path
	}

	p.rows = append(p.rows, row)

	for i, kid := range n.kids {
		kidLead, kidTail := devLine{}, ","

		if i == len(n.kids)-1 {
			kidTail = ""
		}

		if n.kind == 'o' {
			kidLead = devJSONLead(kid.key)
		}

		p.emit(kid, depth+1, kidLead, kidTail, false)
	}

	p.rows = append(p.rows, devRow{line: devJSONPlain(indent, close, tail)})
}

// devJSONLeaf renders one scalar value in its JSON form.
func devJSONLeaf(v reflect.Value, key, path string) *devJSONNode {
	switch v.Kind() {
	case reflect.Bool:
		return devJSONScalar(strconv.FormatBool(v.Bool()), devBoolInk, key, path)
	case reflect.String:
		return devJSONScalar(string(devJSONMarshal(v.String())), devStrInk, key, path)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return devJSONScalar(strconv.FormatInt(v.Int(), 10), devNumInk, key, path)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return devJSONScalar(strconv.FormatUint(v.Uint(), 10), devNumInk, key, path)
	case reflect.Float32, reflect.Float64:
		return devJSONScalar(string(devJSONMarshal(v.Interface())), devNumInk, key, path)
	}

	return devJSONScalar("null", devBoolInk, key, path)
}

// devJSONExtra is the indent + text spans devJSONLine always adds.
const devJSONExtra = 2

// devJSONLine builds indent + lead + text + tail.
func devJSONLine(indent string, lead devLine, text string, ink color.RGBA, tail string) devLine {
	spans := make([]devSpan, 0, len(lead.spans)+devJSONExtra)
	spans = append(spans, devSpan{text: indent, ink: devPunctInk})
	spans = append(spans, lead.spans...)
	spans = append(spans, devSpan{text: text, ink: ink})

	if tail != "" {
		spans = append(spans, devSpan{text: tail, ink: devPunctInk})
	}

	return devLine{spans: spans}
}

// devJSONPlain builds indent + text + tail in punctuation ink.
func devJSONPlain(indent, text, tail string) devLine {
	return devJSONLine(indent, devLine{}, text, devPunctInk, tail)
}

// devJSONLead is the `"key": ` member prefix, key in the key ink.
func devJSONLead(key string) devLine {
	q := string(devJSONMarshal(key))

	return devLine{spans: []devSpan{
		{text: `"`, ink: devPunctInk},
		{text: q[1 : len(q)-1], ink: devKeyInk},
		{text: `": `, ink: devPunctInk},
	}}
}

// devJSONPair is one map entry waiting for its sorted place.
type devJSONPair struct {
	label string
	value reflect.Value
}

// buildMap renders entries sorted by their member name.
func (p *devJSONPrinter) buildMap(v reflect.Value, key, path string) *devJSONNode {
	node := &devJSONNode{kind: 'o', key: key, path: path}
	pairs := make([]devJSONPair, 0, v.Len())

	for _, k := range v.MapKeys() {
		pairs = append(pairs, devJSONPair{label: devJSONMapKey(k), value: v.MapIndex(k)})
	}

	sort.Slice(pairs, func(i, j int) bool { return pairs[i].label < pairs[j].label })

	for _, pair := range pairs {
		node.kids = append(node.kids, p.build(pair.value, pair.label, devJSONJoin(path, pair.label)))
	}

	return node
}

// devJSONMapKey is the member name for one map key.
func devJSONMapKey(k reflect.Value) string {
	if k.Kind() == reflect.String {
		return k.String()
	}

	return fmt.Sprint(k.Interface())
}

// devJSONNode is one value in the rendered document: an object, an array,
// or a scalar, plus the key and dotted path that name it.
type devJSONNode struct {
	kind   byte
	key    string
	path   string
	scalar string
	ink    color.RGBA
	kids   []*devJSONNode
}

// devJSONScalar builds a leaf node.
func devJSONScalar(text string, ink color.RGBA, key, path string) *devJSONNode {
	return &devJSONNode{kind: 's', key: key, path: path, scalar: text, ink: ink}
}

// devJSONNumber reads a json.Number through its string kind.
func devJSONNumber(v reflect.Value) (string, bool) {
	if v.Kind() != reflect.String {
		return "", false
	}

	n, ok := v.Interface().(json.Number)

	return n.String(), ok
}

// build turns one reflected value into a node. key is the member name and
// path the dotted location in the document.
func (p *devJSONPrinter) build(v reflect.Value, key, path string) *devJSONNode {
	if !v.IsValid() {
		return devJSONScalar("null", devBoolInk, key, path)
	}

	if n, ok := devJSONNumber(v); ok {
		return devJSONScalar(n, devNumInk, key, path)
	}

	switch v.Kind() {
	case reflect.Interface, reflect.Pointer:
		if v.IsNil() {
			return devJSONScalar("null", devBoolInk, key, path)
		}

		return p.build(v.Elem(), key, path)
	case reflect.Struct:
		return p.buildStruct(v, key, path)
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Uint8 {
			return devJSONScalar(string(devJSONMarshal(v.Interface())), devStrInk, key, path)
		}

		return p.buildArray(v, key, path)
	case reflect.Map:
		return p.buildMap(v, key, path)
	}

	return devJSONLeaf(v, key, path)
}

// devJSONMarshal is json.Marshal with a null fallback for values it refuses.
func devJSONMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("null")
	}

	return b
}

// devJSONTag splits a json struct tag into its name and options.
func devJSONTag(tag string) (string, string) {
	if i := strings.IndexByte(tag, ','); i >= 0 {
		return tag[:i], tag[i+1:]
	}

	return tag, ""
}

// devJSONEmpty reports the values omitempty drops.
func devJSONEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	}

	return false
}

// devJSONJoin names a child node under its parent.
func devJSONJoin(path, key string) string {
	if path == "" {
		return key
	}

	return path + "." + key
}
