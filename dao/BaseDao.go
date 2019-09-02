package dao

import (
	"database/sql"
	"fmt"
)

func queryForRows(query string, db *sql.DB) (*sql.Rows, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in query", r)
		}
	}()

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	return rows, err
}

func queryForRowsWithParams(statement *sql.Stmt, db *sql.DB, args ...interface{}) (*sql.Rows, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in query", r)
		}
	}()

	rows, err := statement.Query(args...)
	if err != nil {
		return nil, err
	}
	return rows, err
}

func executeQueryWithParams(statement *sql.Stmt, db *sql.DB, args ...interface{}) sql.Result {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in query with params", r)
		}
	}()

	result, err := statement.Exec(args...)
	if err != nil {
		return nil
	}

	if rowsAffected, err := result.RowsAffected(); err != nil && rowsAffected > 0 {
		log.Error("Error occured executing query.", err)
		return nil
	}
	return result
}
