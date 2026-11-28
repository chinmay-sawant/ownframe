package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template/parse"
)

type typeKind int

const (
	tkUnknown typeKind = iota
	tkString
	tkRawHTML
	tkInt
	tkUint
	tkFloat
	tkBool
	tkSlice
	tkMap
	tkPtr
	tkStruct
)

type goTypeInfo struct {
	kind       typeKind
	goName     string
	pkgName    string // empty for local package, else imported package name
	structName string
	elem       *goTypeInfo
}

type structDef struct {
	fields   map[string]ast.Expr
	embedded []string
}

type pkgTypes struct {
	structs    map[string]*structDef            // local struct name -> def
	named      map[string]ast.Expr              // local defined type -> underlying
	methods    map[string]map[string]ast.Expr   // local receiver -> method -> return type
	extStructs map[string]map[string]*structDef // pkgName -> structName -> def
	extNamed   map[string]map[string]ast.Expr   // pkgName -> typeName -> underlying
	extMethods map[string]map[string]map[string]ast.Expr
}

func findModuleRoot(dir string) (root, modPath string) {
	cur, err := filepath.Abs(dir)
	if err != nil {
		return "", ""
	}
	for {
		data, readErr := os.ReadFile(filepath.Join(cur, "go.mod"))
		if readErr == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "module ") {
					return cur, strings.TrimSpace(strings.TrimPrefix(line, "module "))
				}
			}
			return cur, ""
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", ""
		}
		cur = parent
	}
}

func loadPkgTypes(dir string) (*pkgTypes, error) {
	pt := &pkgTypes{
		structs:    make(map[string]*structDef),
		named:      make(map[string]ast.Expr),
		methods:    make(map[string]map[string]ast.Expr),
		extStructs: make(map[string]map[string]*structDef),
		extNamed:   make(map[string]map[string]ast.Expr),
		extMethods: make(map[string]map[string]map[string]ast.Expr),
	}

	imports, err := pt.parseDirInto(dir, pt.structs, pt.named, pt.methods)
	if err != nil {
		return nil, err
	}

	// Also load imported packages within the same Go module so cross-package
	// field types (e.g. terminal.Row) resolve to their exact Go types.
	modRoot, modPath := findModuleRoot(dir)
	if modRoot != "" && modPath != "" {
		loaded := make(map[string]bool)
		for _, impPath := range imports {
			if !strings.HasPrefix(impPath, modPath+"/") {
				continue
			}
			rel := strings.TrimPrefix(impPath, modPath+"/")
			pkgDir := filepath.Join(modRoot, filepath.FromSlash(rel))
			pkgName := filepath.Base(rel)
			if loaded[pkgDir] {
				continue
			}
			loaded[pkgDir] = true
			st := make(map[string]*structDef)
			nm := make(map[string]ast.Expr)
			mt := make(map[string]map[string]ast.Expr)
			if _, loadErr := pt.parseDirInto(pkgDir, st, nm, mt); loadErr == nil {
				pt.extStructs[pkgName] = st
				pt.extNamed[pkgName] = nm
				pt.extMethods[pkgName] = mt
			}
		}
	}

	return pt, nil
}

func (pt *pkgTypes) parseDirInto(
	dir string,
	structs map[string]*structDef,
	named map[string]ast.Expr,
	methods map[string]map[string]ast.Expr,
) ([]string, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, fmt.Errorf("parse Go package in %s: %w", dir, err)
	}

	var imports []string
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			for _, imp := range f.Imports {
				if p, unqErr := strconv.Unquote(imp.Path.Value); unqErr == nil {
					imports = append(imports, p)
				}
			}
			for _, decl := range f.Decls {
				switch d := decl.(type) {
				case *ast.GenDecl:
					if d.Tok != token.TYPE {
						continue
					}
					for _, spec := range d.Specs {
						ts, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}
						if st, ok := ts.Type.(*ast.StructType); ok {
							sd := &structDef{fields: make(map[string]ast.Expr)}
							if st.Fields != nil {
								for _, field := range st.Fields.List {
									if len(field.Names) == 0 {
										if emb := identName(field.Type); emb != "" {
											sd.embedded = append(sd.embedded, emb)
										}
										continue
									}
									for _, name := range field.Names {
										sd.fields[name.Name] = field.Type
									}
								}
							}
							structs[ts.Name.Name] = sd
						} else {
							named[ts.Name.Name] = ts.Type
						}
					}
				case *ast.FuncDecl:
					if d.Recv == nil || len(d.Recv.List) == 0 {
						continue
					}
					if d.Type.Params != nil && len(d.Type.Params.List) > 0 {
						continue
					}
					if d.Type.Results == nil || len(d.Type.Results.List) == 0 {
						continue
					}
					recvName := identName(d.Recv.List[0].Type)
					if recvName == "" {
						continue
					}
					if methods[recvName] == nil {
						methods[recvName] = make(map[string]ast.Expr)
					}
					methods[recvName][d.Name.Name] = d.Type.Results.List[0].Type
				}
			}
		}
	}

	return imports, nil
}

func identName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return identName(e.X)
	}
	return ""
}

func (pt *pkgTypes) resolveExpr(expr ast.Expr) *goTypeInfo {
	return pt.resolveExprIn("", expr)
}

func (pt *pkgTypes) resolveExprIn(pkgName string, expr ast.Expr) *goTypeInfo {
	if expr == nil {
		return &goTypeInfo{kind: tkUnknown}
	}
	switch e := expr.(type) {
	case *ast.Ident:
		return pt.resolveNameIn(pkgName, e.Name)
	case *ast.StarExpr:
		return &goTypeInfo{kind: tkPtr, elem: pt.resolveExprIn(pkgName, e.X)}
	case *ast.ArrayType:
		return &goTypeInfo{kind: tkSlice, elem: pt.resolveExprIn(pkgName, e.Elt)}
	case *ast.MapType:
		return &goTypeInfo{kind: tkMap, elem: pt.resolveExprIn(pkgName, e.Value)}
	case *ast.SelectorExpr:
		if pkgIdent, ok := e.X.(*ast.Ident); ok {
			if pkgIdent.Name == "template" {
				switch e.Sel.Name {
				case "HTML", "CSS", "JS", "URL", "HTMLAttr":
					return &goTypeInfo{kind: tkRawHTML, goName: "template." + e.Sel.Name}
				}
			}
			return pt.resolveNameIn(pkgIdent.Name, e.Sel.Name)
		}
		return &goTypeInfo{kind: tkUnknown}
	case *ast.StructType:
		return &goTypeInfo{kind: tkStruct}
	default:
		return &goTypeInfo{kind: tkUnknown}
	}
}

func (pt *pkgTypes) resolveName(name string) *goTypeInfo {
	return pt.resolveNameIn("", name)
}

func (pt *pkgTypes) resolveNameIn(pkgName, name string) *goTypeInfo {
	switch name {
	case "string":
		return &goTypeInfo{kind: tkString, goName: "string"}
	case "bool":
		return &goTypeInfo{kind: tkBool, goName: "bool"}
	case "int", "int8", "int16", "int32", "int64", "rune":
		return &goTypeInfo{kind: tkInt, goName: name}
	case "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "byte":
		return &goTypeInfo{kind: tkUint, goName: name}
	case "float32", "float64":
		return &goTypeInfo{kind: tkFloat, goName: name}
	}
	if pkgName == "" {
		if _, ok := pt.structs[name]; ok {
			return &goTypeInfo{kind: tkStruct, goName: name, structName: name}
		}
		if underlying, ok := pt.named[name]; ok {
			return pt.resolveExprIn("", underlying)
		}
		return &goTypeInfo{kind: tkUnknown, goName: name}
	}
	if st := pt.extStructs[pkgName]; st != nil {
		if _, ok := st[name]; ok {
			return &goTypeInfo{kind: tkStruct, goName: pkgName + "." + name, pkgName: pkgName, structName: name}
		}
	}
	if nm := pt.extNamed[pkgName]; nm != nil {
		if underlying, ok := nm[name]; ok {
			return pt.resolveExprIn(pkgName, underlying)
		}
	}
	return &goTypeInfo{kind: tkUnknown, goName: pkgName + "." + name}
}

func (pt *pkgTypes) lookupField(t *goTypeInfo, field string) (*goTypeInfo, bool) {
	if t == nil {
		return &goTypeInfo{kind: tkUnknown}, false
	}
	if t.kind == tkPtr {
		return pt.lookupField(t.elem, field)
	}
	if t.kind != tkStruct || t.structName == "" {
		return &goTypeInfo{kind: tkUnknown}, false
	}
	return pt.lookupStructField(t.pkgName, t.structName, field, make(map[string]bool))
}

func (pt *pkgTypes) lookupStructField(pkgName, structName, field string, visited map[string]bool) (*goTypeInfo, bool) {
	key := pkgName + "." + structName
	if visited[key] {
		return &goTypeInfo{kind: tkUnknown}, false
	}
	visited[key] = true

	var sd *structDef
	var ms map[string]ast.Expr
	if pkgName == "" {
		sd = pt.structs[structName]
		ms = pt.methods[structName]
	} else {
		if st := pt.extStructs[pkgName]; st != nil {
			sd = st[structName]
		}
		if em := pt.extMethods[pkgName]; em != nil {
			ms = em[structName]
		}
	}
	if sd == nil {
		return &goTypeInfo{kind: tkUnknown}, false
	}
	if expr, ok := sd.fields[field]; ok {
		return pt.resolveExprIn(pkgName, expr), false
	}
	if ms != nil {
		if ret, ok := ms[field]; ok {
			return pt.resolveExprIn(pkgName, ret), true
		}
	}
	for _, emb := range sd.embedded {
		if info, isMethod := pt.lookupStructField(pkgName, emb, field, visited); info.kind != tkUnknown {
			return info, isMethod
		}
	}
	return &goTypeInfo{kind: tkUnknown}, false
}

// CompileBoundTemplate compiles a full Go html/template source string against an
// existing Go struct type in bindDir and emits a zero-reflection Go renderer.
func CompileBoundTemplate(pkgName, funcName, bindType, bindDir, sourceLabel, input string) ([]byte, error) {
	if !isValidIdent(pkgName) {
		return nil, fmt.Errorf("invalid Go package name %q", pkgName)
	}
	if !isValidIdent(bindType) {
		return nil, fmt.Errorf("invalid bind type name %q", bindType)
	}
	if funcName == "" {
		funcName = "Render" + bindType
	}
	if !isValidIdent(funcName) {
		return nil, fmt.Errorf("invalid func name %q", funcName)
	}

	types, err := loadPkgTypes(bindDir)
	if err != nil {
		return nil, err
	}
	rootType := types.resolveName(bindType)
	if rootType.kind != tkStruct {
		return nil, fmt.Errorf("bind type %q not found as struct in %s", bindType, bindDir)
	}

	builtins := map[string]any{
		"and": func() {}, "or": func() {}, "not": func() {},
		"eq": func() {}, "ne": func() {},
		"lt": func() {}, "le": func() {}, "gt": func() {}, "ge": func() {},
		"len": func() {}, "index": func() {}, "call": func() {},
		"print": func() {}, "printf": func() {}, "println": func() {},
		"html": func() {}, "js": func() {}, "urlquery": func() {},
	}
	trees, err := parse.Parse("root", input, "{{", "}}", builtins)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	rootTree, ok := trees["root"]
	if !ok || rootTree == nil || rootTree.Root == nil {
		return nil, fmt.Errorf("template has no root tree")
	}

	gc := &boundCompiler{
		types:    types,
		trees:    trees,
		rootType: &goTypeInfo{kind: tkPtr, elem: rootType, goName: "*" + bindType},
		bindType: bindType,
		funcName: funcName,
	}

	var bodyBuf bytes.Buffer
	scope := &compileScope{
		dotExpr: "v",
		dotType: gc.rootType,
		vars:    make(map[string]boundVar),
	}
	if err := gc.emitList(&bodyBuf, rootTree.Root, scope, 1); err != nil {
		return nil, err
	}

	// Emit defined sub-templates that were referenced.
	var subBuf bytes.Buffer
	sort.Strings(gc.usedSubNames)
	for _, name := range gc.usedSubNames {
		subTree := trees[name]
		if subTree == nil || subTree.Root == nil {
			continue
		}
		subFn := gc.subFuncName(name)
		var subBody bytes.Buffer
		subScope := &compileScope{
			dotExpr: "v",
			dotType: gc.rootType,
			vars:    make(map[string]boundVar),
		}
		if err := gc.emitList(&subBody, subTree.Root, subScope, 1); err != nil {
			return nil, fmt.Errorf("template %q: %w", name, err)
		}
		fmt.Fprintf(&subBuf, "func %s(b *bytes.Buffer, v *%s) {\n", subFn, bindType)
		if strings.Contains(subBody.String(), "numBuf") {
			subBuf.WriteString("\tvar numBuf [32]byte\n")
		}
		subBuf.Write(subBody.Bytes())
		subBuf.WriteString("}\n\n")
	}

	var out bytes.Buffer
	out.WriteString("// Code generated by ownframegen -bind; DO NOT EDIT.\n")
	if sourceLabel != "" {
		fmt.Fprintf(&out, "// Source: %s\n", sourceLabel)
	}
	fmt.Fprintf(&out, "\npackage %s\n\n", pkgName)

	imports := []string{"bytes", "io"}
	if gc.needFmt {
		imports = append(imports, "fmt")
	}
	if gc.needHTML {
		imports = append(imports, "html")
	}
	if gc.needStrconv {
		imports = append(imports, "strconv")
	}
	sort.Strings(imports)

	out.WriteString("import (\n")
	for _, imp := range imports {
		fmt.Fprintf(&out, "\t%q\n", imp)
	}
	out.WriteString(")\n\n")

	staticBytes := estimateParseStaticBytes(rootTree.Root)
	fmt.Fprintf(&out, "// %sBuf writes the compiled HTML for v directly into b without reflection.\n", funcName)
	fmt.Fprintf(&out, "func %sBuf(b *bytes.Buffer, v *%s) {\n", funcName, bindType)
	out.WriteString("\tif v == nil {\n\t\treturn\n\t}\n")
	if staticBytes > 0 {
		fmt.Fprintf(&out, "\tb.Grow(%d)\n", staticBytes)
	}
	if strings.Contains(bodyBuf.String(), "numBuf") {
		out.WriteString("\tvar numBuf [32]byte\n")
	}
	out.Write(bodyBuf.Bytes())
	out.WriteString("}\n\n")

	fmt.Fprintf(&out, "// %s implements ownframe.Config.Render without html/template reflection.\n", funcName)
	fmt.Fprintf(&out, "func %s(w io.Writer, data any) error {\n", funcName)
	fmt.Fprintf(&out, "\tvar src *%s\n", bindType)
	out.WriteString("\tswitch d := data.(type) {\n")
	fmt.Fprintf(&out, "\tcase *%s:\n\t\tsrc = d\n", bindType)
	fmt.Fprintf(&out, "\tcase %s:\n\t\tsrc = &d\n", bindType)
	out.WriteString("\t}\n")
	out.WriteString("\tif src == nil {\n\t\treturn nil\n\t}\n")
	out.WriteString("\tif buf, ok := w.(*bytes.Buffer); ok {\n")
	fmt.Fprintf(&out, "\t\t%sBuf(buf, src)\n", funcName)
	out.WriteString("\t\treturn nil\n")
	out.WriteString("\t}\n")
	out.WriteString("\tvar buf bytes.Buffer\n")
	fmt.Fprintf(&out, "\t%sBuf(&buf, src)\n", funcName)
	out.WriteString("\t_, err := w.Write(buf.Bytes())\n")
	out.WriteString("\treturn err\n")
	out.WriteString("}\n\n")

	out.Write(subBuf.Bytes())

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		return nil, fmt.Errorf("go/format failed: %w\n--- raw source ---\n%s", err, out.String())
	}
	return formatted, nil
}

type boundVar struct {
	goExpr string
	typ    *goTypeInfo
}

type compileScope struct {
	parent  *compileScope
	dotExpr string
	dotType *goTypeInfo
	vars    map[string]boundVar
}

func (s *compileScope) child(dotExpr string, dotType *goTypeInfo) *compileScope {
	return &compileScope{
		parent:  s,
		dotExpr: dotExpr,
		dotType: dotType,
		vars:    make(map[string]boundVar),
	}
}

func (s *compileScope) lookupVar(name string) (boundVar, bool) {
	for cur := s; cur != nil; cur = cur.parent {
		if v, ok := cur.vars[name]; ok {
			return v, true
		}
	}
	return boundVar{}, false
}

type boundCompiler struct {
	types        *pkgTypes
	trees        map[string]*parse.Tree
	rootType     *goTypeInfo
	bindType     string
	funcName     string
	seq          int
	needHTML     bool
	needStrconv  bool
	needFmt      bool
	usedSubs     map[string]bool
	usedSubNames []string
}

func (gc *boundCompiler) nextID(prefix string) string {
	gc.seq++
	return fmt.Sprintf("%s%d", prefix, gc.seq)
}

func (gc *boundCompiler) subFuncName(name string) string {
	return "render" + gc.funcName + "Sub_" + toPascalCase(name)
}

func (gc *boundCompiler) markSubUsed(name string) {
	if gc.usedSubs == nil {
		gc.usedSubs = make(map[string]bool)
	}
	if !gc.usedSubs[name] {
		gc.usedSubs[name] = true
		gc.usedSubNames = append(gc.usedSubNames, name)
	}
}

func (gc *boundCompiler) emitList(buf *bytes.Buffer, list *parse.ListNode, scope *compileScope, indent int) error {
	if list == nil {
		return nil
	}
	for _, node := range list.Nodes {
		if err := gc.emitNode(buf, node, scope, indent); err != nil {
			return err
		}
	}
	return nil
}

func (gc *boundCompiler) emitNode(buf *bytes.Buffer, node parse.Node, scope *compileScope, indent int) error {
	tabs := strings.Repeat("\t", indent)
	switch n := node.(type) {
	case *parse.TextNode:
		if len(n.Text) > 0 {
			fmt.Fprintf(buf, "%sb.WriteString(%s)\n", tabs, strconv.Quote(string(n.Text)))
		}
		return nil

	case *parse.ActionNode:
		if len(n.Pipe.Decl) > 0 {
			valExpr, valType, err := gc.evalPipe(n.Pipe, scope, false)
			if err != nil {
				return err
			}
			varName := n.Pipe.Decl[0].Ident[0]
			goVar := gc.nextID("v_")
			fmt.Fprintf(buf, "%s%s := %s\n", tabs, goVar, valExpr)
			fmt.Fprintf(buf, "%s_ = %s\n", tabs, goVar)
			scope.vars[varName] = boundVar{goExpr: goVar, typ: valType}
			return nil
		}
		valExpr, valType, err := gc.evalPipe(n.Pipe, scope, false)
		if err != nil {
			return err
		}
		gc.emitWrite(buf, tabs, valExpr, valType)
		return nil

	case *parse.IfNode:
		return gc.emitIf(buf, n, scope, indent, false)

	case *parse.WithNode:
		valExpr, valType, err := gc.evalPipe(n.Pipe, scope, false)
		if err != nil {
			return err
		}
		wVar := gc.nextID("w")
		fmt.Fprintf(buf, "%sif %s := %s; %s {\n", tabs, wVar, valExpr, gc.truthyExpr(wVar, valType))
		wScope := scope.child(wVar, valType)
		if err := gc.emitList(buf, n.List, wScope, indent+1); err != nil {
			return err
		}
		if n.ElseList != nil {
			fmt.Fprintf(buf, "%s} else {\n", tabs)
			if err := gc.emitList(buf, n.ElseList, scope, indent+1); err != nil {
				return err
			}
		}
		fmt.Fprintf(buf, "%s}\n", tabs)
		return nil

	case *parse.RangeNode:
		sliceExpr, sliceType, err := gc.evalPipe(n.Pipe, scope, false)
		if err != nil {
			return err
		}
		elemType := &goTypeInfo{kind: tkUnknown}
		if sliceType != nil && sliceType.elem != nil {
			elemType = sliceType.elem
		}
		idxVar := gc.nextID("i")
		itemVar := gc.nextID("item")

		if n.ElseList != nil {
			fmt.Fprintf(buf, "%sif len(%s) > 0 {\n", tabs, sliceExpr)
			indent++
			tabs = strings.Repeat("\t", indent)
		}

		fmt.Fprintf(buf, "%sfor %s := range %s {\n", tabs, idxVar, sliceExpr)
		dotType := elemType
		if elemType.kind == tkStruct {
			fmt.Fprintf(buf, "%s\t%s := &%s[%s]\n", tabs, itemVar, sliceExpr, idxVar)
			dotType = &goTypeInfo{kind: tkPtr, elem: elemType, goName: "*" + elemType.goName}
		} else {
			fmt.Fprintf(buf, "%s\t%s := %s[%s]\n", tabs, itemVar, sliceExpr, idxVar)
		}
		fmt.Fprintf(buf, "%s\t_ = %s\n", tabs, itemVar)

		rScope := scope.child(itemVar, dotType)
		switch len(n.Pipe.Decl) {
		case 1:
			rScope.vars[n.Pipe.Decl[0].Ident[0]] = boundVar{goExpr: itemVar, typ: dotType}
		case 2:
			rScope.vars[n.Pipe.Decl[0].Ident[0]] = boundVar{goExpr: idxVar, typ: &goTypeInfo{kind: tkInt, goName: "int"}}
			rScope.vars[n.Pipe.Decl[1].Ident[0]] = boundVar{goExpr: itemVar, typ: dotType}
		}
		if err := gc.emitList(buf, n.List, rScope, indent+1); err != nil {
			return err
		}
		fmt.Fprintf(buf, "%s}\n", tabs)

		if n.ElseList != nil {
			indent--
			tabs = strings.Repeat("\t", indent)
			fmt.Fprintf(buf, "%s} else {\n", tabs)
			if err := gc.emitList(buf, n.ElseList, scope, indent+1); err != nil {
				return err
			}
			fmt.Fprintf(buf, "%s}\n", tabs)
		}
		return nil

	case *parse.TemplateNode:
		if _, ok := gc.trees[n.Name]; !ok {
			return fmt.Errorf("undefined template %q", n.Name)
		}
		gc.markSubUsed(n.Name)
		argExpr := "v"
		if n.Pipe != nil {
			e, _, err := gc.evalPipe(n.Pipe, scope, false)
			if err != nil {
				return err
			}
			argExpr = e
		}
		fmt.Fprintf(buf, "%s%s(b, %s)\n", tabs, gc.subFuncName(n.Name), argExpr)
		return nil

	case *parse.CommentNode:
		return nil

	default:
		return fmt.Errorf("unsupported template node %T", node)
	}
}

func (gc *boundCompiler) emitIf(buf *bytes.Buffer, n *parse.IfNode, scope *compileScope, indent int, isElseIf bool) error {
	tabs := strings.Repeat("\t", indent)
	condExpr, condType, err := gc.evalPipe(n.Pipe, scope, true)
	if err != nil {
		return err
	}
	cond := gc.truthyExpr(condExpr, condType)
	if isElseIf {
		fmt.Fprintf(buf, "%s} else if %s {\n", tabs, cond)
	} else {
		fmt.Fprintf(buf, "%sif %s {\n", tabs, cond)
	}
	if err := gc.emitList(buf, n.List, scope, indent+1); err != nil {
		return err
	}
	if n.ElseList != nil {
		if len(n.ElseList.Nodes) == 1 {
			if elseIf, ok := n.ElseList.Nodes[0].(*parse.IfNode); ok {
				return gc.emitIf(buf, elseIf, scope, indent, true)
			}
		}
		fmt.Fprintf(buf, "%s} else {\n", tabs)
		if err := gc.emitList(buf, n.ElseList, scope, indent+1); err != nil {
			return err
		}
	}
	fmt.Fprintf(buf, "%s}\n", tabs)
	return nil
}

func (gc *boundCompiler) emitWrite(buf *bytes.Buffer, tabs, expr string, typ *goTypeInfo) {
	if typ == nil {
		typ = &goTypeInfo{kind: tkUnknown}
	}
	switch typ.kind {
	case tkRawHTML:
		fmt.Fprintf(buf, "%sb.WriteString(string(%s))\n", tabs, expr)
	case tkString:
		gc.needHTML = true
		fmt.Fprintf(buf, "%sb.WriteString(html.EscapeString(%s))\n", tabs, expr)
	case tkInt:
		gc.needStrconv = true
		fmt.Fprintf(buf, "%sb.Write(strconv.AppendInt(numBuf[:0], int64(%s), 10))\n", tabs, expr)
	case tkUint:
		gc.needStrconv = true
		fmt.Fprintf(buf, "%sb.Write(strconv.AppendUint(numBuf[:0], uint64(%s), 10))\n", tabs, expr)
	case tkFloat:
		gc.needStrconv = true
		fmt.Fprintf(buf, "%sb.Write(strconv.AppendFloat(numBuf[:0], float64(%s), 'g', -1, 64))\n", tabs, expr)
	case tkBool:
		gc.needStrconv = true
		fmt.Fprintf(buf, "%sb.WriteString(strconv.FormatBool(%s))\n", tabs, expr)
	default:
		gc.needFmt = true
		gc.needHTML = true
		fmt.Fprintf(buf, "%sb.WriteString(html.EscapeString(fmt.Sprint(%s)))\n", tabs, expr)
	}
}

func (gc *boundCompiler) truthyExpr(expr string, typ *goTypeInfo) string {
	if typ == nil {
		return expr + " != nil"
	}
	switch typ.kind {
	case tkBool:
		return expr
	case tkString, tkRawHTML:
		return expr + ` != ""`
	case tkInt, tkUint, tkFloat:
		return expr + " != 0"
	case tkSlice, tkMap:
		return "len(" + expr + ") > 0"
	case tkPtr:
		return expr + " != nil"
	case tkStruct:
		return "true"
	default:
		return expr + " != nil"
	}
}

func (gc *boundCompiler) evalPipe(pipe *parse.PipeNode, scope *compileScope, boolCtx bool) (string, *goTypeInfo, error) {
	if pipe == nil || len(pipe.Cmds) == 0 {
		return "", nil, fmt.Errorf("empty pipeline")
	}
	if len(pipe.Cmds) > 1 {
		return "", nil, fmt.Errorf("chained pipelines (|) are not yet supported: %s", pipe.String())
	}
	return gc.evalCommand(pipe.Cmds[0], scope, boolCtx)
}

func (gc *boundCompiler) evalCommand(cmd *parse.CommandNode, scope *compileScope, boolCtx bool) (string, *goTypeInfo, error) {
	if len(cmd.Args) == 0 {
		return "", nil, fmt.Errorf("empty command")
	}
	first := cmd.Args[0]
	if id, ok := first.(*parse.IdentifierNode); ok {
		return gc.evalBuiltin(id.Ident, cmd.Args[1:], scope, boolCtx)
	}
	if len(cmd.Args) > 1 {
		return "", nil, fmt.Errorf("unexpected arguments in command %s", cmd.String())
	}
	return gc.evalArg(first, scope, boolCtx)
}

func (gc *boundCompiler) evalBuiltin(name string, args []parse.Node, scope *compileScope, boolCtx bool) (string, *goTypeInfo, error) {
	boolType := &goTypeInfo{kind: tkBool, goName: "bool"}
	switch name {
	case "not":
		if len(args) != 1 {
			return "", nil, fmt.Errorf("not expects 1 arg, got %d", len(args))
		}
		e, t, err := gc.evalArg(args[0], scope, true)
		if err != nil {
			return "", nil, err
		}
		return "!(" + gc.truthyExpr(e, t) + ")", boolType, nil

	case "and":
		if len(args) == 0 {
			return "", nil, fmt.Errorf("and expects >= 1 arg")
		}
		parts := make([]string, len(args))
		for i, a := range args {
			e, t, err := gc.evalArg(a, scope, true)
			if err != nil {
				return "", nil, err
			}
			parts[i] = "(" + gc.truthyExpr(e, t) + ")"
		}
		return "(" + strings.Join(parts, " && ") + ")", boolType, nil

	case "or":
		if len(args) == 0 {
			return "", nil, fmt.Errorf("or expects >= 1 arg")
		}
		parts := make([]string, len(args))
		for i, a := range args {
			e, t, err := gc.evalArg(a, scope, true)
			if err != nil {
				return "", nil, err
			}
			parts[i] = "(" + gc.truthyExpr(e, t) + ")"
		}
		return "(" + strings.Join(parts, " || ") + ")", boolType, nil

	case "eq":
		if len(args) < 2 {
			return "", nil, fmt.Errorf("eq expects >= 2 args")
		}
		lhs, _, err := gc.evalArg(args[0], scope, false)
		if err != nil {
			return "", nil, err
		}
		if len(args) == 2 {
			rhs, _, err := gc.evalArg(args[1], scope, false)
			if err != nil {
				return "", nil, err
			}
			return "(" + lhs + " == " + rhs + ")", boolType, nil
		}
		parts := make([]string, 0, len(args)-1)
		for _, a := range args[1:] {
			rhs, _, err := gc.evalArg(a, scope, false)
			if err != nil {
				return "", nil, err
			}
			parts = append(parts, "("+lhs+" == "+rhs+")")
		}
		return "(" + strings.Join(parts, " || ") + ")", boolType, nil

	case "ne", "lt", "le", "gt", "ge":
		if len(args) != 2 {
			return "", nil, fmt.Errorf("%s expects 2 args, got %d", name, len(args))
		}
		lhs, _, err := gc.evalArg(args[0], scope, false)
		if err != nil {
			return "", nil, err
		}
		rhs, _, err := gc.evalArg(args[1], scope, false)
		if err != nil {
			return "", nil, err
		}
		op := map[string]string{"ne": "!=", "lt": "<", "le": "<=", "gt": ">", "ge": ">="}[name]
		return "(" + lhs + " " + op + " " + rhs + ")", boolType, nil

	case "len":
		if len(args) != 1 {
			return "", nil, fmt.Errorf("len expects 1 arg, got %d", len(args))
		}
		e, _, err := gc.evalArg(args[0], scope, false)
		if err != nil {
			return "", nil, err
		}
		return "len(" + e + ")", &goTypeInfo{kind: tkInt, goName: "int"}, nil

	default:
		return "", nil, fmt.Errorf("unsupported template builtin %q", name)
	}
}

func (gc *boundCompiler) evalArg(node parse.Node, scope *compileScope, boolCtx bool) (string, *goTypeInfo, error) {
	switch a := node.(type) {
	case *parse.DotNode:
		return scope.dotExpr, scope.dotType, nil

	case *parse.FieldNode:
		return gc.resolveFieldChain(scope.dotExpr, scope.dotType, a.Ident)

	case *parse.VariableNode:
		if len(a.Ident) == 0 {
			return "", nil, fmt.Errorf("empty variable")
		}
		var baseExpr string
		var baseType *goTypeInfo
		if a.Ident[0] == "$" {
			baseExpr = "v"
			baseType = gc.rootType
		} else {
			bv, ok := scope.lookupVar(a.Ident[0])
			if !ok {
				return "", nil, fmt.Errorf("undefined template variable %q", a.Ident[0])
			}
			baseExpr = bv.goExpr
			baseType = bv.typ
		}
		return gc.resolveFieldChain(baseExpr, baseType, a.Ident[1:])

	case *parse.PipeNode:
		return gc.evalPipe(a, scope, boolCtx)

	case *parse.StringNode:
		return strconv.Quote(a.Text), &goTypeInfo{kind: tkString, goName: "string"}, nil

	case *parse.NumberNode:
		if a.IsInt {
			return strconv.FormatInt(a.Int64, 10), &goTypeInfo{kind: tkInt, goName: "int"}, nil
		}
		if a.IsUint {
			return strconv.FormatUint(a.Uint64, 10), &goTypeInfo{kind: tkUint, goName: "uint64"}, nil
		}
		if a.IsFloat {
			return strconv.FormatFloat(a.Float64, 'g', -1, 64), &goTypeInfo{kind: tkFloat, goName: "float64"}, nil
		}
		return a.Text, &goTypeInfo{kind: tkInt, goName: "int"}, nil

	case *parse.BoolNode:
		return strconv.FormatBool(a.True), &goTypeInfo{kind: tkBool, goName: "bool"}, nil

	case *parse.ChainNode:
		baseExpr, baseType, err := gc.evalArg(a.Node, scope, false)
		if err != nil {
			return "", nil, err
		}
		return gc.resolveFieldChain(baseExpr, baseType, a.Field)

	default:
		return "", nil, fmt.Errorf("unsupported template expression %T (%s)", node, node.String())
	}
}

func (gc *boundCompiler) resolveFieldChain(baseExpr string, baseType *goTypeInfo, fields []string) (string, *goTypeInfo, error) {
	expr := baseExpr
	curType := baseType
	for _, f := range fields {
		nextType, isMethod := gc.types.lookupField(curType, f)
		if isMethod {
			expr = expr + "." + f + "()"
		} else {
			expr = expr + "." + f
		}
		curType = nextType
	}
	return expr, curType, nil
}

func estimateParseStaticBytes(list *parse.ListNode) int {
	if list == nil {
		return 0
	}
	total := 0
	for _, node := range list.Nodes {
		switch n := node.(type) {
		case *parse.TextNode:
			total += len(n.Text)
		case *parse.IfNode:
			total += estimateParseStaticBytes(n.List) + estimateParseStaticBytes(n.ElseList)
		case *parse.WithNode:
			total += estimateParseStaticBytes(n.List) + estimateParseStaticBytes(n.ElseList)
		case *parse.RangeNode:
			total += estimateParseStaticBytes(n.List) + estimateParseStaticBytes(n.ElseList)
		}
	}
	return total
}

func resolveBoundInputs(inSpec string) ([]string, error) {
	if strings.Contains(inSpec, ",") {
		parts := strings.Split(inSpec, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			resolved, err := resolveInputs(p)
			if err != nil {
				return nil, err
			}
			out = append(out, resolved...)
		}
		return out, nil
	}
	return resolveInputs(inSpec)
}

func defaultBindDir(outPath string) string {
	if outPath != "" && outPath != "-" {
		return filepath.Dir(outPath)
	}
	return "."
}
