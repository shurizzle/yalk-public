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
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type postgreConfig struct {
	IP       string
	Port     string
	Username string
	Password string
	DB       string
	SSLMode  string
}

func connector(config postgreConfig) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", config.IP, config.Port, config.Username, config.Password, config.DB, config.SSLMode)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, err
}

func dbLogin(db *sql.DB, username string) (creds credentials) {
	// var id uint8
	sqlStatement := `SELECT id, username, password FROM public.login_users WHERE username=$1;`
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

func initDb(config postgreConfig) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=%s", config.IP, config.Port, config.Username, config.Password, config.SSLMode)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	// TODO: Put "ping" everywhere before executing SQL statements
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("CREATE DATABASE db_chat")
	if err != nil {
		logger.LogColor("DATABASE", "Error UPDATING profile")
		return nil, err
	}
	db.Close()

	psqlInfo = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", config.IP, config.Port, config.Username, config.Password, config.DB, config.SSLMode)
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	sqlFile, err := os.ReadFile("db_chat.sql")
	if err != nil {
		logger.LogColor("DATABASE", "Cannot read 'db_chat.sql'")
		return nil, err
	}
	_, err = db.Exec(string(sqlFile))
	if err != nil {
		logger.LogColor("DATABASE", "Error executing 'db_chat.sql'")
		return nil, err
	}
	logger.LogColor("DATABASE", "Executed 'db_chat.sql' succesfully.")

	settingsSql := `INSERT INTO public.server_settings (key, value)
					VALUES ('default_channel', 'CHATMAIN'), 
							('server_id', '1');`
	_, err = db.Exec(settingsSql)
	if err != nil {
		logger.LogColor("DATABASE", "Error creating server settings")
	}

	chatmainSql := `INSERT INTO public.server_chats (
		id, type, users, name, created_by, created_date)
		VALUES ('CHATMAIN', 'channel_public', '{}' ,'General', '0', current_timestamp);`
	_, err = db.Exec(chatmainSql)
	if err != nil {
		logger.LogColor("DATABASE", "Error creating CHATMAIN")
	}
	chatrandSql := `INSERT INTO public.server_chats (
			id, type, users, name, created_by, created_date)
			VALUES ('CHATRAND', 'channel_public', '{}' ,'Random', '0', current_timestamp);`
	_, err = db.Exec(chatrandSql)
	if err != nil {
		logger.LogColor("DATABASE", "Error creating CHATRAND")
	}

	defaultPwd, err := bcrypt.GenerateFromPassword([]byte("admin"), 14)
	if err != nil {
		logger.LogColor("DATABASE", "Can't hash admin default password")
		return nil, err
	}
	id := userCreate(db, "1", "admin", string(defaultPwd))
	logger.LogColor("DATABASE", fmt.Sprintf("Created 'admin' with id %v", id))
	// chatCreate(db, "0", "channel_public", "General")

	_, err = chatJoin(db, "1", "CHATMAIN")
	if err != nil {
		logger.LogColor("DATABASE", "Can't join user in CHATMAIN")
		return nil, err
	}
	_, err = chatJoin(db, "1", "CHATRAND")
	if err != nil {
		logger.LogColor("DATABASE", "Can't join user in CHATRAND")
		return nil, err
	}
	logger.LogColor("DATABASE", fmt.Sprintf("Updated id %v", id))

	// logger.LogColor("DATABASE", fmt.Sprintf("Updated id %v", id))
	logger.LogColor("DATABASE", "Succesfully created DB")
	return db, nil

}
