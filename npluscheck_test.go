package npluscheck

import (
	"testing"

	_ "github.com/jmoiron/sqlx"
	"golang.org/x/tools/go/analysis/analysistest"
	_ "gorm.io/driver/sqlite"
	_ "gorm.io/gorm"
)

func TestExample(t *testing.T) {
	td := analysistest.TestData()
	analysistest.Run(t, td, Analyzer, "example")
}

func TestStdDB(t *testing.T) {
	td := analysistest.TestData()
	analysistest.Run(t, td, Analyzer, "db")
}

/*
func TestGorm(t *testing.T) {
	td := analysistest.TestData()
	analysistest.Run(t, td, Analyzer, "gorm_example")
}

func TestSqlx(t *testing.T) {
	td := analysistest.TestData()
	analysistest.Run(t, td, Analyzer, "sqlx_example")
}
*/
