package tcp

import (
	"chat/logger"
	"chat/pg"
	"chat/shared"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func SessionsCreate(db *sql.DB, user_id string, expires time.Time, session_token string, created time.Time) (err error) {
	sqlStatement := `
	INSERT INTO tcp_sessions (user_id, expires, session_token, created)
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
	sqlStatement := `SELECT * FROM tcp_sessions WHERE session_token=$1;`
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
	DELETE FROM tcp_sessions
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

func New_Manager(dbconn *sql.DB) (mgr *SessionsManager) {
	mgr = &SessionsManager{
		DBconn: dbconn,
	}
	return
}

type SessionsManager struct {
	DBconn *sql.DB
}

func Validate(w http.ResponseWriter, r *http.Request) (session shared.HTTP_Session, err error) { // CHANGE TO ACCEPT ONLY THE COOKIE 'c.Cookie'
	c, err := r.Cookie("session_token")
	// Check browser cookies
	if err != nil {
		if err == http.ErrNoCookie {
			logger.LogColor("SESSION", "No Cookie")
			return session, errors.New("session_no_cookie")
		}
		logger.LogColor("SESSION", "General error")
		return session, errors.New("session_err")
	}
	session_token := c.Value
	// If found check against DB
	userSession, err := pg.SessionsRead(active_tcp.DBconn, session_token)
	// if session_token != userSession.Token {
	// 	logger.LogColor("SESSION", "Invalid session provided")
	// 	return session, errors.New("invalid session")
	// }
	session = shared.HTTP_Session{
		Token:   userSession.Token,
		UserID:  userSession.UserID,
		Created: userSession.Created,
		Expiry:  userSession.Expiry,
	}
	if err != nil {
		logger.LogColor("SESSION", "shared.HTTP_Session doesn't exist")
		return session, errors.New("session_not_exist")
	}
	if session.Is_Expired() {
		pg.SessionsDelete(active_tcp.DBconn, session_token)
		http.SetCookie(w, &http.Cookie{ // ! MA NO PORCO DIO, DELETECOOKIE NON SETCOOKIE
			Name:    "session_token",
			Value:   "",
			Expires: time.Now(),
		})
		logger.LogColor("SESSION", "shared.HTTP_Session expired, removed from client")
		return session, errors.New("session_expired")
	}
	return session, nil
}

func GenerateUUID() uuid.UUID {
	id := uuid.New()
	return id
}

func GenerateSaltedUUID(password string) (uuid string) { // ? Should username be salt aswell?
	// 1. salt password with SALT
	// Store Salt in DB
	// 2. SHA512 HASH
	// 3. BASE64 ENC
	// Generate random array of 32 bytes (256 bits)
	saltSize := 32
	bRand := make([]byte, saltSize)
	_, err := rand.Read(bRand[:])
	if err != nil {
		panic(err)
	}
	// Salt password
	bPassword := []byte(password)
	// Hash SHA512
	sha512Hasher := sha512.New()
	bPassword = append(bPassword, bRand...)
	sha512Hasher.Write(bPassword)
	sha_sum := sha512Hasher.Sum(nil)
	// Encode to base64
	uuid = base64.StdEncoding.EncodeToString(sha_sum)
	return
}

func MatchHashAndSalt(usr_uuid string, password string, rand_salt string) error {
	bRand := []byte(rand_salt)
	bPassword := []byte(password)
	sha512Hasher := sha512.New()
	bPassword = append(bPassword, bRand...)
	sha512Hasher.Write(bPassword)
	sha_sum := sha512Hasher.Sum(nil)
	// Encode to base64
	gen_uuid := base64.StdEncoding.EncodeToString(sha_sum)
	if gen_uuid != usr_uuid {
		logger.LogColor("SESSION", "Error matching salted password.")
		return errors.New("err_salt_pass")
	}
	return nil
}

// Returns user privileged session
func New_Session(session_token string, user_id string) (new_session *shared.HTTP_Session) {
	createdAt := time.Now()
	expiresAt := time.Now().Add(720 * time.Hour)
	new_session = &shared.HTTP_Session{
		Token:   session_token,
		UserID:  user_id,
		Created: createdAt,
		Expiry:  expiresAt,
	}
	return
}
