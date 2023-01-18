let conn

function dial() {
    conn = new WebSocket('wss://localhost:4443/websocket/connect')
    conn.onerror = (ev) => { console.log("Error:", ev) }

    conn.onmessage = (ev) => {
        if (typeof ev.data !== "string") {
            console.error("unexpected message type", typeof ev.data)
        }
        
        let payload = JSON.parse(ev.data)
        if (payload.event == "ping")
        {
            conn.send(JSON.stringify({
                "event": "pong",
                "message": payload.data,
            }))
        }
        postMessage(payload)
    }


    conn.onclose = (ev) => {
        console.log(`WebSocket Disconnected code: ${ev.code}, reason: ${ev.reason}`)
        if (ev.code !== 1001) {
            console.log("Reconnecting in 1s")
            setTimeout(dial, 1000)
        }
    }

    conn.onopen = (ev) => {
        console.log("Connected:", ev)
        let greeting_message = JSON.stringify({
            "event": "greeting_message",
            "message": "test_message",
        })
        conn.send(greeting_message)
    }
}

self.onmessage = (ev) => {
    data = ev.data
    switch (data.event) {
        case 'close_connection':
            conn.close()
            break
        default:
            let payload = JSON.stringify(data)
            conn.send(payload)
            console.info("Message sent", ev.data.message, "to chat", ev.data.chat_id)
    }
}

dial()