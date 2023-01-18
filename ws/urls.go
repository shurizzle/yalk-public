package ws

import (
	"chat/logger"
	"chat/pg"
	"chat/shared"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func p_Root(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	logger.LogColor("HTTPS", fmt.Sprintf("Index requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	_, err := Validate(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/chat", http.StatusFound)
}

func p_Login(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("HTTPS", fmt.Sprintf("Login requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	// event := "login"
	defer r.Body.Close()

	//TODO: Change the error query with MessageEvent: https://developer.mozilla.org/en-US/docs/Web/API/MessageEvent

	error_code := r.URL.Query().Get("error") // Error_Code stores the error code from URL Query, err in this case is a clean case so just display the context
	var payload shared.Payload

	// if error_code == "true" {
	// 	payload = shared.Payload{
	// 		Success: true,
	// 		Event:   "error_login",
	// 	}
	// } else {
	// 	payload = shared.Payload{
	// 		Success: true,
	// 		Event:   "",
	// 	}
	// }
	_, err := Validate(w, r)
	if err == nil {
		http.Redirect(w, r, "/chat", http.StatusFound)
		return
	}

	if r.Method == http.MethodPost && error_code == "" {
		var err error
		r.ParseForm()
		if err != nil {
			panic(err)
		}
		login_creds := shared.Credentials{
			Username: r.PostForm.Get("username"),
			Password: r.PostForm.Get("password"),
		}
		db_creds := pg.Login(active_ws.DBconn, login_creds.Username)

		if err != nil {
			log.Print(err)
		}
		err = bcrypt.CompareHashAndPassword([]byte(db_creds.Password), []byte(login_creds.Password))
		if err != nil {
			http.Redirect(w, r, "/login?error=true", http.StatusFound)
			return
		} else {
			// Salting with password
			session_token := GenerateSaltedUUID(login_creds.Password)
			new_session := New_Session(session_token, db_creds.ID)
			// Setup and Admin Rights check
			user_profile, err := pg.UserRead(active_ws.DBconn, db_creds.ID, true)
			if err != nil {
				// ? Here we can implement an "Account setup" as the user_profile is still not created
				logger.LogColor("HTTPS", "Cannot find user")
			}
			// TODO: Move in separate function in http_session_manager.go
			// Create and store session
			pg.SessionsCreate(active_ws.DBconn, new_session.UserID, new_session.Expiry, new_session.Token, new_session.Created)
			if err != nil {
				logger.LogColor("HTTPS", "Could not create the sessions")
				return
			}
			// Give to client the cookie for "session_token" and expiry 120s
			http.SetCookie(w, &http.Cookie{
				Name:    "session_token",
				Value:   new_session.Token,
				Expires: new_session.Expiry,
			})
			if user_profile.Status != "online" {
				pg.UserStatusUpdate(active_ws.DBconn, user_profile.ID, "online", false)

			}
			http.Redirect(w, r, "/chat", http.StatusFound) // Chat if login successful
		}
	}

	base_page := filepath.Join("static", "login.html")
	temp := template.Must(template.New("login.html").ParseFiles(base_page))

	err = temp.Execute(w, payload)
	if err != nil {
		panic(err)
	}
}

func p_Chat(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("HTTPS", fmt.Sprintf("Chat requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	defer r.Body.Close()
	http_session, err := Validate(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	_, err = pg.UserRead(active_ws.DBconn, http_session.UserID, true)
	if err != nil {
		logger.LogColor("HTTPS", "User not found, general error.")
	}

	channel_id := r.URL.Query().Get("id") // Error_Code stores the error code from URL Query, err in this case is a clean case so just display the context
	if channel_id == "" {
		channel_id = "main"
	}

	http_response(w, true, "base.html", nil)
}

func p_Profile(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("HTTPS", fmt.Sprintf("Profile requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	defer r.Body.Close()
	event := "get_profile"
	http_session, err := Validate(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	user_profile, err := pg.UserRead(active_ws.DBconn, http_session.UserID, true)
	if err != nil {
		logger.LogColor("HTTPS", "User not found, general error.")
	}
	all_users := pg.UserReadAll(active_ws.DBconn)

	profile_page_data := map[string]any{"logged_user": user_profile, "server_users": all_users}
	payload := shared.Payload{
		Success: true,
		Origin:  http_session.UserID,
		Event:   event,
		Data:    profile_page_data,
	}
	http_response(w, true, "profile.html", payload)
}

func p_Logout(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("HTTPS", fmt.Sprintf("Logout requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	// context := "logout" // * Unused for now
	defer r.Body.Close()
	http_session, err := Validate(w, r)
	if err == nil {
		logger.LogColor("HTTPS", "Removing session from active_ws.dbconn")
		pg.SessionsDelete(active_ws.DBconn, http_session.Token)
	} else {
		logger.LogColor("HTTPS", "No session found")
	}
	// We need to let the client know that the cookie is expired
	// In the response, we set the session token to an empty
	// value and set its expiry as the current time
	http.SetCookie(w, &http.Cookie{ // ? Session manager?
		Name:    "session_token",
		Value:   "",
		Expires: time.Now(),
	})

	http.Redirect(w, r, "/login", http.StatusFound)
}

// func p_Pinger(w http.ResponseWriter, r *http.Request) { //! Server Core!
// 	// ! MUST INTRODUCE CHECK IF THE USER IS THE SAME AS THE ONE TIED TO SESSION TOKEN
// 	logger.LogColor("HTTPS", fmt.Sprintf("Ping received from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
// 	// context := "logout" // * Unused for now
// 	defer r.Body.Close()

// 	if r.Method != http.MethodPost {
// 		return
// 	}
// 	_, err := Validate(w, r)
// 	if err != nil {
// 		// http.Redirect(w, r, "/login", http.StatusFound)
// 		return
// 	}
// 	// w.WriteHeader(http.StatusOK)
// 	http_response(w, false, nil)
// }
