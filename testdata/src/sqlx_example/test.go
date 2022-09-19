package sqlx_example

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

func Open() *sqlx.DB {
	db, err := sqlx.Connect("postgres", "user=foo dbname=bar sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}
	return db
}

func Insert(db *sqlx.DB) {
	tx := db.MustBegin()
	for i := 0; i < 100; i++ {
		tx.MustExec("INSERT INTO person (first_name, last_name, email) VALUES ($1, $2, $3)", "Jason", "Moiron", "jmoiron@jmoiron.net")
	}
}

func getRows(db *sql.DB) *sql.Rows {
	rows, err := db.Query("SELECT * FROM testlist")
	if err != nil {
		fmt.Println("Err2")
		panic(err.Error())
	}
	return rows
}

func example() {
	db := Open()
	Insert(db)
}
