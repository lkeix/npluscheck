package sqlx_example

import (
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

func Insert() {
	db := Open()
	tx := db.MustBegin()
	for i := 0; i < 100; i++ {
		tx.MustExec("INSERT INTO person (first_name, last_name, email) VALUES ($1, $2, $3)", "Jason", "Moiron", "jmoiron@jmoiron.net")
	}
}
