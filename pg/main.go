// !!!!!!
// !!!!!!
// !!!!!! ALL THIS MUST BECOME AN INTERFACE for the db isntance
// !!!!!!
// !!!!!!

package pg

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Postgre_Client struct {
	IP       string
	Port     string
	Username string
	Password string
	DB       string
	SSLMode  string
}

func NewPostgreConn(dbConf Postgre_Client) (db *sql.DB) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", dbConf.IP, dbConf.Port, dbConf.Username, dbConf.Password, dbConf.DB, dbConf.SSLMode)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return db
}
