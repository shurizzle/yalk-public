package tcp

import (
	"bufio"
	"chat/logger"
	"chat/pg"
	"chat/shared"
	"database/sql"
	"fmt"
	"net"

	"golang.org/x/crypto/bcrypt"
)

var active_tcp *Socket_Server

// type TCPServer struct {
// 	ServerSettings   ServerSettings
// 	MaxClients       int
// 	ConnectedClients int
// 	clients          map[int]*ChatClient
// 	activeUsers      map[int][]byte
// 	ServerListener   net.Listener
// 	newMessage       chan []byte
// 	newClient        chan int
// 	shutDown         chan bool
// }

type Socket_Server struct {
	ip               string
	port             string
	transport        string
	DBconn           *sql.DB
	Clients          map[int]*client
	ConnectedClients int
	channels         shared.Server_Channels
	listener         net.Listener
	message_channel  chan shared.Payload
	activeUsers      map[string][]byte

	// MaxClients       int
}

type client struct {
	ClientID int
	Socket   net.Conn
	Username string
	// closeConn  bool
	isSetup         bool
	message_channel chan shared.Payload
}

func NewSocketServer(channels shared.Server_Channels, dbconn *sql.DB, ip string, port string, transport string) (tcp *Socket_Server) {
	new_listener, err := net.Listen(transport, ip+":"+fmt.Sprint(port))
	if err != nil {
		panic(err)
	}
	server := &Socket_Server{
		DBconn:          dbconn,
		ip:              ip,
		port:            port,
		transport:       transport,
		Clients:         make(map[int]*client),
		listener:        new_listener,
		message_channel: make(chan shared.Payload),
	}

	logger.LogColor("SOCKET", fmt.Sprintf("Started listener: %s:%s", ip, port))
	go server.listen()
	go server.run()
	logger.LogColor("SOCKET", "Started accepting for connections")
	return
}

func (s *Socket_Server) listen() {
	for {
		conn, err := s.listener.Accept()
		logger.LogColor("SOCKET", "Received new connection")

		// From here the connections is accepted
		if err != nil {
			logger.LogColor("SOCKET", fmt.Sprintf("Error accepting: %v", err))
			return
		}
		// context := "login"
		s.ConnectedClients += 1
		id := s.ConnectedClients
		c := &client{
			ClientID:        id,
			Socket:          conn,
			message_channel: s.message_channel,
		}
		msg := fmt.Sprintf("Assigned id [%d]\n", id)
		c.Socket.Write([]byte(msg))

		logger.LogColor("SOCKET", "Login requested")
		msg = "Username?"
		c.Socket.Write([]byte(msg))
		go c.messageListen()
	}
}

func (s *Socket_Server) run() {
	for {
		select {
		case payload := <-s.channels.Conn:
			// case id := <-
			// ? Maybe just s.newMessage <- []byte("new client")
			username := s.activeUsers[payload.Origin]
			welcome_msg := fmt.Sprintf("%v joined the chat!", []byte(username))
			for _, chat_client := range s.Clients {
				_, err := chat_client.Socket.Write([]byte(welcome_msg))
				if err != nil {
					logger.LogColor("SOCKET", fmt.Sprintf("Error writing:%v \n%v", welcome_msg, err))
				}
			}

		// case frame := <-s.channels.Msg:
		// 	for _, chat_client := range s.Clients {
		// 		_, err := chat_client.Socket.Write([]byte(payload.Data.(string)))
		// 		if err != nil {
		// 			logger.LogColor("SOCKET", fmt.Sprintf("Error writing:%v \n%v", payload.Message.Text, err))
		// 		}
		// }

		// case <-s.shutDown:
		// 	log.Print("Shutting down server")
		// 	s.ServerListener.Close()
		// 	return
		default:
			continue
		}
	}
}

func (c *client) messageListen() {
	for {
		br := bufio.NewReader(c.Socket)
		buff, err := br.ReadBytes('\n')
		if err != nil {
			logger.LogColor("SOCKET", fmt.Sprintf("Error reading: %s", err.Error()))
		}
		switch {
		case !c.isSetup:
			if c.Username == "" {
				c.Username = string(buff)
				msg := "Password?"
				c.Socket.Write([]byte(msg))
				continue
			} else {
				db_creds := pg.Login(active_tcp.DBconn, c.Username)
				err = bcrypt.CompareHashAndPassword([]byte(db_creds.Password), []byte(string(buff)))
				if err != nil {
					msg := "Wrong Password, retry"
					c.Socket.Write([]byte(msg))
					continue
				} else {
					session_token := GenerateSaltedUUID(string(buff))
					new_session := New_Session(session_token, db_creds.ID)
					user_profile, err := pg.UserRead(active_tcp.DBconn, db_creds.ID, true)
					if err != nil {
						logger.LogColor("SOCKET", "Cannot find user")
					}
					pg.SessionsCreate(active_tcp.DBconn, new_session.UserID, new_session.Expiry, new_session.Token, new_session.Created)
					if err != nil {
						logger.LogColor("SOCKET", "Could not create the sessions")
						return
					}
					msg := "Login succesful"
					c.Socket.Write([]byte(msg))
					msg = session_token
					c.Socket.Write([]byte(msg))
					payload := shared.Payload{
						Success: true,
						Event:   "tcp_client_connect",
						Data:    user_profile,
					}
					c.message_channel <- payload
				}
			}
		default:
			// tcp_session, err := Validate(string(buff))
			// if err == nil {
			// 	return
			// }
			continue
		}
		user_profile, err := pg.UserRead(active_tcp.DBconn, "", false) // !!!!! WON'T WORK
		if err != nil {
			panic(err)
		}
		logger.LogColor("SOCKET", fmt.Sprintf("received from %s: %s", user_profile.Username, string(buff)))
		// frame := shared.Payload{
		// 	Payload: shared.Payload{
		// 		User: user_profile,
		// 		Message: shared.Chat_Message{
		// 			Text: string(buff),
		// 		},
		// 	},
		// }
		// c.message_channel <- frame
		// if c.closeConn {
		// 	break
		// }
	}

}
