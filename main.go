package main

// ** Server events and meaning ** //

// ** - 'user_conn' -- User connecting to server
// ** - 'user_disconn' -- User disconnecting from server
// ** - 'user_new' -- New user account
// ** - 'user_delete' -- User account deleted
// ** - 'user_update' -- User info update

// ** - 'chat_create' -- New Chat created
// ** - 'chat_delete' -- Chat deleted
// ** - 'chat_message' -- Chat message received
// ** - 'chat_join' -- Chat joined by another user
// ** - 'chat_invite' -- Chat invite received by another user
// ** - 'chat_leave' -- Chat left by another user

import (
	"chat/logger"
	"chat/pg"
	"chat/shared"
	"chat/tcp"
	"chat/ws"
	"encoding/json"
	"io"

	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"nhooyr.io/websocket"
)

var chat_server *server

type server struct {
	instance int
	channels shared.Server_Channels
	modules  server_modules
	ws       *ws.WebServer
	tcp      *tcp.Socket_Server
	dbconn   *sql.DB
}

type server_modules struct {
	webserver  bool
	tcp_server bool
}

func init() {
	godotenv.Load(".env")
	log.Print("\033[H\033[2J")
	var version string = "pre-alpha" // make it os.env
	logger.LogColor("CORE", "Booting..")
	logger.LogColor("CORE", fmt.Sprintf("Chat Server version: %s", version)) // make it os.env
}

func main() {
	var wg sync.WaitGroup

	ip := os.Getenv("HOST_ADDR")
	port := os.Getenv("HTTP_PORT")
	portTls := os.Getenv("HTTPS_PORT")
	url := os.Getenv("WEB_URL")
	socketPort := os.Getenv("SOCKET_PORT")
	socketTransport := os.Getenv("SOCKET_TRANSPORT")

	dbConf := pg.Postgre_Client{
		IP:       os.Getenv("DB_ADDR"),
		Port:     os.Getenv("DB_PORT"),
		DB:       os.Getenv("DB_NAME"),
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	server_dbconn := pg.NewPostgreConn(dbConf)
	chat_server = &server{
		instance: 1,
		channels: shared.Server_Channels{
			Msg:     make(chan shared.Payload, 1),
			Dm:      make(chan map[string]any, 1),
			Notify:  make(chan shared.Payload, 1),
			Cmd:     make(chan shared.Payload),
			Conn:    make(chan shared.Payload),
			Disconn: make(chan shared.Payload),
		},
		modules: server_modules{true, true},
		dbconn:  server_dbconn,
	}

	ws_dbconn := pg.NewPostgreConn(dbConf)
	websock_dbconn := pg.NewPostgreConn(dbConf)
	sm_dbconn := pg.NewPostgreConn(dbConf)
	tcp_dbconn := pg.NewPostgreConn(dbConf)
	chat_server.ws = ws.New_WebServer(chat_server.channels, websock_dbconn, ws_dbconn, nil, nil, sm_dbconn, ip, port, portTls, url)
	chat_server.tcp = tcp.NewSocketServer(chat_server.channels, tcp_dbconn, ip, socketPort, socketTransport)
	// Launch on WaitGrup

	http.HandleFunc("/websocket/connect", chat_server.connectWSOCK)

	go chat_server.runner()
	wg.Add(1)
	wg.Wait()
}

func (server *server) connectWSOCK(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("WEBSOCK", fmt.Sprintf("Requested WebSocket - %s", r.RemoteAddr))
	// event := "websocket_connect"
	http_session, err := ws.Validate(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	user_profile, err := pg.UserRead(server.ws.WSOCK.DBconn, http_session.UserID, true)
	if err != nil {
		logger.LogColor("WEBSOCK", "User not found, general error.")
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		logger.LogColor("WEBSOCK", fmt.Sprintf("Unhautorized session - %s", r.RemoteAddr))
		return
	}

	conn, err := websocket.Accept(w, r, nil) //&websocket.AcceptOptions{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.LogColor("WEBSOCK", fmt.Sprintf("Can't start accepting - %s", r.RemoteAddr))
		return
	}
	defer conn.Close(websocket.StatusInternalError, "Client disconnected")

	notify := r.Context().Done()
	go func() {
		<-notify
		pg.UserStatusUpdate(server.ws.WSOCK.DBconn, http_session.UserID, "offline", false)
		payload := shared.Payload{
			Success: true,
			Origin:  http_session.UserID,
			Event:   "status_update",
			Data:    "offline",
		}
		server.channels.Disconn <- payload
	}()

	client := &ws.WsSub{
		Msgs: make(chan []byte, server.ws.WSOCK.SubscriberMessageBuffer),
		CloseSlow: func() {
			conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
		},
	}
	server.ws.WSOCK.SubscribersMu.Lock()
	server.ws.WSOCK.Clients[http_session.UserID] = client
	server.ws.WSOCK.SubscribersMu.Unlock()

	var localWG sync.WaitGroup

	defer func() {
		go func() {
			server.ws.WSOCK.SubscribersMu.Lock()
			delete(server.ws.WSOCK.Clients, http_session.UserID)
			server.ws.WSOCK.SubscribersMu.Unlock()
		}()
	}()

	ticker := time.NewTicker(time.Second * time.Duration(10))

	// **	Sender - From CLI to SRV	**	//
	localWG.Add(1)
	go func(ticker *time.Ticker) {
		defer localWG.Done()
		for {
			select {
			case <-notify:
				log.Println("Sender - got shutdown signal")
				localWG.Done()
				return
			default:
				t, msg, err := conn.Read(r.Context())
				if err != nil && err != io.EOF {
					statusCode := websocket.CloseStatus(err)
					if statusCode == websocket.StatusGoingAway {
						log.Println("Graceful sender shutdown")
						ticker.Stop()
					} else {
						log.Println("Sender - Error in reading from websocket context, client closed? Check main.go")

					}
					localWG.Done()
					return
				}
				if t.String() == "MessageText" && err == nil {
					err = server.handleUserPayload(msg, http_session.UserID)
					if err != nil {
						log.Printf("Sender - errors in handleUserPayload: %v", err)
						// localWG.Done()
						// return
					}
				}
			}
		}
	}(ticker)

	// **		Receiver from SRV to CLI		**	//
	localWG.Add(1)
	go func(ticker *time.Ticker) {
		defer localWG.Done()
		failedPingCount := 0
		pingTimeoutCount := 5
		pingDelay := 10
		for {
			select {
			case <-notify:
				log.Println("Receiver - got shutdown signal")
				localWG.Done()
				return
			case payload := <-client.Msgs:
				err = ws.WriteTimeout(r.Context(), time.Second*5, conn, payload)
				if err != nil {
					localWG.Done()
					return
				}
			case <-ticker.C:
				if failedPingCount >= pingTimeoutCount {
					fmt.Printf("Client timed out")
					return
				}
				t1 := time.Now().UnixMilli()
				_pingPayload := shared.Payload{Success: true, Event: "ping", Data: t1}
				pingPayload, err := json.Marshal(_pingPayload)
				if err != nil {
					log.Printf("Ping - Marshaling err: %v\n", err)
				}
				err = ws.WriteTimeout(r.Context(), time.Second, conn, pingPayload)
				if err != nil {
					pingDelay += 10
					failedPingCount += 1
					ticker.Reset(time.Duration(time.Second * time.Duration(pingDelay)))
					log.Printf("Ping - Sending err: %v\n", err)
					log.Printf("Ping - Counter now %v\n", failedPingCount)
					log.Printf("Ping - Delay now %vs\n", pingDelay)
				}
			}
		}
	}(ticker)

	logger.LogColor("WEBSOCK", fmt.Sprintf("Connection successful - Username: %v", user_profile.Username))

	// * Initial client payload
	payloadFull, err := ws.FullInfo(server.ws.WSOCK.DBconn, http_session.UserID)
	if err != nil {
		log.Printf("Error in reading full data: %v\n", err)
	}
	greetingsPayload, err := json.Marshal(payloadFull)
	if err != nil {
		log.Printf("Error in initial payload: %v\n", err)
	}
	err = ws.WriteTimeout(r.Context(), time.Second*5, conn, greetingsPayload)

	if err != nil {
		log.Printf("Timeout in initial payload: %v\n", err)
	}
	log.Printf("OK - Full data sent to to ID: %v\n", http_session.UserID)
	localWG.Wait()
	log.Println("After wait, closing connSOCK accept func")
}

// Process and forward in the correct server channel the message received by the client
func (server *server) handleUserPayload(msg []byte, origin string) (err error) {
	var _req any
	var _payload map[string]any
	var payload shared.Payload

	var event string

	err = json.Unmarshal(msg, &_req)
	if err != nil {
		log.Println("Listener - Error deserializing JSON")
		return err
	}

	// * Asserting req types
	switch p := _req.(type) {
	case map[string]any:
		_payload = p
	default:
		log.Println("Listener - can't decode payload")
	}

	switch v := _payload["event"].(type) {
	case string:
		event = v
	default:
		log.Println("Listener - can't decode event")
	}

	payload = shared.Payload{
		Success: true,
		Origin:  origin,
		Event:   event,
	}

	// * Routing event to server
	switch event {
	case "pong":
		switch value := _payload["message"].(type) {
		case float64:
			_t1 := int64(value)
			t1 := time.UnixMilli(_t1)
			// t2 := time.Now().UnixMilli()
			// n, err := shared.Atoi(value)
			if err != nil {
				log.Println("Ping - Err converting to int")
				return err
			}
			log.Printf("User %v ping %v\n", origin, time.Since(t1))

		default:
			log.Println("Pong - can't decode timestamp")
			return err
		}

	case "chat_message":
		var to, text string
		switch v := _payload["to"].(type) {
		case string:
			to = v
		default:
			log.Println("Listener - can't decode chat_id")
			return err
		}

		switch v := _payload["text"].(type) {
		case string:
			text = v
		default:
			log.Println("Listener - can't decode text")
			return err
		}

		if text == "" {
			return fmt.Errorf("empty")
		}

		payload.Data = shared.Chat_Message{
			ID:   pg.MessageCreate(chat_server.dbconn, time.Now().UTC(), text, origin, to),
			Time: time.Now().UTC(),
			From: origin,
			To:   to,
			Text: text,
			Type: "chat_message",
		}

		var chat_info shared.Conversation
		chat_info, err = pg.ChatInfo(server.dbconn, to, true)
		if err != nil {
			log.Println("Listener - chat not found")
			return err
		}
		if chat_info.Type == "channel_public" || chat_info.Type == "channel_private" {
			server.channels.Msg <- payload
			return nil
		}
		if chat_info.Type == "dm" {
			data := map[string]any{"users": chat_info.Users, "payload": payload}
			server.channels.Dm <- data
			return nil
		}
	case "status_update":
		var status string
		var fixed bool
		switch v := _payload["status"].(type) {
		case string:
			status = v
		default:
			log.Println("SendClient - can't decode status")
			return err
		}

		switch v := _payload["fixed"].(type) {
		case bool:
			fixed = v
		default:
			log.Println("SendClient - can't decode statusFixed")
			return err
		}
		err = pg.UserStatusUpdate(server.dbconn, origin, status, fixed)
		if err != nil {
			log.Println("SendClient - can't decode statusFixed")
			return err
		}

		payload.Data, err = pg.UserRead(server.dbconn, origin, true)
		if err != nil {
			log.Println("SendClient - can't read userInfo")
			return err
		}
	case "chat_create":
		var chat_type string
		var name string
		var users []string

		switch value := _payload["name"].(type) {
		case string:
			name = value
		default:
			name = ""
		}

		switch value := _payload["type"].(type) {
		case string:
			chat_type = value
		default:
			if value == "channel_public" && name == "" {
				return err
			}
		}
		switch value := _payload["users"].(type) {
		case string:
			users = append(users, value)
		// case []interface{}:
		case []string:
			users = value
		}

		new_chat_id := pg.ChatCreate(server.ws.WSOCK.DBconn, origin, chat_type, name, users)

		users = append(users, origin)
		new_chat := shared.Conversation{
			ID:    new_chat_id,
			Type:  chat_type,
			Name:  name,
			Users: users,
		}

		payload := shared.Payload{
			Success: true,
			Origin:  origin,
			Event:   event,
			Data:    new_chat,
		}
		server.channels.Notify <- payload

	case "chat_delete":
		var chat_id string
		switch value := _payload["id"].(type) {
		case string:
			chat_id = value
			// default:
			// 	chat_id = ""
		}
		err = pg.ChatDelete(server.ws.WSOCK.DBconn, chat_id)
		if err != nil {
			log.Println("Listener - Error deleting chat")
			return err
		}
		payload := shared.Payload{
			Success: true,
			Origin:  origin,
			Event:   event,
			Data:    chat_id,
		}
		server.channels.Notify <- payload
	case "chat_join":
		var chat_id string
		switch value := _payload["id"].(type) {
		case string:
			chat_id = value
		}
		chat, err := pg.ChatJoin(server.ws.WSOCK.DBconn, origin, chat_id)
		if err != nil {
			logger.LogColor("WEBSOCK", "Error joining chat")
			return err
		}
		payload := shared.Payload{
			Success: true,
			Origin:  origin,
			Event:   event,
			Data:    chat,
		}
		server.channels.Notify <- payload
	case "greeting_message":
		log.Printf("OK - Greetings received from ID %v - payload received: %v", origin, string(msg))
	default:
		log.Println("SendClient - Invalid request")
		payload.Success = false
		return fmt.Errorf("invalid_request")
	}
	return nil
}
