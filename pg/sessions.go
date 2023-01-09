package pg

import (
	"chat/logger"
	"chat/shared"
	"database/sql"
	"fmt"
	"time"
)

func SessionsCreate(db *sql.DB, user_id string, expires time.Time, session_token string, created time.Time) (err error) {
	sqlStatement := `
	INSERT INTO http_sessions (user_id, expires, session_token, created)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (user_id) DO UPDATE
		SET user_id = excluded.user_id,
			expires = excluded.expires,
			session_token = excluded.session_token,
			created = excluded.created;`
	db.QueryRow(sqlStatement, user_id, expires, session_token, created)
	logger.LogColor("DATABASE", "Saved new session")
	return
}

func SessionsRead(db *sql.DB, session_token string) (session shared.HTTP_Session, err error) {
	sqlStatement := `SELECT * FROM http_sessions WHERE session_token=$1;`
	var user_id string
	var expires time.Time
	var db_session_token string
	var created time.Time
	// var ip_address string

	row := db.QueryRow(sqlStatement, session_token).Scan(&user_id, &expires, &db_session_token, &created)
	// Here means: it assigns err with the row.Scan()
	// then "; err" means use "err" in the "switch" statement
	switch row {
	case sql.ErrNoRows:
		logger.LogColor("DATABASE", "No SESSIONS were returned")
		return session, err
	case nil:
		session := shared.HTTP_Session{
			UserID:  user_id,
			Token:   session_token,
			Created: created,
			Expiry:  expires,
		}
		return session, nil

	default:
		logger.LogColor("DATABASE", "Error in postgre.SessionsReads")
		return session, err
	}
}

func SessionsDelete(db *sql.DB, session_token string) {
	sqlStatement := `
	DELETE FROM http_sessions
	WHERE session_token = $1;`
	res, err := db.Exec(sqlStatement, session_token)
	if err != nil {
		logger.LogColor("DATABASE", "Error in postgre_main.DeleteSession")
	}
	_, err = res.RowsAffected()
	if err != nil {
		logger.LogColor("DATABASE", "No SESSIONS were deleted")
		return
	}
	logger.LogColor("DATABASE", fmt.Sprintf("Deleted session token %s", session_token))
}
