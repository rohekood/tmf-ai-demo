package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(postgres.Open("invalid dsn"), &gorm.Config{})
	fmt.Printf("db is nil? %v, err: %v\n", db == nil, err)
	if db != nil {
		sqlDB, e := db.DB()
		fmt.Printf("sqlDB is nil? %v, err: %v\n", sqlDB == nil, e)
	}
}
