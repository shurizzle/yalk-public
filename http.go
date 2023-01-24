package main

import (
	"chat/logger"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type httpServer struct {
	config configNetwork
	db     *sql.DB
}

func startHTTPServer(netConf configNetwork, dbConf postgreConfig) (*httpServer, error) {
	logger.LogColor("WEBSRV", "Starting HTTP and HTTPS listeners..")

	httpServer := &httpServer{
		config: netConf,
		db:     connector(dbConf),
	}

	fs := http.FileServer(http.Dir("./static"))

	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/favicon.ico", favicon)

	http.HandleFunc("/", rootPage)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/logout", logoutPage)
	http.HandleFunc("/chat", chatPage)
	http.HandleFunc("/profile", profilePage)

	httpAddr := httpServer.config.IP + ":" + string(httpServer.config.Port)
	httpsAddr := httpServer.config.IP + ":" + string(httpServer.config.PortTLS)
	go func() {
		logger.LogColor("WEBSRV", "HTTP listener started")
		err := http.ListenAndServe(httpAddr, http.HandlerFunc(redirectToTLS))
		if err != nil {
			// panic(fmt.Sprintf("Error listening HTTP: %v", err))
			panic(err)
		}
	}()

	go func() {
		logger.LogColor("WEBSRV", "HTTPS listener started")
		err := http.ListenAndServeTLS(httpsAddr, "localhost.crt", "localhost.key", nil)
		if err != nil {
			// panic(fmt.Sprintf("Error listening HTTPS: %v", err))
			panic(err)
		}
	}()

	logger.LogColor("WEBSRV", "Loaded succesfully.")
	return httpServer, nil
}

func favicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/images/favicon.ico")
}
func redirectToTLS(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("HTTPS", "Redirecting HTTP requests to HTTPS")
	http.Redirect(w, r, ":443", http.StatusSeeOther)
}

func handleClientData(w http.ResponseWriter, renderTemplate bool, _fileName string, _payload any) {
	if !renderTemplate {
		switch _payload.(type) {
		case dataPayload:
			payload, err := json.Marshal(_payload)
			if err != nil {
				logger.LogColor("HTTPS", "Marshaling error")
				w.WriteHeader(http.StatusInternalServerError) // ? Which is best to write http.responses? Write or WriteHeader?
			}
			_, err = w.Write(payload)
			if err != nil {
				logger.LogColor("HTTPS", "Error writing response")
			}
			w.WriteHeader(http.StatusOK)
			return

		default:
			w.WriteHeader(http.StatusOK)
			return
		}
	} else {
		webapp := filepath.Join("static", _fileName)
		temp := template.Must(template.New(_fileName).ParseFiles(webapp))

		switch _payload.(type) {
		case dataPayload:
			payload, err := json.Marshal(_payload)
			if err != nil {
				logger.LogColor("HTTPS", "Marshaling error")
				w.WriteHeader(http.StatusInternalServerError) // ? Which is best to write http.responses? Write or WriteHeader?
			}
			err = temp.Execute(w, payload)
			if err != nil {
				panic(err)
			}
			w.WriteHeader(http.StatusOK)
			return

		default:
			err := temp.Execute(w, nil)
			if err != nil {
				panic(err)
			}
			w.WriteHeader(http.StatusOK)
		}
	}
}

func rootPage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	logger.LogColor("HTTPS", fmt.Sprintf("Index requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	_, err := sessionValidate(w, r, activeServer.dbconn)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/chat", http.StatusFound)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("HTTPS", fmt.Sprintf("Login requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	// event := "login"
	defer r.Body.Close()

	//TODO: Change the error query with MessageEvent: https://developer.mozilla.org/en-US/docs/Web/API/MessageEvent

	// error_code := r.URL.Query().Get("error") // Error_Code stores the error code from URL Query, err in this case is a clean case so just display the context
	var payload dataPayload

	// if error_code == "true" {
	// 	payload = Payload{
	// 		Success: true,
	// 		Event:   "error_login",
	// 	}
	// } else {
	// 	payload = Payload{
	// 		Success: true,
	// 		Event:   "",
	// 	}
	// }
	_, err := sessionValidate(w, r, activeServer.dbconn)
	if err == nil {
		http.Redirect(w, r, "/chat", http.StatusFound)
		return
	}

	if r.Method == http.MethodPost { //&& error_code == "" {
		var err error
		err = r.ParseForm()
		if err != nil {
			panic(err)
		}
		loginCreds := credentials{
			Username: r.PostForm.Get("username"),
			Password: r.PostForm.Get("password"),
		}
		dbCreds := dbLogin(activeServer.dbconn, loginCreds.Username)

		if err != nil {
			log.Print(err)
		}
		err = bcrypt.CompareHashAndPassword([]byte(dbCreds.Password), []byte(loginCreds.Password))
		if err != nil {
			http.Redirect(w, r, "/login?error=true", http.StatusFound)
			return
		}
		// Salting with password
		token := newUUIDSalted(loginCreds.Password)
		session := sessionCreate(token, dbCreds.ID)
		// Setup and Admin Rights check
		userProfile, err := userRead(activeServer.dbconn, dbCreds.ID, true)
		if err != nil {
			// ? Here we can implement an "Account setup" as the userProfile is still not created
			logger.LogColor("HTTPS", "Cannot find user")
		}
		// TODO: Move in separate function in session_manager.go
		// Create and store session
		err = sessionStore(activeServer.dbconn, session.UserID, session.Expiry, session.Token, session.Created)
		if err != nil {
			logger.LogColor("HTTPS", "Could not create the sessions")
			return
		}
		// Give to client the cookie for "session_token" and expiry 120s
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   session.Token,
			Expires: session.Expiry,
		})
		if userProfile.Status != "online" {
			err := userOnlineUpdate(activeServer.dbconn, userProfile.ID, false)
			if err != nil {
				logger.LogColor("HTTPS", "Cannot change user status")
				return
			}
		}
		http.Redirect(w, r, "/chat", http.StatusFound) // Chat if login successful

	}

	index := filepath.Join("static", "login.html")
	temp := template.Must(template.New("login.html").ParseFiles(index))

	err = temp.Execute(w, payload)
	if err != nil {
		panic(err)
	}
}

func chatPage(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("HTTPS", fmt.Sprintf("Chat requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	defer r.Body.Close()
	session, err := sessionValidate(w, r, activeServer.dbconn)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	_, err = userRead(activeServer.dbconn, session.UserID, true)
	if err != nil {
		logger.LogColor("HTTPS", "User not found, general error.")
	}

	// var channel_id string
	// Error_Code stores the error code from URL Query, err in this case is a clean case so just display the context

	// if r.URL.Query().Get("id") == "" {
	// 	channel_id = r.URL.Query().Get("id")
	// }

	handleClientData(w, true, "base.html", nil)
}

func profilePage(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("HTTPS", fmt.Sprintf("Profile requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	defer r.Body.Close()
	event := "get_profile"
	session, err := sessionValidate(w, r, activeServer.dbconn)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	userProfile, err := userRead(activeServer.dbconn, session.UserID, true)
	if err != nil {
		logger.LogColor("HTTPS", "User not found, general error.")
	}
	users := userReadAll(activeServer.dbconn)

	data := map[string]any{"logged_user": userProfile, "server_users": users}
	payload := dataPayload{
		Success: true,
		Origin:  session.UserID,
		Event:   event,
		Data:    data,
	}
	handleClientData(w, true, "profile.html", payload)
}

func logoutPage(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("HTTPS", fmt.Sprintf("Logout requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	// context := "logout" // * Unused for now
	defer r.Body.Close()
	session, err := sessionValidate(w, r, activeServer.dbconn)
	if err == nil {
		logger.LogColor("HTTPS", "Removing session from active_ws.dbconn")
		sessionDelete(activeServer.dbconn, session.Token)
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
