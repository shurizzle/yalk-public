// ? Use this WebWorker solely to act on non messages - only BG activities

function receive(sse_url) {
    const msg_receiver = new EventSource(sse_url, { withCredentials: true })
    msg_receiver.addEventListener('message', (e) => {
        let json_data = JSON.parse(e.data)
        postMessage(json_data)
        // let received_context = json_data['context']
        // let receivedContext = json_data['type']
        // let event_user_id = json_data['origin']
        // if (json_data.success == true) {
        //     switch (json_data.event) {
        //         case "status_update":
        //             break
        //         case "channel_public":
        //         case "channel_private":
        //         case "dm":
                    
        //             break
        //         default:
        //             break
        //     }
        // }
    }, false);
}

self.addEventListener("message", function (e) {
    receive(e.data.args[0])
})