// !!!!!!
// !!!!!!
// !!!!!! ALL THIS MUST BECOME AN INTERFACE for the db isntance
// !!!!!!
// !!!!!!

package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type postgreConfig struct {
	IP       string
	Port     string
	Username string
	Password string
	DB       string
	SSLMode  string
}

func connector(config postgreConfig) (db *sql.DB) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", config.IP, config.Port, config.Username, config.Password, config.DB, config.SSLMode)
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
