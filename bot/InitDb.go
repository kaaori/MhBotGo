package bot

import (
	"io/ioutil"
	"strconv"

	"mhbotgo.com/dao"

	// TODO
	"log"

	// TODO
	_ "github.com/bvinc/go-sqlite-lite/sqlite3"
)

// ReadDML : Reads the DML from the predefined sql script
func ReadDML(dbLocation string) {
	buf, err := ioutil.ReadFile("./data/MHBot-schemata.sql")
	if err != nil {
		log.Println("Error installing table schemata -> ", err)
		return
	}
	log.Println("DML Loaded, creating tables")
	installDML(string(buf), dbLocation)
}

func installDML(dml string, dbLocation string) {
	DB := dao.GetConnection(dao.ConnString)
	defer DB.Close()

	if DB != nil {
		err := DB.Exec(dml)
		log.Println(strconv.Itoa(DB.TotalChanges()) + " Changes")
		isError(err)

		log.Println("Tables created")
	} else {
		log.Println("DB Is null")
	}

}

func isError(err error) {
	if err != nil {
		// log.Println("Error in SQL setup -> ", err)
		panic(err)
	}
}
