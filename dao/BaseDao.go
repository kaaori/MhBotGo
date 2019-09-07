package dao

import (
	"fmt"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
)

func queryForRows(query string, db *sqlite3.Conn, args ...interface{}) (*sqlite3.Stmt, error) {

	stmt, err := db.Prepare(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in query", r)
		}
	}()
	return stmt, err
}
