package main

import (
	"chat/logger"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type credentials struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type httpSession struct {
	Token   string
	UserID  string
	IP      string
	Created time.Time
	Expiry  time.Time
}

func (s httpSession) isExpired() bool {
	return s.Expiry.Before(time.Now())
}

func sessionValidate(w http.ResponseWriter, r *http.Request, db *sql.DB) (session httpSession, err error) { // CHANGE TO ACCEPT ONLY THE COOKIE 'c.Cookie'
	ua := r.UserAgent()
	log.Printf("User Agent: %v", ua)
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
	token := c.Value
	// If found check against DB
	userSession, err := sessionRead(db, token)
	// if session_token != userSession.Token {
	// 	logger.LogColor("SESSION", "Invalid session provided")
	// 	return session, errors.New("invalid session")
	// }
	session = httpSession{
		Token:   userSession.Token,
		UserID:  userSession.UserID,
		Created: userSession.Created,
		Expiry:  userSession.Expiry,
	}
	if err != nil {
		logger.LogColor("SESSION", "HTTPSession doesn't exist")
		return session, errors.New("session_not_exist")
	}
	if session.isExpired() {
		sessionDelete(db, token)
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   "",
			Expires: time.Now(),
		})
		logger.LogColor("SESSION", "HTTPSession expired, removed from client")
		return session, errors.New("session_expired")
	}
	return session, nil
}

func newUUID() uuid.UUID {
	id := uuid.New()
	return id
}

func newUUIDSalted(password string) (uuid string) { // ? Should username be salt aswell?
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
	sum := sha512Hasher.Sum(nil)
	// Encode to base64
	uuid = base64.StdEncoding.EncodeToString(sum)
	return
}

func matchHashAndSalt(uuid string, password string, salt string) error {
	bRand := []byte(salt)
	bPassword := []byte(password)
	sha512Hasher := sha512.New()
	bPassword = append(bPassword, bRand...)
	sha512Hasher.Write(bPassword)
	sum := sha512Hasher.Sum(nil)
	// Encode to base64
	matchUUID := base64.StdEncoding.EncodeToString(sum)
	if matchUUID != uuid {
		logger.LogColor("SESSION", "Error matching salted password.")
		return errors.New("err_salt_pass")
	}
	return nil
}

// Returns user privileged session
func sessionCreate(token string, id string) (session *httpSession) {
	createdAt := time.Now()
	expiresAt := time.Now().Add(720 * time.Hour)
	session = &httpSession{
		Token:   token,
		UserID:  id,
		Created: createdAt,
		Expiry:  expiresAt,
	}
	return
}

func sessionStore(db *sql.DB, id string, expires time.Time, token string, created time.Time) (err error) {
	sqlStatement := `
	INSERT INTO public.http_sessions (user_id, expires, session_token, created)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (user_id) DO UPDATE
		SET user_id = excluded.user_id,
			expires = excluded.expires,
			session_token = excluded.session_token,
			created = excluded.created;`
	db.QueryRow(sqlStatement, id, expires, token, created)
	logger.LogColor("DATABASE", "Saved new session")
	return
}

func sessionRead(db *sql.DB, token string) (session httpSession, err error) {
	sqlStatement := `SELECT * FROM public.http_sessions WHERE session_token=$1;`
	var id string
	var expires time.Time
	var matchToken string
	var created time.Time

	row := db.QueryRow(sqlStatement, token).Scan(&id, &expires, &matchToken, &created)
	// Here means: it assigns err with the row.Scan()
	// then "; err" means use "err" in the "switch" statement
	switch row {
	case sql.ErrNoRows:
		logger.LogColor("DATABASE", "No SESSIONS were returned")
		return session, err
	case nil:
		session := httpSession{
			UserID:  id,
			Token:   token,
			Created: created,
			Expiry:  expires,
		}
		return session, nil

	default:
		logger.LogColor("DATABASE", "Error in postgre.SessionsReads")
		return session, err
	}
}

func sessionDelete(db *sql.DB, token string) {
	sqlStatement := `
	DELETE FROM public.http_sessions
	WHERE session_token = $1;`
	res, err := db.Exec(sqlStatement, token)
	if err != nil {
		logger.LogColor("DATABASE", "Error in postgre_main.DeleteSession")
	}
	_, err = res.RowsAffected()
	if err != nil {
		logger.LogColor("DATABASE", "No SESSIONS were deleted")
		return
	}
	logger.LogColor("DATABASE", fmt.Sprintf("Deleted session token %s", token))
}
