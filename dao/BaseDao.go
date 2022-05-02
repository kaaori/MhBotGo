package dao

import (
	"fmt"
	"log"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
)

var (
	// ConnString : The connection string to the default DB directory
	ConnString = "file:./data/MHBot.db?cache=shared&mode=rwc"
)

// SetConnString : Updates the connection string for GetConnection calls; mostly used for testing
func SetConnString(newConnString string) {
	ConnString = newConnString
}

// TODO: Refactor this shit to be more generic

func queryForRows(query string, db *sqlite3.Conn, args ...interface{}) (*sqlite3.Stmt, error) {

	stmt, err := db.Prepare(query, args...)
	if err != nil {
		stmt.Close()
		log.Println("Error querying: ", err)
		return nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			stmt.Close()
			fmt.Println("Recovered in query", r)
		}
	}()
	return stmt, err
}
