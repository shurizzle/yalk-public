import { messageRow, NewBubbleText, NewHashtagText } from "./elements.js";

// **** CHAT ****
class chatroom {
    html_button;
    constructor(id, type, name, users, messages) {
        if (messages == null || messages == '') { messages = [] }
        this['id'] = id
        this['type'] = type
        this['name'] = name
        this['users'] = users
        this['messages'] = messages
        this['last'] = 0
    }
}

export function New(res, srv_usr, self_id) {
    let id = res["id"]
    let type = res["type"]
    let name = res["name"]
    let users = res["users"]
    let messages = res["messages"]
    if (res["messages"] != null) { messages = res["messages"] }

    let chat = new chatroom(id, type, name, users, messages)

    let button = document.createElement('button')
    button.id = chat.id
    button.className = 'side-item sidebar-item'

    switch (chat.type) {
        case "dm":
            const self_index = chat.users.indexOf(self_id)
            let _users = chat.users.slice()
            let users_display_names = []

            if (self_index > -1) {
                _users.splice(self_index, 1)
            }
            _users.forEach(element => {
                users_display_names.push(srv_usr[element].display_name)
            })
            var button_items = NewBubbleText(users_display_names.toString())
            button.appendChild(button_items.icon_bubble)
            button.appendChild(button_items.bubble_text)
            break

        case "channel_public":
        case "channel_private":
            var button_items = NewHashtagText(chat.name)
            button.appendChild(button_items.icon_hashtag)
            button.appendChild(button_items.channel_text)
            break
    }

    button.addEventListener('click', function (e) {
        Change(chat)
    })
    chat.html_button = button
    return chat
}

export function Scroll(el) { //? GRAPHIC ELEMENT??
    el.scrollTop = el.scrollHeight
    // if (el.scrollTop != el.scrollHeight() {
    // TODO: This will be needed for detaching live scroll
    // }
}

export function Change(chat) {
    // * Switch global context to current chat id if empty
    if (ylk.Context == '') { ylk.Context = chat.id }
    ylk.Context = chat.id

    let receiveArea = document.getElementById("receive")
    let ctrChans = document.getElementById('sidebar-channels')
    let ctrDms = document.getElementById('sidebar-dms')
    let oldBtn = document.getElementById(ylk.Context)
    let hdr = document.getElementById("chat-header")
    let hTitle = document.getElementById("header-title")
    let hDelete = document.getElementById("header-delete")
    let newhDelete = hDelete.cloneNode(true)
    let chatMessage = []

    // * Defining delete function with input chat data
    newhDelete.addEventListener('click', function (e) {
        let data = {
            "event": "chat_delete",
            "id": chat.id,
        }
        ylk.Sock.postMessage(data)
        ylk.Queued.push(data)
        // requestJson("POST", "/chat/delete", true, { "id": chat.id }).then((v) => {
        // TODO: Let the worker take care of this
        let btnChat = document.getElementById(data.id)
        if (btnChat != undefined) {
            btnChat.remove()
        }
        const index = ylk.Self.joined_chats.indexOf(data.id)
        if (index > -1) {
            ylk.Self.joined_chats.splice(index, 1)
        }
        delete ylk.Chats[data.id]
    })

    // * Empty send, receive and header areas
    let send = document.getElementById('send')
    send.value = ""
    send.style.visibility = "visible"
    document.querySelector("#receive").innerHTML = ''
    hTitle.innerHTML = ''

    // * Create new chat button
    // * Remove active status from previous button
    let newBtn = chat.html_button
    newBtn.classList.add('active')
    if (oldBtn != undefined) {
        oldBtn.classList.remove('active')
    }

    // * Appends to DOM buttons if not already there
    if (chat.type == "dm" && ctrDms.children[chat.id] == undefined) {
        ctrDms.append(newBtn)
    }
    if ((chat.type == "channel_public" || chat.type == "channel_private") && ctrChans.children[chat.id] == undefined) {
        ctrChans.append(newBtn)
    }

    // * Determining chat type for custom icon in header
    switch (chat.type) {
        case "channel_public":
        case "channel_private":
            let icoChan = NewHashtagText(chat.name)
            hTitle.appendChild(icoChan.icon_hashtag)
            hTitle.appendChild(icoChan.channel_text)
            break

        case "dm":
            const i = chat.users.indexOf(ylk.Self.user_id)
            if (i > -1) {
                chat.users.splice(i, 1)
            }
            let displayName = []
            chat.users.forEach(e => {
                displayName.push(ylk.Users[e].display_name)
            })
            let icoDm = NewBubbleText(displayName)
            hTitle.appendChild(icoDm.icon_bubble)
            hTitle.appendChild(icoDm.bubble_text)
            break
    }

    hdr.replaceChild(newhDelete, hDelete)

    if (chat.messages === null || chat.messages === undefined || Object.keys(chat.messages).length === 0) {
        let message = messageRow("", 0, "server_message", "RosmoBOT", "Empty", true, "", "0")
        chatMessage.push(message)
    } else {
        let lm = null
        for (const [key, value] of Object.entries(chat.messages)) {
            let user_id = value.from
            let display_name = ylk.Users[user_id].display_name
            let isAdmin = ylk.Users[user_id].isAdmin
            let color = ylk.Users[user_id].color
            let message = messageRow(lm, user_id, value.type, display_name, value.text, isAdmin, color, value.time, key)
            chatMessage.push(message)
            lm = value
        }
        chat.last = lm.message_id
    }

    chatMessage.forEach(element => {
        receiveArea.appendChild(element)
    })
    Scroll(receiveArea)
}

export function Open(chType, chName, chUsers, srvData, _chId) {
    let data
    let reqChat = srvData.Chats[_chId]

    // * If I can find the chat between all those I have fetched (backend only gives ones I am allowed to see)
    if (reqChat != undefined) {
        // * If user data shows it's joined already, just switch to the chat
        if (srvData.Self.joined_chats.includes(reqChat.id)) {
            data = reqChat
            return data
        }

        if (chType == "dm") {
            let _users = chUsers.push(srvData.Self.user_id)
            for (const [_, value] of Object.entries(srvData.Chats)) {
                if (value.users.every(user => _users.includes(user))){
                    return value
                }
            }
        }
        // * Else ask the server to join it
        data = {
            "event": "chat_join",
            "id": _chId,
        }
        ylk.Queued.push(data)
        ylk.Sock.postMessage(data)
        return null
        // let _res = await requestJson("POST", "/chat/join", true, { "id": _chId })
        // res = ylk.Chats[_chId]
        // ylk.Self.joined_chats.push(res.id)
        // if (res.data[_chId] != undefined) {

        // }
    }

    data = {
        "event": "chat_create",
        "type": chType,
        "users": chUsers,
        "name": chName
    }
    ylk.Sock.postMessage(data)
    ylk.Queued.push(data)
    return null

    // let _res = await requestJson("POST", "/chat/create", true, data)
    // _res['data']['origin'] = _res["origin"]=
    // if (document.getElementById(res.id) == null) { 
    // if (document.querySelector("#"+res.id).parentElement.id == 'search-popup-content') {
    // ? Approach of QuerySelectorAll and checking parentElement.id seems better, less code and can be done with a switch statement.
    // let ctrChan = document.getElementById('sidebar-channels')
    // let ctrDms = document.getElementById('sidebar-dms')

    // if (res.type == "dm" && ctrDms.children[res.id] == undefined) {
    //     ctrDms.append(res.html_button)
    // }
    // if ((res.type == "channel_public" || res.type == "channel_private") && ctrChan.children[res.id] == undefined) {
    //     ctrChan.append(res.html_button)
    // }
}