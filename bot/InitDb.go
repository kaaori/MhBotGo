package bot

import (
	"io/ioutil"
	"strconv"

	"github.com/kaaori/MhBotGo/dao"
	logging "github.com/kaaori/mhbotgo/log"

	// TODO

	// TODO
	_ "github.com/bvinc/go-sqlite-lite/sqlite3"
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
	DB := dao.GetConnection(dao.ConnString)
	defer DB.Close()

	if DB != nil {
		err := DB.Exec(dml)
		log.Info(strconv.Itoa(DB.TotalChanges()) + " Changes")
		isError(err)

		log.Info("Tables created")
	} else {
		log.Error("DB Is null")
	}

}

func isError(err error) {
	if err != nil {
		// log.Error("Error in SQL setup -> ", err)
		panic(err)
	}
}
