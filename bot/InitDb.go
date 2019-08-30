package bot

import (
	"io/ioutil"

	logging "github.com/kaaori/mhbotgo/log"

	// TODO
	"database/sql"

	// TODO
	_ "github.com/mattn/go-sqlite3"
)

var (
	log = logging.NewLog()
)

// ReadDML : Reads the DML from the predefined sql script
func ReadDML(dbLocation string) {
	buf, err := ioutil.ReadFile("./data/MHBot-schemata.sql")
	if err != nil {
		log.Error("Error installing table schemata -> ", err)
		return
	}
	log.Info("DML Loaded, creating tables")
	installDML(string(buf), dbLocation)
}

func installDML(dml string, dbLocation string) {
	db, err := sql.Open("sqlite3", dbLocation)
	isError(err)
	defer db.Close()

	// = used over := when assigning to "existing" vars only (err is assigned upon opening a connection)
	// _ used to ignore the first return value of the Exec function, as we don't need it
	_, err = db.Exec(dml)
	isError(err)

	log.Info("Tables created")
}

func isError(err error) {
	if err != nil {
		// log.Error("Error in SQL setup -> ", err)
		panic(err)
	}
}
