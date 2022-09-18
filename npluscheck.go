package npluscheck

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"reflect"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
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
	Run:  runNew,
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

func runNew(pass *analysis.Pass) (interface{}, error) {
	inspect, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	conf := types.Config{
		Importer: importer.Default(),
	}
	info := &types.Info{
		Defs:       map[*ast.Ident]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
	}

	for _, file := range pass.Files {
		conf.Check(pass.Pkg.Name(), pass.Fset, []*ast.File{file}, info)
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

func run(pass *analysis.Pass) (interface{}, error) {
	dirs, _ := os.Getwd()

	fset := token.NewFileSet()

	pkgs := make([]map[string]*ast.Package, 0)

	pkg, err := parser.ParseDir(fset, dirs, nil, 0)
	if err != nil {
		return nil, err
	}
	if len(pkg) != 0 {
		pkgs = append(pkgs, pkg)
	}

	packages := flatten(pkgs)

	conf := types.Config{
		Importer: importer.Default(),
	}
	info := &types.Info{
		Defs:       map[*ast.Ident]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
	}

	for _, pkg := range packages {
		for fileName := range pkg.Files {
			conf.Check(fileName, fset, []*ast.File{pkg.Files[fileName]}, info)
		}
	}

	for _, pkg := range packages {
		ast.Inspect(pkg, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.Ident:
				info.ObjectOf(n)
			case *ast.FuncDecl:
				useDBAPIInFor(info, n)
			}
			return true
		})
	}
	return nil, nil
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

func useDBAPIInFor(info *types.Info, fd *ast.FuncDecl) {
	funcs := extractCalledFuncs(fd.Body.List, false)
	for _, f := range funcs {
		switch expr := f.expr.(type) {
		case *ast.CallExpr:
			// fmt.Println(f.Fun)
			// fmt.Println(info.Selections)
		case *ast.SelectorExpr:
			if typ, ok := info.Selections[expr]; ok {
				if contain(UsuallyDataBasePackages, typ.Obj().Pkg().Path()) && f.calledInFor {
					fmt.Println("called in for statement")
				}
			}
		}
	}
}

func contain(strs []string, str string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
}
