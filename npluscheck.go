package npluscheck

import (
	"go/token"

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

type dependence struct {
	execSQL  bool
	children []token.Pos
}

func run(pass *analysis.Pass) (interface{}, error) {
	return nil, nil
}
