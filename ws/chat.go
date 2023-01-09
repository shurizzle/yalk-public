package ws

import (
	"chat/logger"
	"chat/pg"
	"chat/shared"
	"encoding/json"
	"fmt"
	"net/http"
)

func ChatCreate(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("API", fmt.Sprintf("CHAT - NEW - %s", r.RemoteAddr))
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		logger.LogColor("API", fmt.Sprintf("[405] CHAT - NEW - %s", r.RemoteAddr))
		return
	}
	event := "chat_create"
	defer r.Body.Close()

	http_session, err := Validate(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		logger.LogColor("API", fmt.Sprintf("[401] CHAT - NEW - %s", r.RemoteAddr))
		return
	}
	var r_payload map[string]any
	err = json.NewDecoder(r.Body).Decode(&r_payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.LogColor("API", fmt.Sprintf("[400] CHAT - NEW - %s", r.RemoteAddr))
		return
	}

	var chat_type string
	var name string
	var users []string

	switch value := r_payload["name"].(type) {
	case string:
		name = value
	default:
		name = ""
	}

	switch value := r_payload["chat_type"].(type) {
	case string:
		chat_type = value
	default:
		if value == "channel_public" && name == "" {
			w.WriteHeader(http.StatusBadRequest)
			logger.LogColor("API", fmt.Sprintf("[400] CHAT - NEW - %s", r.RemoteAddr))
			return
		}
	}
	switch value := r_payload["users"].(type) {
	case string:
		users = append(users, value)
	// case []interface{}:
	case []string:
		users = value
	}

	// 	// w.WriteHeader(http.StatusBadRequest)
	// 	// logger.LogColor("API", fmt.Sprintf("[400] CHAT - NEW - %s", r.RemoteAddr))
	// 	// return
	new_chat_id := pg.ChatCreate(active_ws.API.DBconn, http_session.UserID, chat_type, name, users)

	users = append(users, http_session.UserID)
	new_chat := shared.Conversation{
		ID:    new_chat_id,
		Type:  chat_type,
		Name:  name,
		Users: users,
	}

	payload := shared.Data_Payload{
		Success: true,
		Origin:  http_session.UserID,
		Event:   event,
		Data:    new_chat,
	}
	http_response(w, false, payload)
	logger.LogColor("API", fmt.Sprintf("[200] CHAT - NEW - %s", r.RemoteAddr))
	active_api.channels.Notify <- payload
}

func ChatReadAll(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("API", fmt.Sprintf("CHAT_ALL - GET - from %s", r.RemoteAddr))
	event := "api_get"
	defer r.Body.Close()
	http_session, err := Validate(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		logger.LogColor("API", fmt.Sprintf("[401] CHAT - NEW - %s", r.RemoteAddr))
		return
	}

	all_chats := pg.ChatReadAll(active_ws.API.DBconn, http_session.UserID)
	payload := shared.Data_Payload{
		Success: true,
		Origin:  http_session.UserID,
		Event:   event,
		Data:    all_chats,
	}
	http_response(w, false, payload)
}

func ChatDelete(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("API", fmt.Sprintf("CHAT - DELETE - from %s", r.RemoteAddr))
	defer r.Body.Close()
	event := "chat_delete"
	http_session, err := Validate(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		logger.LogColor("API", fmt.Sprintf("[401] CHAT - DELETE - %s", r.RemoteAddr))
		return
	}
	if r.Method != http.MethodPost { // !!!!! change in delete
		w.WriteHeader(http.StatusMethodNotAllowed)
		logger.LogColor("API", fmt.Sprintf("[405] CHAT - DELETE - %s", r.RemoteAddr))
		return
	}
	var r_payload map[string]any
	err = json.NewDecoder(r.Body).Decode(&r_payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.LogColor("API", fmt.Sprintf("[400] CHAT - DELETE - %s", r.RemoteAddr))
		return
	}

	var chat_id string
	switch value := r_payload["id"].(type) {
	case string:
		chat_id = value
		// default:
		// 	chat_id = ""
	}
	err = pg.ChatDelete(active_api.DBconn, chat_id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.LogColor("API", fmt.Sprintf("[500] CHAT - DELETE - %s", r.RemoteAddr))
		return
	}

	payload := shared.Data_Payload{
		Success: true,
		Origin:  http_session.UserID,
		Event:   event,
		Data:    chat_id,
	}
	http_response(w, false, payload)
	logger.LogColor("API", fmt.Sprintf("[200] CHAT - DELETE - %s", r.RemoteAddr))
	active_api.channels.Notify <- payload
}

func ChatJoin(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("API", fmt.Sprintf("Chat join from %s", r.RemoteAddr))
	event := "chat_join"
	defer r.Body.Close()
	http_session, err := Validate(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var r_payload map[string]any
	err = json.NewDecoder(r.Body).Decode(&r_payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.LogColor("API", fmt.Sprintf("[400] CHAT - DELETE - %s", r.RemoteAddr))
		return
	}

	var chat_id string
	switch value := r_payload["id"].(type) {
	case string:
		chat_id = value
	}

	chat, err := pg.ChatJoin(active_api.DBconn, http_session.UserID, chat_id)
	if err != nil {
		logger.LogColor("HTTPS", "Getting chat error")
		w.WriteHeader(http.StatusInternalServerError) // ? Which is best to write http.responses? Write or WriteHeader?
		return
	}

	payload := shared.Data_Payload{
		Success: true,
		Origin:  http_session.UserID,
		Event:   event,
		Data:    chat,
	}
	http_response(w, false, payload)
	logger.LogColor("API", fmt.Sprintf("[200] CHAT - JOIN - %s", r.RemoteAddr))
	active_api.channels.Notify <- payload
}
