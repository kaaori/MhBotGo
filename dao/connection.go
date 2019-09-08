package dao

import (

	// TODO
	"time"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
	logging "github.com/kaaori/mhbotgo/log"
)

var (
	log = logging.NewLog()
	// DB : The current DB session
	DB *sqlite3.Conn
)

func init() {
	log.Info("Initialising DB...")
	db, err := sqlite3.Open("./data/MHBot.db")
	// db.SetMaxOpenConns(1)
	// db.SetConnMaxLifetime(0)
	if err != nil {
		log.Error("Error Getting DB connection", err)
		panic("Error connecting to sqlite")
	}
	db.BusyTimeout(5 * time.Second)
	DB = db
}
