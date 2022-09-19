package npluscheck

import (
	"go/ast"
	"go/token"
	"go/types"
	"reflect"

	_ "github.com/jmoiron/sqlx"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	_ "gorm.io/driver/sqlite"
	_ "gorm.io/gorm"
)

var UsuallyDataBasePackages = []string{
	"database/sql",
	"github.com/Masterminds/squirrel",
	"gorm.io/gorm",
	"gopkg.in/gorp.v1",
	"github.com/jmoiron/sqlx",
}

var Analyzer = &analysis.Analyzer{
	Name: "npluscheck",
	Doc:  "npluscheck finds execution SQL in iteration",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	ResultType: reflect.TypeOf((*Indicate)(nil)),
}

type customFunc struct {
	expr        interface{}
	calledInFor bool
}

type Indicate struct {
	mp map[token.Pos]bool
}

func run(pass *analysis.Pass) (interface{}, error) {
	info := pass.TypesInfo
	inspect, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	result := &Indicate{
		mp: map[token.Pos]bool{},
	}
	inspect.Nodes(nodeFilter, func(n ast.Node, push bool) bool {
		if !push {
			return false
		}

		result.mp = checkNplus(pass, info, n.(*ast.FuncDecl))
		return true
	})
	return result, nil
}

func checkNplus(pass *analysis.Pass, info *types.Info, fd *ast.FuncDecl) map[token.Pos]bool {
	funcs := extractCalledFuncs(fd.Body.List, false)
	callinLoop := make(map[token.Pos]bool)
	for _, f := range funcs {
		switch expr := f.expr.(type) {
		case *ast.SelectorExpr:
			if typ, ok := info.Selections[expr]; ok {
				if contain(UsuallyDataBasePackages, typ.Obj().Pkg().Path()) && f.calledInFor {
					callinLoop[expr.Pos()] = true
					pass.Reportf(expr.Pos(), "may call DB API in loop")
				}
			}
		}
	}
	return callinLoop
}

func flatten(pkgs []map[string]*ast.Package) map[string]*ast.Package {
	flattened := make(map[string]*ast.Package)
	for _, pkg := range pkgs {
		for key := range pkg {
			flattened[key] = pkg[key]
		}
	}
	return flattened
}

func extractCalledFuncs(stmts []ast.Stmt, calledInFor bool) []customFunc {
	funcs := []customFunc{}
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *ast.ExprStmt:
			if callexpr, ok := stmt.X.(*ast.CallExpr); ok {
				if sel := extractSelectorExprFun(callexpr); sel != nil {
					funcs = append(funcs, customFunc{
						calledInFor: calledInFor,
						expr:        sel,
					})
					continue
				}
				funcs = append(funcs, customFunc{
					calledInFor: calledInFor,
					expr:        callexpr,
				})
			}
		case *ast.ForStmt:
			funcs = append(funcs, extractCalledFuncs(stmt.Body.List, true)...)
		case *ast.IfStmt:
			funcs = append(funcs, extractCalledFuncs(stmt.Body.List, calledInFor)...)
		case *ast.AssignStmt:
			for _, expr := range stmt.Rhs {
				if rhs, ok := expr.(*ast.CallExpr); ok {
					if sel := extractSelectorExprFun(rhs); sel != nil {
						funcs = append(funcs, customFunc{
							calledInFor: calledInFor,
							expr:        sel,
						})
						continue
					}
					funcs = append(funcs, customFunc{
						calledInFor: calledInFor,
						expr:        rhs,
					})
				}
			}
		}
	}
	return funcs
}

func extractSelectorExprFun(expr *ast.CallExpr) *ast.SelectorExpr {
	if sel, ok := expr.Fun.(*ast.SelectorExpr); ok {
		return sel
	}
	return nil
}

func contain(strs []string, str string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
}
