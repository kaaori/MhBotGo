package dao

import (
	"fmt"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
)

// TODO: Refactor this shit to be more generic

func queryForRows(query string, db *sqlite3.Conn, args ...interface{}) (*sqlite3.Stmt, error) {

	stmt, err := db.Prepare(query, args...)
	if err != nil {
		stmt.Close()
		log.Error("Error querying: ", err)
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
