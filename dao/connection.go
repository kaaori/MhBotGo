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
	// DB *sqlite3.Conn
)

// GetConnection : Gets a MHBot DB connection
func GetConnection(connString string) *sqlite3.Conn {
	DB, err := sqlite3.Open(connString)
	if err != nil {
		log.Error("Error Getting DB connection", err)
		panic("Error connecting to sqlite")
	}
	DB.BusyTimeout(300 * time.Millisecond)
	return DB
}
