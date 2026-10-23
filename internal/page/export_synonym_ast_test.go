package page

import (
	"errors"
	"go/ast"
)

var errCaller = errors.New("page: unknown test caller")

func pageRecvName(fn *ast.FuncDecl) (string, bool) {
	if fn.Recv == nil || len(fn.Recv.List) != 1 {
		return "", false
	}
	f := fn.Recv.List[0]
	star, ok := f.Type.(*ast.StarExpr)
	if !ok {
		return "", false
	}
	id, ok := star.X.(*ast.Ident)
	if !ok || id.Name != "Page" || len(f.Names) != 1 {
		return "", false
	}
	return f.Names[0].Name, true
}

func isPageSynonym(recv string, body *ast.BlockStmt) bool {
	if len(body.List) != 1 {
		return false
	}
	switch s := body.List[0].(type) {
	case *ast.ExprStmt:
		return pageCall(recv, s.X)
	case *ast.ReturnStmt:
		return len(s.Results) == 1 && pageCall(recv, s.Results[0])
	}
	return false
}

func pageCall(recv string, e ast.Expr) bool {
	call, ok := e.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == recv
}
