package main

import (
	"encoding/json"
	"log"
)

func (server *server) runner() {
	for {
		select {
		case payload := <-server.channels.Conn:
			payload.Event = "status_update"
			server.channels.Notify <- payload

		case payload := <-server.channels.Disconn:
			delete(server.ws.Clients, payload.Origin)
			payload.Event = "status_update"
			server.channels.Notify <- payload

		case payload := <-server.channels.Msg:
			_payload, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Marshaling err")
			}
			for _, client_chan := range server.ws.Clients {
				client_chan <- _payload
			}

		case payload := <-server.channels.Dm:
			// user_id := payload.Origin
			dest := payload.Destinations

			// TODO: Move Chat logger to send to logger goroutine, broker will manage the instancing of all the goroutines only
			Payload, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Marshaling err")
			}
			// sender := server.ws.Clients[user_id]

			for _, id := range dest {
				receiver := server.ws.Clients[id]
				if receiver != nil {
					receiver <- Payload
				}
			}
			// sender <- Payload

		case payload := <-server.channels.Notify:
			// * Will depend on the event type, to be done
			// payload.Message.MessageID = PgrLogChat(&broker.db, time.Now(), "", string(payload.Message), "server-message", payload.Origin)
			// if payload.Event == "login" {
			// }
			Payload, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Marshaling err")
			}
			for _, client_chan := range server.ws.Clients {
				client_chan <- Payload
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
