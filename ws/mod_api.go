package ws

import (
	"chat/logger"
	"chat/shared"
	"database/sql"
	"net/http"
)

var active_api *ApiServer

type ApiServer struct {
	DBconn   *sql.DB
	channels shared.Server_Channels
	// url_mapping map[string]any // "URL, Function for dynamic handler loader"
}

func New_Api(dbconn *sql.DB, channels shared.Server_Channels) (api *ApiServer) {
	logger.LogColor("API", "Starting")

	api = &ApiServer{
		DBconn:   dbconn,
		channels: channels,
	}
	active_api = api

	// Server Settings
	http.HandleFunc("/settings", ServerSettings)

	// User API Handlers
	http.HandleFunc("/user", UserRead)
	http.HandleFunc("/user/all", User_Read_All)
	http.HandleFunc("/user/update", User_Update)
	http.HandleFunc("/user/update/status", User_Update_Status)
	http.HandleFunc("/user/update/avatar", User_Update_Avatar)

	// Admin API Handlers
	http.HandleFunc("/admin/user/add", Admin_User_Add)
	http.HandleFunc("/admin/user/update", Admin_User_Update)
	http.HandleFunc("/admin/user/delete", Admin_User_Delete)

	// Chat API Handlers
	http.HandleFunc("/chat/all", ChatReadAll)
	http.HandleFunc("/chat/create", ChatCreate)
	http.HandleFunc("/chat/join", ChatJoin)
	http.HandleFunc("/chat/delete", ChatDelete)

	// Channel API Handlers
	// http.HandleFunc("/channel", ChannelFull)
	// http.HandleFunc("/channel/all", ChannelFullAll)
	// http.HandleFunc("/channel/get/info", ChannelInfo)
	// // http.HandleFunc("/channel/create", ChannelCreate)
	// http.HandleFunc("/channel/join", ChannelJoin)
	// http.HandleFunc("/channel/joined", ChanJoined)
	// http.HandleFunc("/channel/get/messages", ChannelMessages)

	// DMs API Handlers
	// http.HandleFunc("/dm", DmFull)
	// // http.HandleFunc("/dm/create", DmCreate)
	// http.HandleFunc("/dm/get", DmInfo)
	// http.HandleFunc("/dm/joined", DmReadJoined)
	// http.HandleFunc("/dm/get/messages", DmMessages)

	return
}
