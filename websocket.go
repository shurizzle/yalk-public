package main

import (
	"chat/logger"
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"nhooyr.io/websocket"
)

// var active_wsock *WebSocketServer

type websocketServer struct {
	SubscriberMessageBuffer int
	PublishLimiter          *rate.Limiter
	SubscribersMu           sync.Mutex
	channels                syncChannels
	Clients                 map[string]*websocketClient
	DBconn                  *sql.DB
}

type websocketClient struct {
	Msgs      chan []byte
	CloseSlow func()
}

func newWebsocketServer(db *sql.DB, channels syncChannels) *websocketServer {
	wss := &websocketServer{
		DBconn:                  db,
		channels:                channels,
		SubscriberMessageBuffer: 16,
		Clients:                 make(map[string]*websocketClient),
		PublishLimiter:          rate.NewLimiter(rate.Every(time.Millisecond*100), 8),
	}

	// active_wsock = wss

	return wss
}

func (websock *websocketServer) DeleteSubscriber(s *websocketClient, id string) {

}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}

func fullData(dbconn *sql.DB, id string) (payload dataPayload, err error) {
	_data := make(map[string]any)
	logger.LogColor("WEBSOCK", fmt.Sprintf("Full server info requested from %s", id)) // TODO Write Logger() function in core.go
	userAll := userReadAll(dbconn)
	userSelf, err := userRead(dbconn, id, true)
	if err != nil {
		logger.LogColor("WEBSOCK", "User not found, general error.")
		return payload, err
	}
	serverSettings := serverSettingsRead(dbconn)
	allChats := chatReadall(dbconn, id)

	_data["users"] = userAll
	_data["self"] = userSelf
	_data["settings"] = serverSettings
	_data["chats"] = allChats

	payload = dataPayload{
		Success: true,
		Origin:  id,
		Event:   "init",
		Data:    _data,
	}
	return payload, nil
}
