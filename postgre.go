// !!!!!!
// !!!!!!
// !!!!!! ALL THIS MUST BECOME AN INTERFACE for the db isntance
// !!!!!!
// !!!!!!

package main

import (
	"chat/logger"
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

func dbLogin(db *sql.DB, username string) (creds credentials) {
	// var id uint8
	sqlStatement := `SELECT id, username, password FROM login_users WHERE username=$1;`
	var id string
	var hash string

	row := db.QueryRow(sqlStatement, username)
	// Here means: it assigns err with the row.Scan()
	// then "; err" means use "err" in the "switch" statement
	switch err := row.Scan(&id, &username, &hash); err {
	case sql.ErrNoRows:
		logger.LogColor("DATABASE", "No rows were returned!")
		return
	case nil:
		creds := credentials{ //! WRONG!
			ID:       id,
			Username: username,
			Password: hash,
		}
		return creds
	default:
		return
	}
}
