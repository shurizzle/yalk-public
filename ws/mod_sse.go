package ws

import (
	"chat/logger"
	"database/sql"
)

type SSEServer struct {
	DBconn *sql.DB
	// url    string
}

func new_SSEServer(dbconn *sql.DB) (sse *SSEServer) {
	sse = &SSEServer{
		DBconn: dbconn,
	}
	// http.HandleFunc("/events/receive", active_ws.receive)
	logger.LogColor("SSE", "Server ready and listening")
	return
}

// func (sse *SSEServer) runner() {

// }

// func (broker *Broker) watchdog(w http.ResponseWriter, r *http.Request) {
// 	current_time := time.Now().Unix()
// 	fmt.Println(current_time)
// 	if r.Method != http.MethodPost {
// 		logger.LogColor("SSE", "Send - Invalid Method [GET]")
// 		return
// 	}
// 	http_session, err := Validate(w, r)
// 	if err != nil {
// 		http.Redirect(w, r, "/login", http.StatusFound)
// 		return
// 	}
// 	_, err = pg.UserRead(db, http_session.UserID)
// 	if err != nil {
// 		logger.LogColor("SSE", "Error in finding user - Send")
// 		return
// 	}
// }

// func (broker *Broker) close(w http.ResponseWriter, r *http.Request) {

// }
