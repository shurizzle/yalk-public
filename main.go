package main

import (
	"chat/logger"
	"chat/pg"
	"chat/shared"
	"chat/tcp"
	"chat/ws"
	"context"
	"errors"
	"io"

	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
	"nhooyr.io/websocket"
)

var chat_server *server

// !!!!!! LEGGERE
// ! https://github.com/golang/lint/issues/179#issuecomment-160277988

// type chat_server interface { // TODO: More appropriate name and learn that crap
// 	Start()
// 	Stop()
// 	runner() // Must be run as goroutine
// }

type server struct {
	instance int
	channels shared.Server_Channels
	modules  server_modules
	ws       *ws.WebServer
	websock  *WebSocketServer
	tcp      *tcp.Socket_Server
	dbconn   *sql.DB
}

type WebSocketServer struct {
	subscriberMessageBuffer int
	publishLimiter          *rate.Limiter
	subscribersMu           sync.Mutex
	channels                shared.Server_Channels
	// SM       *SessionsManager
	Clients map[*wsSub]struct{}
	DBconn  *sql.DB
}

type wsSub struct {
	msgs      chan []byte
	closeSlow func()
}

func NewWebSocketServer(channels shared.Server_Channels, db *sql.DB) *WebSocketServer {
	wss := &WebSocketServer{
		DBconn:                  db,
		channels:                shared.Server_Channels{},
		subscriberMessageBuffer: 16,
		Clients:                 make(map[*wsSub]struct{}),
		publishLimiter:          rate.NewLimiter(rate.Every(time.Millisecond*100), 8),
	}
	return wss
}

func (websock *WebSocketServer) Connect(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("WEBSOCK", fmt.Sprintf("Requested WebSocket - %s", r.RemoteAddr))
	// event := "websocket_connect"

	c, err := websocket.Accept(w, r, nil) //&websocket.AcceptOptions{})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.LogColor("WEBSOCK", fmt.Sprintf("WebSocket can't start accepting - %s", r.RemoteAddr))
		return
	}
	defer c.Close(websocket.StatusInternalError, "Client disconnected")

	err = websock.subscribe(r.Context(), c)
	if errors.Is(err, context.Canceled) {
		logger.LogColor("WEBSOCK", fmt.Sprintf("WebSocket Context Canceled - %s", r.RemoteAddr))
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure || websocket.CloseStatus(err) == websocket.StatusGoingAway {
		logger.LogColor("WEBSOCK", fmt.Sprintf("WebSocket Normal Closure - %s", r.RemoteAddr))
		return
	}
	if err != nil {
		logger.LogColor("WEBSOCK", fmt.Sprintf("WebSocket General Error - %s", r.RemoteAddr))
		return
	}
}

func (websock *WebSocketServer) subscribe(ctx context.Context, c *websocket.Conn) error {
	ctx = c.CloseRead(ctx)

	s := &wsSub{
		msgs: make(chan []byte, websock.subscriberMessageBuffer),
		closeSlow: func() {
			c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
		},
	}
	websock.addSubscriber(s)
	defer websock.deleteSubscriber(s)
	logger.LogColor("WEBSOCK", "Connection successful")

	for {
		select {
		case msg := <-s.msgs:
			err := writeTimeout(ctx, time.Second*5, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (websock *WebSocketServer) publishHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	body := http.MaxBytesReader(w, r.Body, 8192)
	msg, err := io.ReadAll(body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
		return
	}
	websock.publish(msg)

	w.WriteHeader(http.StatusAccepted)
}

func (websock *WebSocketServer) publish(msg []byte) {
	websock.subscribersMu.Lock()
	defer websock.subscribersMu.Unlock()

	websock.publishLimiter.Wait(context.Background())

	for s := range websock.Clients {
		select {
		case s.msgs <- msg:
		default:
			go s.closeSlow()
		}
	}
}

func (websock *WebSocketServer) addSubscriber(s *wsSub) {
	websock.subscribersMu.Lock()
	websock.Clients[s] = struct{}{}
	websock.subscribersMu.Unlock()
}

func (websock *WebSocketServer) deleteSubscriber(s *wsSub) {
	websock.subscribersMu.Lock()
	delete(websock.Clients, s)
	websock.subscribersMu.Unlock()
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
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
			Msg:     make(chan shared.Data_Payload, 1),
			Dm:      make(chan shared.Data_Payload, 1),
			Notify:  make(chan shared.Data_Payload, 1),
			Cmd:     make(chan shared.Data_Payload),
			Conn:    make(chan shared.Data_Payload),
			Disconn: make(chan shared.Data_Payload),
		},
		modules: server_modules{true, true},
		dbconn:  server_dbconn,
	}

	ws_dbconn := pg.NewPostgreConn(dbConf)
	websock_dbconn := pg.NewPostgreConn(dbConf)
	api_dbconn := pg.NewPostgreConn(dbConf)
	sse_dbconn := pg.NewPostgreConn(dbConf)
	sm_dbconn := pg.NewPostgreConn(dbConf)
	tcp_dbconn := pg.NewPostgreConn(dbConf)
	chat_server.ws = ws.New_WebServer(chat_server.channels, ws_dbconn, api_dbconn, sse_dbconn, sm_dbconn, ip, port, portTls, url)
	chat_server.tcp = tcp.NewSocketServer(chat_server.channels, tcp_dbconn, ip, socketPort, socketTransport)
	chat_server.websock = NewWebSocketServer(chat_server.channels, websock_dbconn) //, ip, socketPort, socketTransport)

	// Launch on WaitGrup
	http.HandleFunc("/events/send", chat_server.send)
	http.HandleFunc("/websocket/connect", chat_server.websock.Connect)
	http.HandleFunc("/websocket/send", chat_server.websock.publishHandler)
	http.HandleFunc("/events/receive", chat_server.receive)
	go chat_server.runner()
	wg.Add(1)
	wg.Wait()
}

func (server *server) receive(w http.ResponseWriter, r *http.Request) {
	event := "status_update"
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	http_session, err := ws.Validate(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		logger.LogColor("CORE", "Main HTTP session not found, closing.")
		r.Body.Close()
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", server.ws.URL)
	w.Header().Set("Access-Control-Allow-shared.Credentials", "true")

	chat_server.ws.Clients[http_session.UserID] = make(chan []byte, 1)

	payload := shared.Data_Payload{
		Success: true,
		Origin:  http_session.UserID,
		Event:   event,
	}

	db_status, err := pg.UserStatusRead(server.ws.SSE.DBconn, http_session.UserID)
	if err != nil {
		logger.LogColor("SSE", "Error fetching user status")
		// w.WriteHeader(http.StatusUnauthorized)
		// return
	}
	if db_status != "online" {
		logger.LogColor("SSE", "Updating global user status")
		chat_server.channels.Conn <- payload
	} else {
		logger.LogColor("SSE", "Already online, skipping update status")
	}

	// * This deferred function notifies if the connections dies/crashes
	defer func() {
		go func() {
			// time.Sleep(10 * time.Second)
			pg.UserStatusUpdate(server.ws.SSE.DBconn, http_session.UserID, "offline", false)
			payload.Event = "status_update"
			payload.Data = "offline"
			server.channels.Disconn <- payload
		}()

	}()

	notify := r.Context().Done()

	// * This routine instead exits if the connections is closed by the client
	// ? Maybe it's better to leave the connection dying to not notify everyone?
	go func() {
		<-notify
		server.channels.Disconn <- payload
		// time.Sleep(10 * time.Second)
		pg.UserStatusUpdate(server.ws.SSE.DBconn, http_session.UserID, "offline", false)
		payload.Event = "status_update"
		payload.Data = "offline"
		// * CHECK
	}()
	// ! AND THIS MUST BECOME WEBSOCKETS..
	for {
		// * Why ORiginchann? because here is where the data is pushed to the browser
		fmt.Fprintf(w, "data: %s\n\n", <-server.ws.Clients[http_session.UserID])
		flusher.Flush()

	}
}

func (server *server) send(w http.ResponseWriter, r *http.Request) {
	var err error
	var text string
	// time := time.Now() //.Format("Mon Jan 2 - 15:04")
	// isCommand := false

	if r.Method != http.MethodPost {
		logger.LogColor("SSE", "Send - Invalid Method [GET]")
		return
	}
	http_session, err := ws.Validate(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	user_profile, err := pg.UserRead(server.dbconn, http_session.UserID, true)
	if err != nil {
		logger.LogColor("SSE", "Error in finding user - Send")
		return
	}
	if err := r.ParseForm(); err != nil {
		log.Println("Error parsing request")
		return
	}
	chat_id := r.FormValue("chat_id")
	// type_context := r.FormValue("type_context")
	text = r.FormValue("message")

	if chat_id == "" || text == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	chat_info, err := pg.ChatInfo(server.dbconn, chat_id, true)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload := shared.Data_Payload{
		Origin:  http_session.UserID,
		Success: true,
		Event:   "chat_message",
		Data: shared.Chat_Message{
			ID:        pg.MessageCreate(chat_server.dbconn, time.Now().UTC(), text, user_profile.ID, chat_id),
			Timestamp: time.Now().UTC(),
			From:      fmt.Sprintf("%v", user_profile.ID),
			To:        chat_id,
			Text:      text,
			Type:      "chat_message",
		},
	}

	if chat_info.Type == "channel_public" || chat_info.Type == "channel_private" {
		server.channels.Msg <- payload
		w.WriteHeader(http.StatusOK)
	}

	if chat_info.Type == "dm" {
		payload.Destinations = chat_info.Users
		server.channels.Dm <- payload
		w.WriteHeader(http.StatusOK)
	}
}

func Stop() {
	logger.LogColor("CORE", "Shutting down.")
}
