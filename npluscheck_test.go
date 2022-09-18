package npluscheck

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestRun(t *testing.T) {
	td := analysistest.TestData()
	analysistest.Run(t, td, Analyzer, "example")
}
