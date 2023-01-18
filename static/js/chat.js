import { MessageRow, NewBubbleText, NewHashtagText } from "./elements.js";

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
    if (ylk.Context == '') { ylk.Context = chat.id }

    let ctrChans = document.getElementById('sidebar-channels')
    let ctrDms = document.getElementById('sidebar-dms')
    let oldBtn = document.getElementById(ylk.Context)
    let newBtn = chat.html_button

    ylk.Context = chat.id

    document.querySelector("#send").value = ""
    document.querySelector("#receive").innerHTML = ''


    if (oldBtn != undefined) {
        oldBtn.classList.remove('active')
    }
    if (chat.type == "dm" && ctrDms.children[chat.id] == undefined) {
        ctrDms.append(newBtn)
    }
    if ((chat.type == "channel_public" || chat.type == "channel_private") && ctrChans.children[chat.id] == undefined) {
        ctrChans.append(newBtn)
    }

    newBtn.classList.add('active')

    let hdr = document.getElementById("chat-header")
    let hTitle = document.getElementById("header-title")
    hTitle.innerHTML = ''

    let hDelete = document.getElementById("header-delete")
    let new_hDelete = hDelete.cloneNode(true)
    new_hDelete.addEventListener('click', function (e) {
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

    hdr.replaceChild(new_hDelete, hDelete)
    let receive_area = document.getElementById("receive")

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
            let usrDN = []
            chat.users.forEach(e => {
                usrDN.push(ylk.Users[e].display_name)
            })
            let icoDm = NewBubbleText(usrDN)
            hTitle.appendChild(icoDm.icon_bubble)
            hTitle.appendChild(icoDm.bubble_text)
            break
    }
    let chat_messages = []
    if (chat.messages === null || chat.messages === undefined || Object.keys(chat.messages).length === 0) {
        let message = MessageRow("", 0, "server_message", "RosmoBOT", "Empty", true, "", "0")
        chat_messages.push(message)
    } else {
        let lm = null
        for (const [key, value] of Object.entries(chat.messages)) {
            let user_id = value.from
            let display_name = ylk.Users[user_id].display_name
            let is_admin = ylk.Users[user_id].is_admin
            let color = ylk.Users[user_id].color
            let message = MessageRow(lm, user_id, value.type, display_name, value.text, is_admin, color, value.time, key)
            chat_messages.push(message)
            lm = value
        }
    }
    chat_messages.forEach(element => {
        receive_area.appendChild(element)
    })
    Scroll(receive_area)
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
        ylk.Sock.postMessage(data)
        ylk.Queued.push(data)
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