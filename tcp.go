package main

import (
	"bufio"
	"chat/logger"
	"database/sql"
	"fmt"
	"net"

	"golang.org/x/crypto/bcrypt"
)

var activeTCP *tcpServer

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

type tcpServer struct {
	ip               string
	port             string
	transport        string
	DBconn           *sql.DB
	Clients          map[int]*client
	ConnectedClients int
	channels         syncChannels
	listener         net.Listener
	messageChannel   chan dataPayload
	activeUsers      map[string][]byte

	// MaxClients       int
}

type client struct {
	ClientID int
	Socket   net.Conn
	Username string
	// closeConn  bool
	isSetup bool
	channel chan dataPayload
}

func newTCPSerer(channels syncChannels, dbconn *sql.DB, ip string, port string, transport string) (tcp *tcpServer) {
	listener, err := net.Listen(transport, ip+":"+fmt.Sprint(port))
	if err != nil {
		panic(err)
	}
	server := &tcpServer{
		DBconn:         dbconn,
		ip:             ip,
		port:           port,
		transport:      transport,
		Clients:        make(map[int]*client),
		listener:       listener,
		messageChannel: make(chan dataPayload),
	}

	logger.LogColor("SOCKET", fmt.Sprintf("Started listener: %s:%s", ip, port))
	go server.listen()
	go server.run()
	logger.LogColor("SOCKET", "Started accepting for connections")
	return
}

func (s *tcpServer) listen() {
	for {
		conn, err := s.listener.Accept()
		logger.LogColor("SOCKET", "Received new connection")

		// From here the connections is accepted
		if err != nil {
			logger.LogColor("SOCKET", fmt.Sprintf("Error accepting: %v", err))
			return
		}
		// context := "login"
		s.ConnectedClients++
		id := s.ConnectedClients
		c := &client{
			ClientID: id,
			Socket:   conn,
			channel:  s.messageChannel,
		}
		msg := fmt.Sprintf("Assigned id [%d]\n", id)
		c.Socket.Write([]byte(msg))

		logger.LogColor("SOCKET", "Login requested")
		msg = "Username?"
		c.Socket.Write([]byte(msg))
		go c.messageListen()
	}
}

func (s *tcpServer) run() {
	for {
		select {
		case payload := <-s.channels.Conn:
			// case id := <-
			// ? Maybe just s.newMessage <- []byte("new client")
			username := s.activeUsers[payload.Origin]
			welcomeMsg := fmt.Sprintf("%v joined the chat!", []byte(username))
			for _, chatClient := range s.Clients {
				_, err := chatClient.Socket.Write([]byte(welcomeMsg))
				if err != nil {
					logger.LogColor("SOCKET", fmt.Sprintf("Error writing:%v \n%v", welcomeMsg, err))
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
				_creds := dbLogin(activeTCP.DBconn, c.Username)
				err = bcrypt.CompareHashAndPassword([]byte(_creds.Password), []byte(string(buff)))
				if err != nil {
					msg := "Wrong Password, retry"
					c.Socket.Write([]byte(msg))
					continue
				} else {
					token := newUUIDSalted(string(buff))
					session := sessionCreate(token, _creds.ID)
					userProfile, err := userRead(activeTCP.DBconn, _creds.ID, true)
					if err != nil {
						logger.LogColor("SOCKET", "Cannot find user")
					}
					sessionStore(activeTCP.DBconn, session.UserID, session.Expiry, session.Token, session.Created)
					if err != nil {
						logger.LogColor("SOCKET", "Could not create the sessions")
						return
					}
					msg := "Login succesful"
					c.Socket.Write([]byte(msg))
					msg = token
					c.Socket.Write([]byte(msg))
					payload := dataPayload{
						Success: true,
						Event:   "tcp_client_connect",
						Data:    userProfile,
					}
					c.channel <- payload
				}
			}
		default:
			// tcp_session, err := Validate(string(buff))
			// if err == nil {
			// 	return
			// }
			continue
		}
		userProfile, err := userRead(activeTCP.DBconn, "", false) // !!!!! WON'T WORK
		if err != nil {
			panic(err)
		}
		logger.LogColor("SOCKET", fmt.Sprintf("received from %s: %s", userProfile.Username, string(buff)))
		// frame := Payload{
		// 	Payload: Payload{
		// 		User: userProfile,
		// 		Message: Chat_Message{
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
