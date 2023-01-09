function dial() {
    const conn = new WebSocket('wss://localhost/websocket/connect')

    conn.addEventListener("close", ev => {
        console.info("WebSocket Disconnected code: " + ev.code + ", reason: " + ev.reason)
        if (ev.code !== 1001) {
            console.info("Reconnecting in 1s")
            setTimeout(dial, 1000)
        }
    })
    conn.addEventListener("open", ev => {
        console.info("websocket connected")
    })

    conn.addEventListener("message", ev => {
        if (typeof ev.data !== "string") {
            console.error("unexpected message type", typeof ev.data)
            return
        }
        postMessage(ev.data)
    })
}

self.addEventListener("message", function (e) {
    receive(e.data.args[0])
})

dial()
