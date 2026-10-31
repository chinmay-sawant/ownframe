package atspi

import (
	"encoding/binary"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func testServerView(t *testing.T) (*serverView, *int32) {
	t.Helper()
	app := Map(sampleTree())
	normalizeRootName(&app)
	id := int32(0)
	view := newServerView(app, DefaultName, dbusRef{Name: ":1.1", Path: accessibleRootPath}, 100, 200, 800, 600, &id)
	return view, &id
}

func objectPathFor(t *testing.T, view *serverView, id host.AccessibilityNodeID) string {
	t.Helper()
	for _, obj := range view.app.Objects {
		if obj.ID == id {
			return obj.Path
		}
	}
	t.Fatalf("missing object %s", id)
	return ""
}

func callView(view *serverView, path, iface, member string, body []byte) callResult {
	return callViewFrom(view, "", path, iface, member, body)
}

func callViewFrom(view *serverView, sender, path, iface, member string, body []byte) callResult {
	return view.handle(&dbusMsg{
		msgType: dbusTypeMethodCall,
		path:    path,
		iface:   iface,
		member:  member,
		sig:     "",
		sender:  sender,
		body:    body,
	})
}

func TestServeAccessibleBasics(t *testing.T) {
	view, _ := testServerView(t)
	namePath := objectPathFor(t, view, "id:name")
	root := accessibleRootPath
	docPath := objectPathFor(t, view, "document")

	if got := decodeStringResult(t, callView(view, namePath, atspiAccessible, "GetName", nil)); got != "Name" {
		t.Fatalf("name %q", got)
	}
	if got := decodeUint32Result(t, callView(view, namePath, atspiAccessible, "GetRole", nil)); got != 79 {
		t.Fatalf("role %d", got)
	}
	if got := len(decodeArraySO(t, callView(view, root, atspiAccessible, "GetChildren", nil).body)); got != 1 {
		t.Fatalf("app root children %d", got)
	}
	if got := len(decodeArraySO(t, callView(view, docPath, atspiAccessible, "GetChildren", nil).body)); got != 4 {
		t.Fatalf("document children %d", got)
	}
	if got := decodeInt32Result(t, callView(view, namePath, atspiAccessible, "GetIndexInParent", nil)); got != 0 {
		t.Fatalf("index %d", got)
	}
	parentName, parentPath := decodeStructSOResult(t, callView(view, namePath, atspiAccessible, "GetParent", nil))
	if parentName != DefaultName || parentPath != docPath {
		t.Fatalf("parent %q %q", parentName, parentPath)
	}

	states := decodeArrayUint32(t, callView(view, namePath, atspiAccessible, "GetState", nil).body)
	want := []int{stateEditable, stateEnabled, stateFocusable, stateFocused, stateSensitive, stateShowing, stateVisible}
	for _, bit := range want {
		if states[bit/stateWordBits]&(1<<uint(bit%stateWordBits)) == 0 {
			t.Fatalf("state bit %d missing from %v", bit, states)
		}
	}
	interfaces := decodeArrayString(t, callView(view, namePath, atspiAccessible, "GetInterfaces", nil).body)
	if !containsString(interfaces, atspiText) || !containsString(interfaces, atspiAction) {
		t.Fatalf("interfaces %v", interfaces)
	}
}

func TestServeProperties(t *testing.T) {
	view, appID := testServerView(t)
	namePath := objectPathFor(t, view, "id:name")

	var get dbusEncoder
	get.stringVal(atspiAccessible)
	get.stringVal("Name")
	res := callView(view, namePath, dbusProperties, "Get", get.bytes())
	sig, got := decodeVariantString(t, res.body)
	if sig != "s" || got != "Name" {
		t.Fatalf("name property %q %q", sig, got)
	}

	var set dbusEncoder
	set.stringVal(atspiApplication)
	set.stringVal("Id")
	set.variant("i", func(e *dbusEncoder) { e.int32Val(7) })
	if res := callViewFrom(view, ":1.1", accessibleRootPath, dbusProperties, "Set", set.bytes()); res.errName != "" {
		t.Fatalf("set id: %s", res.errText)
	}
	if *appID != 7 {
		t.Fatalf("app id %d", *appID)
	}
	var getId dbusEncoder
	getId.stringVal(atspiApplication)
	getId.stringVal("Id")
	res = callView(view, accessibleRootPath, dbusProperties, "Get", getId.bytes())
	sig, val, err := decodeVariant(t, res.body)
	if err != nil || sig != "i" || val != 7 {
		t.Fatalf("id property %q %d %v", sig, val, err)
	}
}

func TestServeText(t *testing.T) {
	view, _ := testServerView(t)
	namePath := objectPathFor(t, view, "id:name")

	var body dbusEncoder
	body.int32Val(0)
	body.int32Val(-1)
	if got := decodeStringResult(t, callView(view, namePath, atspiText, "GetText", body.bytes())); got != "Ada" {
		t.Fatalf("text %q", got)
	}

	var char dbusEncoder
	char.int32Val(1)
	if got := decodeInt32Result(t, callView(view, namePath, atspiText, "GetCharacterAtOffset", char.bytes())); got != 'd' {
		t.Fatalf("char %d", got)
	}

	var prop dbusEncoder
	prop.stringVal(atspiText)
	prop.stringVal("CharacterCount")
	res := callView(view, namePath, dbusProperties, "Get", prop.bytes())
	if sig, got, err := decodeVariant(t, res.body); err != nil || sig != "i" || got != 3 {
		t.Fatalf("character count %q %d %v", sig, got, err)
	}
}

func TestServeComponent(t *testing.T) {
	tree := sampleTree()
	tree.Nodes[0].Bounds = host.AccessibilityBounds{X: 0, Y: 0, W: 200, H: 100}
	tree.Nodes[0].HasBounds = true
	tree.Nodes[1].Bounds = host.AccessibilityBounds{X: 10, Y: 20, W: 50, H: 30}
	tree.Nodes[1].HasBounds = true
	app := Map(tree)
	id := int32(0)
	view := newServerView(app, DefaultName, dbusRef{}, 100, 200, 800, 600, &id)
	namePath := objectPathFor(t, view, "id:name")

	var body dbusEncoder
	body.uint32Val(0)
	x, y, w, h := decodeRectResult(t, callView(view, namePath, atspiComponent, "GetExtents", body.bytes()))
	if x != 110 || y != 220 || w != 50 || h != 30 {
		t.Fatalf("screen rect %d %d %d %d", x, y, w, h)
	}

	var point dbusEncoder
	point.int32Val(120)
	point.int32Val(230)
	point.uint32Val(0)
	hitName, hitPath := decodeStructSOResult(t, callView(view, accessibleRootPath, atspiComponent, "GetAccessibleAtPoint", point.bytes()))
	if hitName != DefaultName || hitPath != namePath {
		t.Fatalf("hit %q %q", hitName, hitPath)
	}

	var parentCoord dbusEncoder
	parentCoord.uint32Val(2)
	x, y, _, _ = decodeRectResult(t, callView(view, namePath, atspiComponent, "GetExtents", parentCoord.bytes()))
	if x != 10 || y != 20 {
		t.Fatalf("parent rect %d %d", x, y)
	}
}

func TestEncodeEventSignatures(t *testing.T) {
	sig, body, member, err := encodeEvent(Event{Path: "/org/a11y/atspi/accessible/n1", Kind: "state-changed", Detail: "focused", Enabled: true}, DefaultName, accessibleRootPath)
	if err != nil {
		t.Fatal(err)
	}
	if sig != "siiva{sv}" || member != "StateChanged" {
		t.Fatalf("sig %q member %q", sig, member)
	}
	d := dbusDecoder{body: body}
	detail, _ := d.stringVal()
	enabled, _ := d.int32Val()
	if extra, err := d.int32Val(); err != nil || extra != 0 {
		t.Fatalf("extra int %d %v", extra, err)
	}
	variantSig, variant, err := d.variant()
	if err != nil || detail != "focused" || enabled != 1 || variantSig != "i" {
		t.Fatalf("event %q %d %q %v", detail, enabled, variantSig, err)
	}
	if n, _ := variant.int32Val(); n != 0 {
		t.Fatalf("variant %d", n)
	}
	if dictLen := binary.LittleEndian.Uint32(body[d.off:]); dictLen != 0 {
		t.Fatalf("dict len %d", dictLen)
	}

	sig, body, member, err = encodeEvent(Event{Path: "/org/a11y/atspi/accessible/n1", Kind: "children-changed", Detail: "add", Child: "/org/a11y/atspi/accessible/n2", Index: 2}, DefaultName, accessibleRootPath)
	if err != nil {
		t.Fatal(err)
	}
	if sig != "siiva{sv}" || member != "ChildrenChanged" {
		t.Fatalf("sig %q member %q", sig, member)
	}
	d = dbusDecoder{body: body}
	detail, _ = d.stringVal()
	index, _ := d.int32Val()
	if extra, err := d.int32Val(); err != nil || extra != 0 {
		t.Fatalf("extra int %d %v", extra, err)
	}
	variantSig, variant, err = d.variant()
	if err != nil || detail != "add" || index != 2 || variantSig != "(so)" {
		t.Fatalf("event %q %d %q %v", detail, index, variantSig, err)
	}
	childName, _ := variant.stringVal()
	childPath, _ := variant.stringVal()
	if childName != DefaultName || childPath != "/org/a11y/atspi/accessible/n2" {
		t.Fatalf("child %q %q", childName, childPath)
	}
}

func TestServeTextOffsetsAlign(t *testing.T) {
	tree := sampleTree()
	tree.Nodes[1].Value = "one two three"
	app := Map(tree)
	id := int32(0)
	view := newServerView(app, DefaultName, dbusRef{}, 0, 0, 800, 600, &id)
	namePath := objectPathFor(t, view, "id:name")

	decodeSII := func(res callResult) (string, int32, int32) {
		t.Helper()
		if res.errName != "" {
			t.Fatalf("%s: %s", res.errName, res.errText)
		}
		d := dbusDecoder{body: res.body}
		s, err := d.stringVal()
		if err != nil {
			t.Fatal(err)
		}
		start, _ := d.int32Val()
		end, _ := d.int32Val()
		if d.off != len(res.body) {
			t.Fatalf("sii body has %d trailing bytes", len(res.body)-d.off)
		}
		return s, start, end
	}

	var at dbusEncoder
	at.int32Val(5)
	at.uint32Val(1) // WORD_START
	s, start, end := decodeSII(callView(view, namePath, atspiText, "GetTextAtOffset", at.bytes()))
	if s != "two" || start != 4 || end != 7 {
		t.Fatalf("at offset = %q %d %d", s, start, end)
	}

	var after dbusEncoder
	after.int32Val(0)
	after.uint32Val(1)
	s, start, end = decodeSII(callView(view, namePath, atspiText, "GetTextAfterOffset", after.bytes()))
	if s != "two" || start != 4 || end != 7 {
		t.Fatalf("after = %q %d %d", s, start, end)
	}

	var before dbusEncoder
	before.int32Val(9)
	before.uint32Val(1)
	s, start, end = decodeSII(callView(view, namePath, atspiText, "GetTextBeforeOffset", before.bytes()))
	if s != "two" || start != 4 || end != 7 {
		t.Fatalf("before = %q %d %d", s, start, end)
	}
}

func TestServePointOutsideAndLayer(t *testing.T) {
	view, _ := testServerView(t)
	var point dbusEncoder
	point.int32Val(4000)
	point.int32Val(4000)
	point.uint32Val(0)
	name, path := decodeStructSOResult(t, callView(view, accessibleRootPath, atspiComponent, "GetAccessibleAtPoint", point.bytes()))
	if name != "" || path != atspiNullPath {
		t.Fatalf("outside hit = %q %q", name, path)
	}
	if got := decodeUint32Result(t, callView(view, accessibleRootPath, atspiComponent, "GetLayer", nil)); got != 7 {
		t.Fatalf("root layer = %d", got)
	}
}

func TestServeRejectsUnknownPropertySet(t *testing.T) {
	view, _ := testServerView(t)
	var body dbusEncoder
	body.stringVal(atspiAccessible)
	body.stringVal("NoSuchProperty")
	body.variant("s", func(e *dbusEncoder) { e.stringVal("x") })
	res := callView(view, accessibleRootPath, dbusProperties, "Set", body.bytes())
	if res.errName != "org.freedesktop.DBus.Error.UnknownProperty" {
		t.Fatalf("error = %q %q", res.errName, res.errText)
	}
}

func TestServeRootIndexAndParent(t *testing.T) {
	app := Map(sampleTree())
	id := int32(0)
	view := newServerView(app, DefaultName, dbusRef{}, 0, 0, 800, 600, &id)
	if got := decodeInt32Result(t, callView(view, accessibleRootPath, atspiAccessible, "GetIndexInParent", nil)); got != -1 {
		t.Fatalf("root index = %d", got)
	}
	parentName, parentPath := decodeStructSOResult(t, callView(view, accessibleRootPath, atspiAccessible, "GetParent", nil))
	if parentName != "" || parentPath != atspiNullPath {
		t.Fatalf("unseeded parent = %q %q", parentName, parentPath)
	}
}

func TestServeChildIndex(t *testing.T) {
	view, _ := testServerView(t)
	docPath := objectPathFor(t, view, "document")
	pwPath := objectPathFor(t, view, "id:pw")

	if got := decodeInt32Result(t, callView(view, docPath, atspiAccessible, "GetIndexInParent", nil)); got != 0 {
		t.Fatalf("document index = %d", got)
	}
	if got := decodeInt32Result(t, callView(view, pwPath, atspiAccessible, "GetIndexInParent", nil)); got != 2 {
		t.Fatalf("password index = %d", got)
	}

	orphan := Object{Path: objectPath("orphan"), Parent: docPath, Role: RoleStatic}
	view.objects[orphan.Path] = orphan
	if got := decodeInt32Result(t, callView(view, orphan.Path, atspiAccessible, "GetIndexInParent", nil)); got != -1 {
		t.Fatalf("orphan index = %d", got)
	}
}

func TestTextRangeClamps(t *testing.T) {
	const value = "héllo 世界"
	cases := []struct {
		start, end int32
		want       string
	}{
		{0, -1, value},
		{1, 3, "él"},
		{2, 2, ""},
		{-5, 2, "hé"},
		{5, 99, " 世界"},
		{99, 99, ""},
		{4, 1, ""},
	}
	for _, c := range cases {
		if got := textRange(value, c.start, c.end); got != c.want {
			t.Fatalf("textRange(%q, %d, %d) = %q, want %q", value, c.start, c.end, got, c.want)
		}
	}
}

func decodeStringResult(t *testing.T, res callResult) string {
	t.Helper()
	if res.errName != "" {
		t.Fatalf("%s: %s", res.errName, res.errText)
	}
	got, err := decodeStringBody(res.body)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func decodeUint32Result(t *testing.T, res callResult) uint32 {
	t.Helper()
	if res.errName != "" {
		t.Fatalf("%s: %s", res.errName, res.errText)
	}
	got, err := decodeUint32Body(res.body)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func decodeInt32Result(t *testing.T, res callResult) int32 {
	t.Helper()
	got := decodeUint32Result(t, res)
	return int32(got) //nolint:gosec // reason: the bits are reinterpreted as a signed D-Bus int32.
}

func decodeStructSOResult(t *testing.T, res callResult) (string, string) {
	t.Helper()
	if res.errName != "" {
		t.Fatalf("%s: %s", res.errName, res.errText)
	}
	name, path, err := decodeSO(res.body)
	if err != nil {
		t.Fatal(err)
	}
	return name, path
}

func decodeArrayUint32(t *testing.T, body []byte) []uint32 {
	t.Helper()
	if len(body) < 4 {
		t.Fatal("short au")
	}
	payload := int(binary.LittleEndian.Uint32(body[0:4]))
	if payload%4 != 0 || len(body) < 4+payload {
		t.Fatalf("bad au payload %d in %d", payload, len(body))
	}
	n := payload / 4
	out := make([]uint32, n)
	for i := range out {
		out[i] = binary.LittleEndian.Uint32(body[4+i*4 : 8+i*4])
	}
	return out
}

func decodeArraySO(t *testing.T, body []byte) [][2]string {
	t.Helper()
	if len(body) < 4 {
		t.Fatal("short a(so)")
	}
	payload := int(binary.LittleEndian.Uint32(body[0:4]))
	d := dbusDecoder{body: body, off: 4}
	if err := d.align(dbusAlign8); err != nil {
		t.Fatal(err)
	}
	end := d.off + payload
	if end > len(body) {
		t.Fatalf("bad a(so) payload %d in %d", payload, len(body))
	}
	var out [][2]string
	for d.off < end {
		if err := d.align(dbusAlign8); err != nil {
			t.Fatal(err)
		}
		name, err := d.stringVal()
		if err != nil {
			t.Fatal(err)
		}
		path, err := d.stringVal()
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, [2]string{name, path})
	}
	return out
}

func decodeArrayString(t *testing.T, body []byte) []string {
	t.Helper()
	if len(body) < 4 {
		t.Fatal("short as")
	}
	payload := int(binary.LittleEndian.Uint32(body[0:4]))
	d := dbusDecoder{body: body, off: 4}
	end := d.off + payload
	if end > len(body) {
		t.Fatalf("bad as payload %d in %d", payload, len(body))
	}
	var out []string
	for d.off < end {
		s, err := d.stringVal()
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, s)
	}
	return out
}

func decodeVariantString(t *testing.T, body []byte) (string, string) {
	t.Helper()
	d := dbusDecoder{body: body}
	sig, val, err := d.variant()
	if err != nil {
		t.Fatal(err)
	}
	s, err := val.stringVal()
	if err != nil {
		t.Fatal(err)
	}
	return sig, s
}

func decodeVariant(t *testing.T, body []byte) (string, int32, error) {
	t.Helper()
	d := dbusDecoder{body: body}
	sig, val, err := d.variant()
	if err != nil {
		return "", 0, err
	}
	if sig == "i" {
		n, err := val.int32Val()
		return sig, n, err
	}
	return sig, 0, nil
}

func decodeRectResult(t *testing.T, res callResult) (int32, int32, int32, int32) {
	t.Helper()
	if res.errName != "" {
		t.Fatalf("%s: %s", res.errName, res.errText)
	}
	d := dbusDecoder{body: res.body}
	if err := d.align(dbusAlign8); err != nil {
		t.Fatal(err)
	}
	x, _ := d.int32Val()
	y, _ := d.int32Val()
	w, _ := d.int32Val()
	h, _ := d.int32Val()
	return x, y, w, h
}

func containsString(vals []string, want string) bool {
	for _, v := range vals {
		if v == want {
			return true
		}
	}
	return false
}
