package bot

import (
	"database/sql"

	_ "github.com/kaaori/MhBotGo/dao"
)

// DB : The core database functions of the bot
type DB struct {
	db *sql.DB
}

// InitDB : Init an MH db instance
func initDB(dbFilePath string, driver string) (*DB, error) {
	return dbLoad(dbFilePath, driver)
}

func dbLoad(dbFilePath string, driver string) (*DB, error) {
	conn, err := sql.Open(driver, dbFilePath)
	botDB := DB{db: conn}
	if err != nil {
		return &botDB, err
	}

	err = botDB.db.Ping()
	if err != nil {
		return &botDB, err
	}
	return &botDB, err
}
