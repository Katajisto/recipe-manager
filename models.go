package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Recipe struct {
	Name        string
	Description string
	// Skip ingredient list things for now, since I feel like that is way too much effort for this MVP.
}

func (r Recipe) Add() error {
	_, err := Db.Exec("INSERT INTO recipes (name, desc) values($1, $2);", r.Name, r.Description)
	return err
}

func GetRecipes() []Recipe {
	recipes := make([]Recipe, 0)

	rows, err := Db.Query("SELECT * FROM recipes;")
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var id int
		var name string
		var desc string
		rows.Scan(&id, &name, &desc)
		recipes = append(recipes, Recipe{name, desc})
	}

	log.Println(recipes)
	return recipes
}

var Db *sql.DB

func createTable() {
	Db.Exec("CREATE TABLE IF NOT EXISTS recipes(id INTEGER PRIMARY KEY, name TEXT NOT NULL, desc TEXT);")
}

func createDb() {
	var err error
	Db, err = sql.Open("sqlite3", "data.db")
	if err != nil {
		panic(err)
	}
	createTable()
}

func closeDb() {
	Db.Close()
}
