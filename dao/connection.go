package dao

import (
	"database/sql"
	"log"

	// TODO
	_ "github.com/mattn/go-sqlite3"
)

func get() *sql.DB {
	db, err := sql.Open("sqlite3", "../data/MHBot.db")
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	return db
}
