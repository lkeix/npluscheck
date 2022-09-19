package gorm_example

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID   string
	Name string
}

func Open() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}

func Insert(u User) {
	db := Open()
	db.Model(&User{}).Create(u)
}

func example() {
	u := User{
		ID:   "hoge",
		Name: "Fuga",
	}
	for i := 0; i < 10; i++ {
		Insert(u)
	}
}
