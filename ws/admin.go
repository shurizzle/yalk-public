package ws

import (
	"chat/logger"
	"chat/pg"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func Admin_User_Add(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("HTTPS", fmt.Sprintf("Add User requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	// page := "logout" // * Unused for now
	defer r.Body.Close()
	_, err := Validate(w, r) // TODO: Add active_ws.API.DBconn "admin command executed by" table
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	r.ParseForm()

	username := r.PostForm.Get("username-new")
	password := r.PostForm.Get("password-new")
	is_admin := r.PostForm.Get("admin")

	enc_password, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		logger.LogColor("HTTPS", "Error in encrpyting password for new user")
		return
	}
	new_user_id := pg.UserCreate(active_ws.API.DBconn, username, string(enc_password))
	pg.ProfileCreate(active_ws.API.DBconn, new_user_id, "testDisplayName", "yellow", is_admin)
	if err := os.Mkdir(fmt.Sprintf("static/data/user_avatars/%s", new_user_id), os.ModePerm); err != nil {
		log.Fatal(err)
	}
	if err := os.Link("static/data/user_avatars/default/avatar.png", fmt.Sprintf("static/data/user_avatars/%s/avatar.png", new_user_id)); err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(new_user_id))
}

func Admin_User_Delete(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("HTTPS", fmt.Sprintf("User DELETE requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	// page := "delete" // * Unused for now
	defer r.Body.Close()
	if r.Method != http.MethodDelete {
		return
	}
	_, err := Validate(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	id := r.URL.Query().Get("id")

	pg.UserDelete(active_ws.API.DBconn, id)
	pg.ProfileDelete(active_ws.API.DBconn, id)

	w.WriteHeader(http.StatusAccepted)
}
func Admin_User_Update(w http.ResponseWriter, r *http.Request) {
	// ! MUST INTRODUCE CHECK IF THE USER IS THE SAME AS THE ONE TIED TO SESSION TOKEN
	logger.LogColor("HTTPS", fmt.Sprintf("User Update requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	// page := "logout" // * Unused for now
	defer r.Body.Close()
	http_session, err := Validate(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	user_profile, err := pg.UserRead(active_ws.API.DBconn, http_session.UserID, true)
	if err != nil {
		logger.LogColor("HTTPS", "User not found, general error.")
		return
	}
	if user_profile.IsAdmin != "true" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	r.ParseForm()
	_user_id := r.PostForm.Get("user_id")
	display_name := r.PostForm.Get("display_name")
	is_admin := r.PostForm.Get("is_admin")
	color := r.PostForm.Get("color")

	user_id := _user_id
	if err != nil {
		logger.LogColor("HTTPS", "Can't convert user ID to int")
		return
	}

	pg.ProfileUpdate(active_ws.API.DBconn, user_id, display_name, color, is_admin)
	// buildSSEFrame()
	w.WriteHeader(http.StatusAccepted)

}
