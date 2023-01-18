package main

import (
	"chat/shared"
	"encoding/json"
	"fmt"
	"log"
)

func (server *server) runner() {
	for {
		select {
		case payload := <-server.channels.Conn:
			fmt.Println("Router: Conn received")
			payload.Event = "status_update"
			server.channels.Notify <- payload

		case payload := <-server.channels.Disconn:
			fmt.Println("Router: Disconn received")
			delete(server.ws.Clients, payload.Origin)
			payload.Event = "status_update"
			server.channels.Notify <- payload

		case payload := <-server.channels.Msg:
			fmt.Println("Router: Mess received")
			_payload, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Marshaling err")
			}
			for _, client_chan := range server.ws.Clients {
				client_chan <- _payload
			}
			for _, wsClient := range server.ws.WSOCK.Clients {
				// fmt.Printf("wsClient: %v\n", wsClient)
				wsClient.Msgs <- _payload
			}

		case _p := <-server.channels.Dm:
			fmt.Println("Router: Dm received")
			dest := _p["users"].([]string)
			payload := _p["payload"].(shared.Payload)
			// TODO: Move Chat logger to send to logger goroutine, broker will manage the instancing of all the goroutines only
			_payload, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Marshaling err")
			}
			// sender := server.ws.Clients[user_id]

			for _, id := range dest {
				sseClient := server.ws.Clients[id]
				wsClient := server.ws.WSOCK.Clients[id]

				if sseClient != nil {
					sseClient <- _payload
				}
				if wsClient != nil {
					wsClient.Msgs <- _payload
				}
			}
			// sender <- Payload

		case _payload := <-server.channels.Notify:
			fmt.Println("Router: Notify received")
			// * Will depend on the event type, to be done
			// payload.Message.MessageID = PgrLogChat(&broker.db, time.Now(), "", string(payload.Message), "server-message", payload.Origin)
			// if payload.Event == "login" {
			// }
			payload, err := json.Marshal(_payload)
			if err != nil {
				log.Printf("Marshaling err")
			}
			for _, client_chan := range server.ws.Clients {
				client_chan <- payload
			}
			for _, wsClient := range server.ws.WSOCK.Clients {
				// fmt.Printf("wsClient: %v\n", wsClient)
				wsClient.Msgs <- payload
			}

			// case payload := <-server.channels.Cmd:
			// 	// TODO: Check user admin status
			// fields := strings.Fields(payload.Data["message"].(map[string]string))
			// command := fields[0]
			// srvMessage := fields[1]
			// //? CHANGE PLACE?
			// switch command {
			// case "/test":
			// 	payload.Event = "server_message"
			// 	payload.Message.Text = srvMessage
			// 	server.channels.Notify <- payload
		}
	}
}
