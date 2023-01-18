package ws

import (
	"chat/logger"
	"chat/pg"
	"chat/shared"
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"nhooyr.io/websocket"
)

// var active_wsock *WebSocketServer

type WebSocketServer struct {
	SubscriberMessageBuffer int
	PublishLimiter          *rate.Limiter
	SubscribersMu           sync.Mutex
	channels                shared.Server_Channels
	Clients                 map[string]*WsSub
	DBconn                  *sql.DB
}

type WsSub struct {
	Msgs      chan []byte
	CloseSlow func()
}

func NewWebSocketServer(db *sql.DB, channels shared.Server_Channels) *WebSocketServer {
	wss := &WebSocketServer{
		DBconn:                  db,
		channels:                channels,
		SubscriberMessageBuffer: 16,
		Clients:                 make(map[string]*WsSub),
		PublishLimiter:          rate.NewLimiter(rate.Every(time.Millisecond*100), 8),
	}

	// active_wsock = wss

	return wss
}

func (websock *WebSocketServer) DeleteSubscriber(s *WsSub, user_id string) {

}

func WriteTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}

func FullInfo(dbconn *sql.DB, user_id string) (payload shared.Payload, err error) {
	_data := make(map[string]any)
	logger.LogColor("WEBSOCK", fmt.Sprintf("Full server info requested from %s", user_id)) // TODO Write Logger() function in core.go
	userAll := pg.UserReadAll(dbconn)
	userSelf, err := pg.UserRead(dbconn, user_id, true)
	if err != nil {
		logger.LogColor("WEBSOCK", "User not found, general error.")
		return payload, err
	}
	serverSettings := pg.ServerSettingsRead(dbconn)
	allChats := pg.ChatReadAll(dbconn, user_id)

	_data["users"] = userAll
	_data["self"] = userSelf
	_data["settings"] = serverSettings
	_data["chats"] = allChats

	payload = shared.Payload{
		Success: true,
		Origin:  user_id,
		Event:   "user_conn",
		Data:    _data,
	}
	return payload, nil
}
