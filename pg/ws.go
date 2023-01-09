package pg

import (
	"chat/logger"
	"chat/shared"
	"database/sql"
)

func Login(db *sql.DB, username string) (db_creds shared.Credentials) {
	// var id uint8
	sqlStatement := `SELECT id, username, password FROM login_users WHERE username=$1;`
	var id string
	var p_hash string

	row := db.QueryRow(sqlStatement, username)
	// Here means: it assigns err with the row.Scan()
	// then "; err" means use "err" in the "switch" statement
	switch err := row.Scan(&id, &username, &p_hash); err {
	case sql.ErrNoRows:
		logger.LogColor("DATABASE", "No rows were returned!")
		return
	case nil:
		db_creds := shared.Credentials{ //! WRONG!
			ID:       id,
			Username: username,
			Password: p_hash,
		}
		return db_creds
	default:
		panic(err)
	}
}
