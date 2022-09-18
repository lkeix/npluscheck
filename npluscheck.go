package npluscheck

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

var Analyzer = &analysis.Analyzer{
	Name: "npluscheck",
	Doc:  "npluscheck finds execution SQL in iteration",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

type parsedFunc struct {
	execSQL bool
	name    string
}

func setup() []string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	dirs := make([]string, 0)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	return dirs
}

func run(pass *analysis.Pass) (interface{}, error) {
	dirs := setup()

	fset := token.NewFileSet()

	pkgs := make([]map[string]*ast.Package, 0)

	for _, dir := range dirs {
		pkg, err := parser.ParseDir(fset, dir, nil, 0)
		if err != nil {
			return nil, err
		}
		if len(pkg) != 0 {
			pkgs = append(pkgs, pkg)
		}
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
				useDBAPIIn(info, n)
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

func extractCalledFuncs(stmts []ast.Stmt) []interface{} {
	funcs := []interface{}{}
	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case *ast.ExprStmt:
			if callexpr, ok := stmt.X.(*ast.CallExpr); ok {
				if sel := extractSelectorExprFun(callexpr); sel != nil {
					funcs = append(funcs, stmt.X)
					continue
				}
				funcs = append(funcs, callexpr)
			}
		case *ast.ForStmt:
			funcs = append(funcs, extractCalledFuncs(stmt.Body.List)...)
		case *ast.IfStmt:
			funcs = append(funcs, extractCalledFuncs(stmt.Body.List)...)
		case *ast.AssignStmt:
			for _, expr := range stmt.Rhs {
				if rhs, ok := expr.(*ast.CallExpr); ok {
					if sel := extractSelectorExprFun(rhs); sel != nil {
						funcs = append(funcs, sel)
						continue
					}
					funcs = append(funcs, rhs)
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

func useDBAPIIn(info *types.Info, fd *ast.FuncDecl) {
	// fmt.Printf("%v:\n", fd.Name)
	funcs := extractCalledFuncs(fd.Body.List)
	for _, f := range funcs {
		switch f := f.(type) {
		case *ast.CallExpr:
			// fmt.Println(f.Fun)
			fmt.Println(info.Types[f])
		case *ast.SelectorExpr:
			fmt.Println(info.Selections[f])
		}
	}
}
