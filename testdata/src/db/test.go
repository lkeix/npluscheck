package db

import (
	"database/sql"
	"fmt"
)

func Open() *sql.DB {
	dsn := "username:passwd@tcp(127.0.0.1:3306)/test?parseTime=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("Err1")
	}
	return db
}

func getRows(db *sql.DB) *sql.Rows {
	rows, err := db.Query("SELECT * FROM testlist")
	if err != nil {
		fmt.Println("Err2")
		panic(err.Error())
	}
	return rows
}

func execTest(db *sql.DB) error {
	for i := 0; i < 1000; i++ {
		db.Exec("insert into testlist (id) values (?)", i)
	}
	return nil
}

func dbExample() {
	conn := Open()
	rows := getRows(conn)
	getRows(conn)
	execTest(conn)
	for true {
		getRows(conn)
	}
	fmt.Println(rows)
}
