package main

import (
	"github.com/lkeix/npluscheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(npluscheck.Analyzer)
}
