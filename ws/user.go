package ws

import (
	"chat/logger"
	"chat/pg"
	"chat/shared"
	"fmt"
	"io"
	"net/http"
	"os"
)

func UserRead(w http.ResponseWriter, r *http.Request) {
	// ! MUST INTRODUCE CHECK IF THE USER IS THE SAME AS THE ONE TIED TO SESSION TOKEN
	logger.LogColor("HTTPS", fmt.Sprintf("User read requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	// page := "logout" // * Unused for now
	defer r.Body.Close()
	http_session, err := Validate(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var user_profile shared.User_Profile

	user_id_lookup := r.URL.Query().Get("id")

	usrIsSelf := func(l string, s string) bool {
		if l != s {
			return true
		} else {
			return false
		}
	}(user_id_lookup, http_session.UserID)

	if user_id_lookup != "" {
		user_profile, err = pg.UserRead(active_ws.API.DBconn, user_id_lookup, usrIsSelf)
		if err != nil {
			logger.LogColor("HTTPS", "User not found, general error.")
			return
		}
	} else {
		user_profile, err = pg.UserRead(active_ws.API.DBconn, http_session.UserID, true)
		if err != nil {
			logger.LogColor("HTTPS", "User not found, general error.")
			return
		}
	}
	payload := shared.Data_Payload{
		Success: true,
		Event:   "user_read",
		Origin:  http_session.UserID,
		Data:    user_profile,
	}
	http_response(w, false, payload)
}

func User_Update(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("HTTPS", fmt.Sprintf("User Update requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	// page := "logout" // * Unused for now
	defer r.Body.Close()
	http_session, err := Validate(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	r.ParseForm()

	display_name := r.PostForm.Get("display_name")
	color := r.PostForm.Get("color_pick")

	pg.ProfileUpdate(active_ws.API.DBconn, http_session.UserID, display_name, color, "")
	user_profile, err := pg.UserRead(active_ws.API.DBconn, http_session.UserID, true)
	if err != nil {
		logger.LogColor("SSE", "Error in finding user - Send")
		return
	}

	payload := shared.Data_Payload{
		Success: true,
		Event:   "user_update",
		Origin:  http_session.UserID,
		Data:    user_profile,
	}
	http_response(w, false, payload)
	active_api.channels.Notify <- payload
}

func User_Update_Status(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("HTTPS", fmt.Sprintf("User Update status requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	// page := "logout" // * Unused for now
	defer r.Body.Close()
	http_session, err := Validate(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	user_profile, err := pg.UserRead(active_ws.API.DBconn, http_session.UserID, true)
	if err != nil {
		logger.LogColor("HTTPS", "User not found, general error.")
	}
	r.ParseForm()

	status := r.PostForm.Get("status")

	if err != nil {
		logger.LogColor("HTTPS", "Can't convert user ID to int")
		return
	}

	pg.UserStatusUpdate(active_ws.API.DBconn, user_profile.ID, status, false)

	payload := shared.Data_Payload{
		Success: true,
		Event:   "status_update",
		Origin:  http_session.UserID,
		Data:    user_profile,
	}
	http_response(w, false, payload)
	active_api.channels.Notify <- payload
}

func User_Update_Avatar(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("HTTPS", fmt.Sprintf("User Profile Picture Update requested from %s", r.RemoteAddr)) // TODO Write Logger() function in core.go
	// page := "logout" // * Unused for now
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		return
	}
	http_session, err := Validate(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	user_profile, err := pg.UserRead(active_api.DBconn, http_session.UserID, true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.ParseMultipartForm(32 << 20)
	file_pic, _, err := r.FormFile("picture")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file_pic.Close()
	f, err := os.OpenFile(fmt.Sprintf("static/data/user_avatars/%s/avatar.png", http_session.UserID), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file_pic)
	w.Header().Add("Cache-control", "no-cache")
	// w.WriteHeader(http.StatusFound)
	payload := shared.Data_Payload{
		Success: true,
		Origin:  http_session.UserID,
		Event:   "status_update",
		Data:    user_profile,
	}
	http_response(w, false, payload)
	active_api.channels.Notify <- payload
}

func User_Read_All(w http.ResponseWriter, r *http.Request) {
	event := "api_get"
	logger.LogColor("HTTPS", fmt.Sprintf("User read ALL requested from %s", r.RemoteAddr))
	// page := "logout" // * Unused for now
	defer r.Body.Close()
	http_session, err := Validate(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	server_users := pg.UserReadAll(active_api.DBconn)

	payload := shared.Data_Payload{
		Success: true,
		Origin:  http_session.UserID,
		Event:   event,
		Data:    server_users,
	}
	http_response(w, false, payload)
}
