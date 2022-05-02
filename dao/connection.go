package dao

import (

	// TODO
	"log"
	"time"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
)

// GetConnection : Gets a MHBot DB connection
func GetConnection(connString string) *sqlite3.Conn {
	DB, err := sqlite3.Open(connString)
	if err != nil {
		log.Fatal("Error Getting DB connection", err)
		panic("Error connecting to sqlite")
	}
	DB.BusyTimeout(300 * time.Millisecond)
	return DB
}
