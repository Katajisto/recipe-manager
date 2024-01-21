package main

import "database/sql"
import _ "github.com/mattn/go-sqlite3"

type Recipe struct {
	Name        string
	Description string
	// Skip ingredient list things for now, since I feel like that is way too much effort for this MVP.
}

var Db *sql.DB

func createDb() {
	var err error
	Db, err = sql.Open("sqlite3", "data.db")
	if err != nil {
		panic(err)
	}
}

func closeDb() {
	Db.Close()
}
