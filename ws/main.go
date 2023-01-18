package ws

import (
	"chat/logger"
	"chat/shared"
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
)

var active_ws *WebServer

type WebServer struct {
	// API      *ApiServer
	// SSE      *SSEServer
	URL      string
	SM       *SessionsManager
	WSOCK    *WebSocketServer
	Clients  map[string]chan []byte
	channels shared.Server_Channels
	ip       string
	port     string
	tlsPort  string
	DBconn   *sql.DB
}

func New_WebServer(channels shared.Server_Channels, wsock_dbconn *sql.DB, dbconn *sql.DB, api_dbconn *sql.DB, sse_dbconn *sql.DB, sm_dbconn *sql.DB, ip string, port string, portTls string, url string) (ws *WebServer) {
	ws = &WebServer{
		DBconn:  dbconn,
		Clients: make(map[string]chan []byte),
		URL:     url,
		ip:      ip,
		port:    port,
		tlsPort: portTls,
	}
	logger.LogColor("WEBSRV", "Starting HTTP and HTTPS listeners..")

	// Fileserver declaration
	fs := http.FileServer(http.Dir("./static"))

	// File handlers
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/favicon.ico", favicon_redir)

	// HTTP Handlers
	http.HandleFunc("/", p_Root)
	http.HandleFunc("/login", p_Login)
	http.HandleFunc("/logout", p_Logout)
	http.HandleFunc("/chat", p_Chat)
	http.HandleFunc("/profile", p_Profile)

	http_addr := ip + ":" + string(port)
	https_addr := ip + ":" + string(portTls)
	go http.ListenAndServe(http_addr, http.HandlerFunc(tls_redir))
	logger.LogColor("WEBSRV", "HTTP listener started")

	go http.ListenAndServeTLS(https_addr, "localhost.crt", "localhost.key", nil)
	logger.LogColor("WEBSRV", "HTTPS listener started")

	ws.WSOCK = NewWebSocketServer(wsock_dbconn, channels) //, ip, socketPort, socketTransport)

	// * Starting API Server
	// api_dbconn := pg.NewPostgreConn(IP, Port, Name, Username, Password, SSLMode)
	// ws.API = New_Api(api_dbconn, channels)

	// sse_dbconn := pg.NewPostgreConn(IP, Port, Name, Username, Password, SSLMode)
	// ws.SSE = new_SSEServer(sse_dbconn)

	// * Starting Session Manager
	// sm_dbconn := pg.NewPostgreConn(IP, Port, Name, Username, Password, SSLMode)
	ws.SM = New_Manager(sm_dbconn)

	active_ws = ws
	return
}

func favicon_redir(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/images/favicon.ico")
}
func tls_redir(w http.ResponseWriter, r *http.Request) {
	logger.LogColor("HTTPS", "Redirecting HTTP requests to HTTPS")
	http.Redirect(w, r, ":443", http.StatusSeeOther)
}

func http_response(w http.ResponseWriter, renderTemplate bool, _fileName string, _payload any) {
	if !renderTemplate {
		switch _payload.(type) {
		case shared.Payload:
			payload, err := json.Marshal(_payload)
			if err != nil {
				logger.LogColor("HTTPS", "Marshaling error")
				w.WriteHeader(http.StatusInternalServerError) // ? Which is best to write http.responses? Write or WriteHeader?
			}
			w.Write(payload)
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
		case shared.Payload:
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
