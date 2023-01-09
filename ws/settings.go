package ws

import (
	"chat/logger"
	"chat/pg"
	"chat/shared"
	"fmt"
	"net/http"
)

func ServerSettings(w http.ResponseWriter, r *http.Request) {
	// context := "/settings"
	event := "api_get"
	logger.LogColor("API", fmt.Sprintf("Server Settings requested from %s", r.RemoteAddr))
	defer r.Body.Close()
	http_session, err := Validate(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	server_settings := pg.ServerSettingsRead(active_api.DBconn)

	payload := shared.Data_Payload{
		Success: true,
		Origin:  http_session.UserID,
		Event:   event,
		Data:    server_settings,
	}
	http_response(w, false, payload)
}
