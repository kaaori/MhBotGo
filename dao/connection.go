package dao

import (
	"database/sql"

	// TODO
	logging "github.com/kaaori/mhbotgo/log"
)

var (
	log = logging.NewLog()
)

func get() *sql.DB {
	db, err := sql.Open("sqlite3", "./data/MHBot.db")
	if err != nil {
		log.Error("Error Getting DB connection", err)
		return nil
	}
	return db
}
