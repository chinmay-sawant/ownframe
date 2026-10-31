package atspi

import (
	"math"
	"os"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	atspiAccessible  = "org.a11y.atspi.Accessible"
	atspiApplication = "org.a11y.atspi.Application"
	atspiComponent   = "org.a11y.atspi.Component"
	atspiText        = "org.a11y.atspi.Text"
	atspiValue       = "org.a11y.atspi.Value"
	atspiAction      = "org.a11y.atspi.Action"
	atspiCache       = "org.a11y.atspi.Cache"
	atspiCachePath   = "/org/a11y/atspi/cache"
	dbusPeer         = "org.freedesktop.DBus.Peer"
	dbusIntrospect   = "org.freedesktop.DBus.Introspectable"
	dbusProperties   = "org.freedesktop.DBus.Properties"
	atspiNullPath    = "/org/a11y/atspi/null"

	atspiCoordParent     = 2
	componentLayerWidget = 3
	componentLayerWindow = 7

	textBoundaryWordStart     = 1
	textBoundaryWordEnd       = 2
	textBoundarySentenceStart = 3
	textBoundarySentenceEnd   = 4
	textBoundaryLineStart     = 5
	textBoundaryLineEnd       = 6

	textGranularityChar      = 0
	textGranularityWord      = 1
	textGranularitySentence  = 2
	textGranularityLine      = 3
	textGranularityParagraph = 4
)

type dbusRef struct {
	Name string
	Path string
}

type callResult struct {
	sig     string
	body    []byte
	errName string
	errText string
}

func okResult(sig string, body []byte) callResult {
	return callResult{sig: sig, body: body}
}

func emptyResult() callResult {
	return callResult{}
}

func errResult(name, text string) callResult {
	return callResult{errName: name, errText: text}
}

type serverView struct {
	app     Application
	objects map[string]Object
	busName string
	parent  dbusRef
	originX int
	originY int
	width   int
	height  int
	appID   *int32
}

func newServerView(app Application, busName string, parent dbusRef, originX, originY, width, height int, appID *int32) *serverView {
	objects := make(map[string]Object, len(app.Objects))
	for _, obj := range app.Objects {
		objects[obj.Path] = obj
	}
	return &serverView{
		app:     app,
		objects: objects,
		busName: busName,
		parent:  parent,
		originX: originX,
		originY: originY,
		width:   width,
		height:  height,
		appID:   appID,
	}
}

func (v *serverView) object(path string) (Object, bool) {
	if path == atspiNullPath {
		return Object{}, false
	}
	obj, ok := v.objects[path]
	return obj, ok
}

func (v *serverView) ref(path string) [2]string {
	return [2]string{v.busName, path}
}

func (v *serverView) rootRef() [2]string {
	return v.ref(v.app.RootPath)
}

func (v *serverView) parentRef(obj Object) [2]string {
	if obj.Path == v.app.RootPath {
		if v.parent.Name == "" || v.parent.Path == "" {
			return [2]string{"", atspiNullPath}
		}
		return [2]string{v.parent.Name, v.parent.Path}
	}
	if obj.Parent != "" {
		return v.ref(obj.Parent)
	}
	return v.rootRef()
}

func (v *serverView) indexInParent(obj Object) int32 {
	if obj.Path == v.app.RootPath {
		// The application root is a child of the desktop, not of a
		// published child list. libatspi writes a non-negative index into
		// the parent's child array, so 0 would overwrite another app.
		return -1
	}
	parent, ok := v.objects[obj.Parent]
	if !ok {
		return -1
	}
	for i, child := range parent.Children {
		if child == obj.Path {
			return int32(i)
		}
	}
	return -1
}

func (v *serverView) interfaces(obj Object) []string {
	out := []string{atspiAccessible, atspiComponent}
	if obj.Path == v.app.RootPath {
		out = append(out, atspiApplication)
	}
	if textCapable(obj) {
		out = append(out, atspiText)
	}
	if obj.Role == RoleComboBox {
		out = append(out, atspiValue)
	}
	if len(obj.Actions) > 0 {
		out = append(out, atspiAction)
	}
	return out
}

func textCapable(obj Object) bool {
	switch obj.Role {
	case RoleDocumentWeb, RoleParagraph, RoleHeading, RoleStatic, RoleLink,
		RoleEntry, RolePasswordText, RoleListItem, RoleComboBox:
		return true
	}
	return obj.Value != ""
}

func (v *serverView) handle(msg *dbusMsg) callResult {
	if msg.path == atspiCachePath {
		switch msg.iface {
		case atspiCache:
			return v.handleCache(msg)
		case dbusProperties:
			return v.handleCacheProperties(msg)
		case dbusPeer:
			return v.handlePeer(msg)
		case dbusIntrospect:
			return v.handleIntrospect(msg)
		}
		return errResult("org.freedesktop.DBus.Error.UnknownInterface", "unknown interface "+msg.iface)
	}
	switch msg.iface {
	case dbusPeer:
		return v.handlePeer(msg)
	case dbusIntrospect:
		return v.handleIntrospect(msg)
	case dbusProperties:
		return v.handleProperties(msg)
	case atspiAccessible:
		return v.handleAccessible(msg)
	case atspiApplication:
		return v.handleApplication(msg)
	case atspiComponent:
		return v.handleComponent(msg)
	case atspiText:
		return v.handleText(msg)
	case atspiAction:
		return v.handleAction(msg)
	}
	return errResult("org.freedesktop.DBus.Error.UnknownInterface", "unknown interface "+msg.iface)
}

func (v *serverView) handlePeer(msg *dbusMsg) callResult {
	switch msg.member {
	case "Ping":
		return emptyResult()
	case "GetMachineId":
		id, _ := os.ReadFile("/etc/machine-id")
		return okResult("s", encodeStringValue(strings.TrimSpace(string(id))))
	}
	return errResult("org.freedesktop.DBus.Error.UnknownMethod", "unknown method "+msg.member)
}

func (v *serverView) handleCache(msg *dbusMsg) callResult {
	if msg.member != "GetItems" {
		return errResult("org.freedesktop.DBus.Error.UnknownMethod", "unknown method "+msg.member)
	}
	var e dbusEncoder
	e.array("((so)(so)(so)iiassusau)", func(e *dbusEncoder) {
		for _, obj := range v.app.Objects {
			e.structValue(func(e *dbusEncoder) {
				writeRef(e, v.ref(obj.Path))
				writeRef(e, v.rootRef())
				writeRef(e, v.parentRef(obj))
				e.int32Val(v.indexInParent(obj))
				e.int32Val(int32(len(obj.Children))) //nolint:gosec // reason: child counts are far below MaxInt32.
				interfaces := v.interfaces(obj)
				e.array("s", func(e *dbusEncoder) {
					for _, iface := range interfaces {
						e.stringVal(iface)
					}
				})
				e.stringVal(obj.Name)
				e.uint32Val(RoleNumber(obj.Role))
				e.stringVal(obj.Description)
				states := stateWords(obj.State)
				e.array("u", func(e *dbusEncoder) {
					for _, word := range states {
						e.uint32Val(word)
					}
				})
			})
		}
	})
	return okResult("a((so)(so)(so)iiassusau)", e.bytes())
}

func (v *serverView) handleCacheProperties(msg *dbusMsg) callResult {
	if msg.member != "Get" {
		return errResult("org.freedesktop.DBus.Error.UnknownMethod", "unknown method "+msg.member)
	}
	d := dbusDecoder{body: msg.body}
	iface, err := d.stringVal()
	if err != nil {
		return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
	}
	name, err := d.stringVal()
	if err != nil {
		return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
	}
	if iface == atspiCache && name == "version" {
		return okResult("v", encodePropertyValue("u", func(e *dbusEncoder) { e.uint32Val(1) }))
	}
	return errResult("org.freedesktop.DBus.Error.UnknownProperty", "unknown property "+iface+"."+name)
}

func writeRef(e *dbusEncoder, ref [2]string) {
	e.structValue(func(e *dbusEncoder) {
		e.stringVal(ref[0])
		e.stringVal(ref[1])
	})
}

func (v *serverView) handleIntrospect(msg *dbusMsg) callResult {
	if msg.path == atspiCachePath {
		return okResult("s", encodeStringValue(cacheIntrospectXML))
	}
	return okResult("s", encodeStringValue(introspectXML(v.objectAt(msg.path), v.app.RootPath)))
}

func (v *serverView) objectAt(path string) Object {
	obj, _ := v.object(path)
	return obj
}

func (v *serverView) handleProperties(msg *dbusMsg) callResult {
	obj, ok := v.object(msg.path)
	if !ok {
		return errResult("org.freedesktop.DBus.Error.UnknownObject", "unknown object "+msg.path)
	}
	switch msg.member {
	case "Get":
		d := dbusDecoder{body: msg.body}
		iface, err := d.stringVal()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		name, err := d.stringVal()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		prop, ok := v.property(obj, iface, name)
		if !ok {
			return errResult("org.freedesktop.DBus.Error.UnknownProperty", "unknown property "+iface+"."+name)
		}
		return okResult("v", encodePropertyValue(prop.sig, prop.write))
	case "GetAll":
		d := dbusDecoder{body: msg.body}
		iface, err := d.stringVal()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		props := v.properties(obj, iface)
		names := make([]string, 0, len(props))
		for name := range props {
			names = append(names, name)
		}
		sort.Strings(names)
		var e dbusEncoder
		e.array("{sv}", func(e *dbusEncoder) {
			for _, name := range names {
				prop := props[name]
				e.align(dbusAlign8)
				e.stringVal(name)
				e.variant(prop.sig, prop.write)
			}
		})
		return okResult("a{sv}", e.bytes())
	case "Set":
		d := dbusDecoder{body: msg.body}
		iface, err := d.stringVal()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		name, err := d.stringVal()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		sig, val, err := d.variant()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		if _, ok := v.property(obj, iface, name); !ok {
			return errResult("org.freedesktop.DBus.Error.UnknownProperty", "unknown property "+iface+"."+name)
		}
		if iface == atspiApplication && name == "Id" && obj.Path == v.app.RootPath {
			if sig != "i" {
				return errResult("org.freedesktop.DBus.Error.InvalidArgs", "Application.Id needs an int32")
			}
			id, err := val.int32Val()
			if err != nil {
				return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
			}
			*v.appID = id
			return emptyResult()
		}
		return errResult("org.freedesktop.DBus.Error.PropertyReadOnly", "property "+iface+"."+name+" is read-only")
	}
	return errResult("org.freedesktop.DBus.Error.UnknownMethod", "unknown method "+msg.member)
}

type propertyValue struct {
	sig   string
	write func(*dbusEncoder)
}

func (v *serverView) property(obj Object, iface, name string) (propertyValue, bool) {
	prop, ok := v.properties(obj, iface)[name]
	return prop, ok
}

func (v *serverView) properties(obj Object, iface string) map[string]propertyValue {
	props := map[string]propertyValue{}
	switch iface {
	case atspiAccessible:
		props["version"] = propertyValue{"u", func(e *dbusEncoder) { e.uint32Val(1) }}
		props["Name"] = propertyValue{"s", func(e *dbusEncoder) { e.stringVal(obj.Name) }}
		props["Description"] = propertyValue{"s", func(e *dbusEncoder) { e.stringVal(obj.Description) }}
		props["Parent"] = propertyValue{"(so)", func(e *dbusEncoder) {
			parent := v.parentRef(obj)
			e.structValue(func(e *dbusEncoder) {
				e.stringVal(parent[0])
				e.stringVal(parent[1])
			})
		}}
		props["ChildCount"] = propertyValue{"i", func(e *dbusEncoder) { e.int32Val(int32(len(obj.Children))) }} //nolint:gosec // reason: child counts are far below MaxInt32.
		props["Locale"] = propertyValue{"s", func(e *dbusEncoder) { e.stringVal("") }}
		props["AccessibleId"] = propertyValue{"s", func(e *dbusEncoder) { e.stringVal(string(obj.ID)) }}
		props["HelpText"] = propertyValue{"s", func(e *dbusEncoder) { e.stringVal("") }}
	case atspiApplication:
		if obj.Path != v.app.RootPath {
			return nil
		}
		props["ToolkitName"] = propertyValue{"s", func(e *dbusEncoder) { e.stringVal("ownframe") }}
		props["Version"] = propertyValue{"s", func(e *dbusEncoder) { e.stringVal("") }}
		props["ToolkitVersion"] = propertyValue{"s", func(e *dbusEncoder) { e.stringVal("") }}
		props["AtspiVersion"] = propertyValue{"s", func(e *dbusEncoder) { e.stringVal("2.1") }}
		props["InterfaceVersion"] = propertyValue{"u", func(e *dbusEncoder) { e.uint32Val(1) }}
		props["Id"] = propertyValue{"i", func(e *dbusEncoder) { e.int32Val(*v.appID) }}
	case atspiComponent:
		props["version"] = propertyValue{"u", func(e *dbusEncoder) { e.uint32Val(1) }}
	case atspiText:
		if !textCapable(obj) {
			return nil
		}
		props["version"] = propertyValue{"u", func(e *dbusEncoder) { e.uint32Val(1) }}
		props["CharacterCount"] = propertyValue{"i", func(e *dbusEncoder) {
			e.int32Val(int32(utf8.RuneCountInString(obj.Value))) //nolint:gosec // reason: a page value is far below MaxInt32.
		}}
		props["CaretOffset"] = propertyValue{"i", func(e *dbusEncoder) { e.int32Val(0) }}
	case atspiValue:
		if obj.Role != RoleComboBox {
			return nil
		}
		props["version"] = propertyValue{"u", func(e *dbusEncoder) { e.uint32Val(1) }}
		props["MinimumValue"] = propertyValue{"d", func(e *dbusEncoder) { e.doubleVal(0) }}
		props["MaximumValue"] = propertyValue{"d", func(e *dbusEncoder) { e.doubleVal(0) }}
		props["MinimumIncrement"] = propertyValue{"d", func(e *dbusEncoder) { e.doubleVal(0) }}
		props["CurrentValue"] = propertyValue{"d", func(e *dbusEncoder) { e.doubleVal(0) }}
		props["Text"] = propertyValue{"s", func(e *dbusEncoder) { e.stringVal(obj.Value) }}
	case atspiAction:
		if len(obj.Actions) == 0 {
			return nil
		}
		props["version"] = propertyValue{"u", func(e *dbusEncoder) { e.uint32Val(1) }}
		props["NActions"] = propertyValue{"i", func(e *dbusEncoder) { e.int32Val(int32(len(obj.Actions))) }} //nolint:gosec // reason: a node has only a few actions.
	}
	return props
}

func (v *serverView) handleAccessible(msg *dbusMsg) callResult {
	obj, ok := v.object(msg.path)
	if !ok {
		return errResult("org.freedesktop.DBus.Error.UnknownObject", "unknown object "+msg.path)
	}
	switch msg.member {
	case "GetName":
		return okResult("s", encodeStringValue(obj.Name))
	case "GetDescription":
		return okResult("s", encodeStringValue(obj.Description))
	case "GetParent":
		parent := v.parentRef(obj)
		return okResult("(so)", encodeStructSO(parent[0], parent[1]))
	case "GetChildAtIndex":
		d := dbusDecoder{body: msg.body}
		index, err := d.int32Val()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		if index < 0 || int(index) >= len(obj.Children) {
			return okResult("(so)", encodeStructSO("", atspiNullPath))
		}
		ref := v.ref(obj.Children[index])
		return okResult("(so)", encodeStructSO(ref[0], ref[1]))
	case "GetChildren":
		refs := make([][2]string, 0, len(obj.Children))
		for _, path := range obj.Children {
			refs = append(refs, v.ref(path))
		}
		return okResult("a(so)", encodeArraySO(refs))
	case "GetIndexInParent":
		return okResult("i", encodeInt32Value(v.indexInParent(obj)))
	case "GetRelationSet":
		return okResult("a(ua(so))", encodeEmptyRelationSet())
	case "GetRole":
		return okResult("u", encodeUint32Value(RoleNumber(obj.Role)))
	case "GetRoleName":
		return okResult("s", encodeStringValue(string(obj.Role)))
	case "GetLocalizedRoleName":
		return okResult("s", encodeStringValue(string(obj.Role)))
	case "GetState":
		return okResult("au", encodeArrayUint32(stateWords(obj.State)))
	case "GetAttributes":
		return okResult("a{ss}", encodeStringMap(v.attributes(obj)))
	case "GetApplication":
		root := v.rootRef()
		return okResult("(so)", encodeStructSO(root[0], root[1]))
	case "GetInterfaces":
		return okResult("as", encodeArrayString(v.interfaces(obj)))
	}
	return errResult("org.freedesktop.DBus.Error.UnknownMethod", "unknown method "+msg.member)
}

func (v *serverView) attributes(obj Object) map[string]string {
	out := map[string]string{"toolkit": "ownframe"}
	if obj.ID != "" {
		out["id"] = string(obj.ID)
	}
	return out
}

func (v *serverView) handleApplication(msg *dbusMsg) callResult {
	obj, ok := v.object(msg.path)
	if !ok || obj.Path != v.app.RootPath {
		return errResult("org.freedesktop.DBus.Error.UnknownObject", "unknown application object "+msg.path)
	}
	switch msg.member {
	case "GetLocale":
		return okResult("s", encodeStringValue("C"))
	case "GetApplicationBusAddress":
		return okResult("s", encodeStringValue(""))
	}
	return errResult("org.freedesktop.DBus.Error.UnknownMethod", "unknown method "+msg.member)
}

func (v *serverView) handleComponent(msg *dbusMsg) callResult {
	obj, ok := v.object(msg.path)
	if !ok {
		return errResult("org.freedesktop.DBus.Error.UnknownObject", "unknown object "+msg.path)
	}
	switch msg.member {
	case "Contains":
		d := dbusDecoder{body: msg.body}
		x, err := d.int32Val()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		y, err := d.int32Val()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		coord, err := d.uint32Val()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		bx, by, bw, bh := v.rect(obj, coord)
		return okResult("b", encodeBoolValue(x >= bx && x < bx+bw && y >= by && y < by+bh))
	case "GetAccessibleAtPoint":
		d := dbusDecoder{body: msg.body}
		x, err := d.int32Val()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		y, err := d.int32Val()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		coord, err := d.uint32Val()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		hit := v.hitTest(obj, x, y, coord)
		if hit.Path == "" {
			return okResult("(so)", encodeStructSO("", atspiNullPath))
		}
		ref := v.ref(hit.Path)
		return okResult("(so)", encodeStructSO(ref[0], ref[1]))
	case "GetExtents":
		coord, err := v.coordArg(msg)
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		x, y, w, h := v.rect(obj, coord)
		return okResult("(iiii)", encodeRect(x, y, w, h))
	case "GetPosition":
		coord, err := v.coordArg(msg)
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		x, y, _, _ := v.rect(obj, coord)
		return okResult("ii", append(encodeInt32Value(x), encodeInt32Value(y)...))
	case "GetSize":
		_, _, w, h := v.rect(obj, 1)
		return okResult("ii", append(encodeInt32Value(w), encodeInt32Value(h)...))
	case "GetLayer":
		if obj.Path == v.app.RootPath {
			return okResult("u", encodeUint32Value(componentLayerWindow))
		}
		return okResult("u", encodeUint32Value(componentLayerWidget))
	case "GetMDIZOrder":
		var e dbusEncoder
		e.int16Val(-1)
		return okResult("n", e.bytes())
	case "GrabFocus":
		return okResult("b", encodeBoolValue(false))
	case "GetAlpha":
		return okResult("d", func() []byte {
			var e dbusEncoder
			e.doubleVal(1)
			return e.bytes()
		}())
	case "SetExtents", "SetPosition", "SetSize", "ScrollTo", "ScrollToPoint":
		return okResult("b", encodeBoolValue(false))
	}
	return errResult("org.freedesktop.DBus.Error.UnknownMethod", "unknown method "+msg.member)
}

func (v *serverView) coordArg(msg *dbusMsg) (uint32, error) {
	d := dbusDecoder{body: msg.body}
	return d.uint32Val()
}

func (v *serverView) rect(obj Object, coord uint32) (int32, int32, int32, int32) {
	if !obj.HasBounds {
		if obj.Path == v.app.RootPath && v.width > 0 && v.height > 0 {
			if coord == 0 {
				return clampInt32(v.originX), clampInt32(v.originY), clampInt32(v.width), clampInt32(v.height)
			}
			return 0, 0, clampInt32(v.width), clampInt32(v.height)
		}
		return 0, 0, 0, 0
	}
	x := obj.Bounds.X
	y := obj.Bounds.Y
	switch coord {
	case 0:
		x += float64(v.originX)
		y += float64(v.originY)
	case atspiCoordParent:
		if parent, ok := v.objects[obj.Parent]; ok && parent.HasBounds {
			x -= parent.Bounds.X
			y -= parent.Bounds.Y
		}
	}
	return roundInt32(x), roundInt32(y), roundInt32(obj.Bounds.W), roundInt32(obj.Bounds.H)
}

func roundInt32(v float64) int32 {
	r := math.Round(v)
	if r > math.MaxInt32 {
		return math.MaxInt32
	}
	if r < math.MinInt32 {
		return math.MinInt32
	}
	return int32(r)
}

func clampInt32(v int) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < math.MinInt32 {
		return math.MinInt32
	}
	return int32(v)
}

func (v *serverView) hitTest(obj Object, x, y int32, coord uint32) Object {
	bx, by, bw, bh := v.rect(obj, coord)
	if x < bx || x >= bx+bw || y < by || y >= by+bh {
		return Object{}
	}
	px, py := x, y
	if coord == atspiCoordParent {
		px -= bx
		py -= by
	}
	for _, childPath := range obj.Children {
		child, ok := v.objects[childPath]
		if !ok {
			continue
		}
		cbx, cby, cbw, cbh := v.rect(child, coord)
		if px >= cbx && px < cbx+cbw && py >= cby && py < cby+cbh {
			return v.hitTest(child, px, py, coord)
		}
	}
	return obj
}

func (v *serverView) handleAction(msg *dbusMsg) callResult {
	obj, ok := v.object(msg.path)
	if !ok {
		return errResult("org.freedesktop.DBus.Error.UnknownObject", "unknown object "+msg.path)
	}
	indexArg := func() (int32, callResult) {
		d := dbusDecoder{body: msg.body}
		index, err := d.int32Val()
		if err != nil {
			return 0, errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		if index < 0 || int(index) >= len(obj.Actions) {
			return 0, errResult("org.a11y.atspi.Error.InvalidIndex", "action index out of range")
		}
		return index, emptyResult()
	}
	switch msg.member {
	case "GetName", "GetDescription", "GetKeyBinding":
		index, res := indexArg()
		if res.errName != "" {
			return res
		}
		action := obj.Actions[index]
		name := atspiActionName(action)
		if msg.member == "GetDescription" {
			return okResult("s", encodeStringValue(name))
		}
		if msg.member == "GetKeyBinding" {
			return okResult("s", encodeStringValue(""))
		}
		return okResult("s", encodeStringValue(name))
	case "GetLocalizedName":
		index, res := indexArg()
		if res.errName != "" {
			return res
		}
		return okResult("s", encodeStringValue(atspiActionName(obj.Actions[index])))
	case "GetActions":
		actions := make([][3]string, 0, len(obj.Actions))
		for _, action := range obj.Actions {
			name := atspiActionName(action)
			actions = append(actions, [3]string{name, name, ""})
		}
		return okResult("a(sss)", encodeActionArray(actions))
	case "DoAction":
		return okResult("b", encodeBoolValue(false))
	}
	return errResult("org.freedesktop.DBus.Error.UnknownMethod", "unknown method "+msg.member)
}

func atspiActionName(action string) string {
	switch action {
	case ActionGrabFocus:
		return "focus"
	case ActionClick:
		return "click"
	case ActionSetText:
		return "set-text"
	}
	return action
}

func (v *serverView) handleText(msg *dbusMsg) callResult {
	obj, ok := v.object(msg.path)
	if !ok || !textCapable(obj) {
		return errResult("org.freedesktop.DBus.Error.UnknownObject", "object has no text interface")
	}
	text := []rune(obj.Value)
	argOffset := func() (int32, callResult) {
		d := dbusDecoder{body: msg.body}
		offset, err := d.int32Val()
		if err != nil {
			return 0, errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		return offset, emptyResult()
	}
	switch msg.member {
	case "GetText":
		d := dbusDecoder{body: msg.body}
		start, err := d.int32Val()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		end, err := d.int32Val()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		return okResult("s", encodeStringValue(sliceRunes(text, start, end)))
	case "GetStringAtOffset", "GetTextAtOffset", "GetTextBeforeOffset", "GetTextAfterOffset":
		offset, res := argOffset()
		if res.errName != "" {
			return res
		}
		typeArg := uint32(0)
		d := dbusDecoder{body: msg.body}
		_, _ = d.int32Val()
		if b, err := d.uint32Val(); err == nil {
			typeArg = b
		}
		granularity := typeArg
		if msg.member != "GetStringAtOffset" {
			granularity = textBoundaryGranularity(typeArg)
		}
		chunk, start, end := textChunk(text, offset, granularity)
		switch msg.member {
		case "GetTextBeforeOffset":
			pos := start - 1
			for pos >= 0 && unicode.IsSpace(text[pos]) {
				pos--
			}
			if pos >= 0 {
				chunk, start, end = textChunk(text, pos, granularity)
			}
		case "GetTextAfterOffset":
			pos := end
			n := int32(len(text)) //nolint:gosec // reason: page text is far below MaxInt32.
			for pos < n && unicode.IsSpace(text[pos]) {
				pos++
			}
			if pos < n {
				chunk, start, end = textChunk(text, pos, granularity)
			}
		}
		var e dbusEncoder
		e.stringVal(chunk)
		e.int32Val(start)
		e.int32Val(end)
		return okResult("sii", e.bytes())
	case "SetCaretOffset":
		return okResult("b", encodeBoolValue(false))
	case "GetCharacterAtOffset":
		offset, res := argOffset()
		if res.errName != "" {
			return res
		}
		r := int32(0)
		if offset >= 0 && int(offset) < len(text) {
			r = text[offset]
		}
		return okResult("i", encodeInt32Value(r))
	case "GetAttributeValue":
		return okResult("s", encodeStringValue(""))
	case "GetAttributes", "GetAttributeRun":
		d := dbusDecoder{body: msg.body}
		offset, err := d.int32Val()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		n := int32(len(text)) //nolint:gosec // reason: page text is far below MaxInt32.
		if offset < 0 {
			offset = 0
		}
		if offset > n {
			offset = n
		}
		end := offset + 1
		if end > n {
			end = n
		}
		var e dbusEncoder
		e.array("{ss}", func(*dbusEncoder) {})
		e.int32Val(offset)
		e.int32Val(end)
		return okResult("a{ss}ii", e.bytes())
	case "GetDefaultAttributes", "GetDefaultAttributeSet":
		return okResult("a{ss}", encodeStringMap(nil))
	case "GetCharacterExtents", "GetRangeExtents":
		d := dbusDecoder{body: msg.body}
		start, err := d.int32Val()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		if msg.member == "GetRangeExtents" {
			if _, err := d.int32Val(); err != nil {
				return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
			}
		}
		_ = start
		coord, err := d.uint32Val()
		if err != nil {
			return errResult("org.freedesktop.DBus.Error.InvalidArgs", err.Error())
		}
		x, y, w, h := v.rect(obj, coord)
		return okResult("iiii", encodeRect(x, y, w, h))
	case "GetOffsetAtPoint":
		return okResult("i", encodeInt32Value(0))
	case "GetNSelections":
		return okResult("i", encodeInt32Value(0))
	case "GetSelection":
		return okResult("ii", append(encodeInt32Value(0), encodeInt32Value(0)...))
	case "AddSelection", "RemoveSelection", "SetSelection", "ScrollSubstringTo", "ScrollSubstringToPoint":
		return okResult("b", encodeBoolValue(false))
	case "GetBoundedRanges":
		var e dbusEncoder
		e.array("(iisv)", func(*dbusEncoder) {})
		return okResult("a(iisv)", e.bytes())
	}
	return errResult("org.freedesktop.DBus.Error.UnknownMethod", "unknown method "+msg.member)
}

func sliceRunes(text []rune, start, end int32) string {
	n := int32(len(text)) //nolint:gosec // reason: page text is far below MaxInt32.
	if end < 0 || end > n {
		end = n
	}
	if start < 0 {
		start = 0
	}
	if start > end {
		start = end
	}
	return string(text[start:end])
}

func textBoundaryGranularity(boundary uint32) uint32 {
	switch boundary {
	case textBoundaryWordStart, textBoundaryWordEnd:
		return textGranularityWord
	case textBoundarySentenceStart, textBoundarySentenceEnd:
		return textGranularitySentence
	case textBoundaryLineStart, textBoundaryLineEnd:
		return textGranularityLine
	default:
		return textGranularityChar
	}
}

func textChunk(text []rune, offset int32, boundary uint32) (string, int32, int32) {
	n := int32(len(text)) //nolint:gosec // reason: page text is far below MaxInt32.
	if offset < 0 {
		offset = 0
	}
	if offset > n {
		offset = n
	}
	if n == 0 {
		return "", 0, 0
	}
	if offset == n {
		return "", n, n
	}
	switch boundary {
	case textGranularityWord:
		for offset < n && unicode.IsSpace(text[offset]) {
			offset++
		}
		if offset == n {
			return "", n, n
		}
		start := offset
		for start > 0 && !unicode.IsSpace(text[start-1]) {
			start--
		}
		end := offset
		for end < n && !unicode.IsSpace(text[end]) {
			end++
		}
		return string(text[start:end]), start, end
	case textGranularitySentence: // sentence
		start := offset
		for start > 0 && !strings.ContainsRune(".!?", text[start-1]) {
			start--
		}
		end := offset
		for end < n {
			if strings.ContainsRune(".!?", text[end]) {
				end++
				break
			}
			end++
		}
		return string(text[start:end]), start, end
	case textGranularityLine: // line
		start := offset
		for start > 0 && text[start-1] != '\n' {
			start--
		}
		end := offset
		for end < n && text[end] != '\n' {
			end++
		}
		return string(text[start:end]), start, end
	case textGranularityParagraph: // paragraph
		return string(text), 0, n
	default: // character
		end := offset + 1
		if end > n {
			end = n
		}
		return string(text[offset:end]), offset, end
	}
}

func introspectXML(obj Object, rootPath string) string {
	interfaces := []string{atspiAccessible, atspiComponent, atspiText, atspiValue, atspiAction}
	if obj.Path == rootPath {
		interfaces = append(interfaces, atspiApplication)
	}
	var b strings.Builder
	b.WriteString(`<node>`)
	for _, iface := range interfaces {
		switch iface {
		case atspiApplication:
			b.WriteString(`<interface name="org.a11y.atspi.Application">
<property name="ToolkitName" type="s" access="read"/>
<property name="Version" type="s" access="read"/>
<property name="ToolkitVersion" type="s" access="read"/>
<property name="AtspiVersion" type="s" access="read"/>
<property name="InterfaceVersion" type="u" access="read"/>
<property name="Id" type="i" access="readwrite"/>
<method name="GetLocale"><arg type="u" direction="in"/><arg type="s" direction="out"/></method>
<method name="GetApplicationBusAddress"><arg type="s" direction="out"/></method>
</interface>`)
		case atspiAccessible:
			b.WriteString(`<interface name="org.a11y.atspi.Accessible">
<method name="GetName"><arg type="s" direction="out"/></method>
<method name="GetDescription"><arg type="s" direction="out"/></method>
<method name="GetParent"><arg type="(so)" direction="out"/></method>
<method name="GetChildAtIndex"><arg type="i" direction="in"/><arg type="(so)" direction="out"/></method>
<method name="GetChildren"><arg type="a(so)" direction="out"/></method>
<method name="GetIndexInParent"><arg type="i" direction="out"/></method>
<method name="GetRelationSet"><arg type="a(ua(so))" direction="out"/></method>
<method name="GetRole"><arg type="u" direction="out"/></method>
<method name="GetRoleName"><arg type="s" direction="out"/></method>
<method name="GetLocalizedRoleName"><arg type="s" direction="out"/></method>
<method name="GetState"><arg type="au" direction="out"/></method>
<method name="GetAttributes"><arg type="a{ss}" direction="out"/></method>
<method name="GetApplication"><arg type="(so)" direction="out"/></method>
<method name="GetInterfaces"><arg type="as" direction="out"/></method>
</interface>`)
		case atspiComponent:
			b.WriteString(`<interface name="org.a11y.atspi.Component">
<method name="GetExtents"><arg type="u" direction="in"/><arg type="(iiii)" direction="out"/></method>
<method name="GetPosition"><arg type="u" direction="in"/><arg type="i" direction="out"/><arg type="i" direction="out"/></method>
<method name="GetSize"><arg type="i" direction="out"/><arg type="i" direction="out"/></method>
<method name="Contains"><arg type="i" direction="in"/><arg type="i" direction="in"/><arg type="u" direction="in"/><arg type="b" direction="out"/></method>
<method name="GetAccessibleAtPoint"><arg type="i" direction="in"/><arg type="i" direction="in"/><arg type="u" direction="in"/><arg type="(so)" direction="out"/></method>
<method name="GetLayer"><arg type="u" direction="out"/></method>
<method name="GetMDIZOrder"><arg type="n" direction="out"/></method>
<method name="GrabFocus"><arg type="b" direction="out"/></method>
</interface>`)
		case atspiText:
			b.WriteString(`<interface name="org.a11y.atspi.Text">
<method name="GetText"><arg type="i" direction="in"/><arg type="i" direction="in"/><arg type="s" direction="out"/></method>
<method name="GetCharacterAtOffset"><arg type="i" direction="in"/><arg type="i" direction="out"/></method>
<method name="GetStringAtOffset"><arg type="i" direction="in"/><arg type="u" direction="in"/><arg type="s" direction="out"/><arg type="i" direction="out"/><arg type="i" direction="out"/></method>
<method name="GetTextAtOffset"><arg type="i" direction="in"/><arg type="u" direction="in"/><arg type="s" direction="out"/><arg type="i" direction="out"/><arg type="i" direction="out"/></method>
<method name="GetTextBeforeOffset"><arg type="i" direction="in"/><arg type="u" direction="in"/><arg type="s" direction="out"/><arg type="i" direction="out"/><arg type="i" direction="out"/></method>
<method name="GetTextAfterOffset"><arg type="i" direction="in"/><arg type="u" direction="in"/><arg type="s" direction="out"/><arg type="i" direction="out"/><arg type="i" direction="out"/></method>
<method name="GetNSelections"><arg type="i" direction="out"/></method>
<method name="GetSelection"><arg type="i" direction="in"/><arg type="i" direction="out"/><arg type="i" direction="out"/></method>
<method name="GetCharacterExtents"><arg type="i" direction="in"/><arg type="u" direction="in"/><arg type="i" direction="out"/><arg type="i" direction="out"/><arg type="i" direction="out"/><arg type="i" direction="out"/></method>
<method name="GetRangeExtents"><arg type="i" direction="in"/><arg type="i" direction="in"/><arg type="u" direction="in"/><arg type="i" direction="out"/><arg type="i" direction="out"/><arg type="i" direction="out"/><arg type="i" direction="out"/></method>
<method name="GetOffsetAtPoint"><arg type="i" direction="in"/><arg type="i" direction="in"/><arg type="u" direction="in"/><arg type="i" direction="out"/></method>
<method name="GetAttributes"><arg type="i" direction="in"/><arg type="a{ss}" direction="out"/><arg type="i" direction="out"/><arg type="i" direction="out"/></method>
</interface>`)
		case atspiValue:
			b.WriteString(`<interface name="org.a11y.atspi.Value">
<property name="MinimumValue" type="d" access="read"/>
<property name="MaximumValue" type="d" access="read"/>
<property name="MinimumIncrement" type="d" access="read"/>
<property name="CurrentValue" type="d" access="read"/>
<property name="Text" type="s" access="read"/>
</interface>`)
		case atspiAction:
			b.WriteString(`<interface name="org.a11y.atspi.Action">
<property name="NActions" type="i" access="read"/>
<method name="GetName"><arg type="i" direction="in"/><arg type="s" direction="out"/></method>
<method name="GetDescription"><arg type="i" direction="in"/><arg type="s" direction="out"/></method>
<method name="GetLocalizedName"><arg type="i" direction="in"/><arg type="s" direction="out"/></method>
<method name="GetKeyBinding"><arg type="i" direction="in"/><arg type="s" direction="out"/></method>
<method name="GetActions"><arg type="a(sss)" direction="out"/></method>
<method name="DoAction"><arg type="i" direction="in"/><arg type="b" direction="out"/></method>
</interface>`)
		}
	}
	b.WriteString(`<interface name="org.freedesktop.DBus.Properties">
<method name="Get"><arg type="s" direction="in"/><arg type="s" direction="in"/><arg type="v" direction="out"/></method>
<method name="GetAll"><arg type="s" direction="in"/><arg type="a{sv}" direction="out"/></method>
<method name="Set"><arg type="s" direction="in"/><arg type="s" direction="in"/><arg type="v" direction="in"/></method>
</interface></node>`)
	return b.String()
}

const cacheIntrospectXML = `<node><interface name="org.a11y.atspi.Cache">
<property name="version" type="u" access="read"/>
<method name="GetItems"><arg type="a((so)(so)(so)iiassusau)" direction="out"/></method>
<signal name="AddAccessible"><arg type="((so)(so)(so)iiassusau)"/></signal>
<signal name="RemoveAccessible"><arg type="(so)"/></signal>
<signal name="Ready"/>
</interface></node>`
