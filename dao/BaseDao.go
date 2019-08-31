package dao

import "database/sql"

func queryForRows(query string, db *sql.DB) (*sql.Rows, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	return rows, err
}

func queryForRowsWithParams(statement *sql.Stmt, db *sql.DB, args ...interface{}) (*sql.Rows, error) {
	rows, err := statement.Query(args...)
	if err != nil {
		return nil, err
	}
	return rows, err
}

func executeQueryWithParams(statement *sql.Stmt, db *sql.DB, args ...interface{}) sql.Result {
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
