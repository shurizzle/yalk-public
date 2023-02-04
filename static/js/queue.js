import { DebugBox, DragElement } from "./ngula/debug.js"
import { UserRow, UserClass } from "./user.js"
import * as chat from "./chat.js"

let sidebar = document.getElementById('sidebar')
let usersContainer = document.getElementById('server-users')

export function handleEventsQueue(payload) {
    // if (ylk.Queued.length == 0) {
    //     return
    // }
    ylk.Queued.forEach((item, _) => {
        if (item.event != payload.event) {
            return // ! ???????
        }
        let payloadData = payload.data
        if (payload.event == "init") {
            init(payloadData)
        }
        if (payload.event == "chat_create") {
            chatCreate(payload)
        }
        if (payload.event == "chat_delete") {
            chatDelete(payload)
        }
        if (payload.event == "chat_join") {
            chatJoin(payloadData)
        }
        // if (payload.event == "user_update") {
        //     userUpdate(payload)
        // }
    })
}

function init(payloadData) {
    for (const [key, value] of Object.entries(payloadData)) {
        if (key === "users") {
            for (const [k, v] of Object.entries(value)) {
                let _cu = UserClass(v)
                ylk.Users[k] = v
                let _usrRow = UserRow(false, _cu)
                usersContainer.appendChild(_usrRow)
            }
        }
        if (key === "self") {
            ylk.Self = UserClass(value)
            sidebar.appendChild(UserRow(true, value))
            if (ylk.Self.isAdmin == "true") {
                let debugBox = DebugBox()
                document.body.prepend(debugBox)
                DragElement(debugBox)
            }
        }
        if (key === "settings") {
            for (const [k, v] of Object.entries(value)) {
                ylk.Settings[k] = v
            }
        }
        if (key === "chats") {
            for (const [k, v] of Object.entries(value)) {
                ylk.Chats[k] = chat.New(v, payloadData["users"], payloadData["self"].user_id)
            }
        }
    }
    ylk.Self.joined_chats.forEach(element => {
        let chat = ylk.Chats[element]
        if (chat.type === "channel_public" || chat.type === "channel_private") {
            document.querySelector('#sidebar-channels').append(chat.html_button)
        }
        if (chat.type === "dm") {
            document.querySelector('#sidebar-dms').append(chat.html_button)
        }
    })
    chat.Change(ylk.Chats[ylk.Settings.default_channel])
    document.querySelector('#loading-screen').style.display = 'none'
    return
}

function chatCreate(payload) {
    if (payload.data.type != item.type && payload.origin != ylk.Self.user_id && payload.data.name != item.name) {
        return
    }
    ylk.Queued.splice(index, 1)
    ylk.Chats[payload.data.id] = chat.New(payload.data, ylk.Users, ylk.Self.user_id)
    chat.Change(ylk.Chats[payload.data.id])
}

function chatDelete(payload) {
    if (payload.data.type != item.type && payload.origin != ylk.Self.user_id && payload.data.name != item.name) {
        return
    }
    ylk.Queued.splice(index, 1)
    chat.Change(ylk.Chats[ylk.Settings.default_channel])
}

function chatJoin(payloadData) {
    let chat = ylk.Chats[payloadData.id]
    if (payloadData.id == item.id) {
        ylk.Self.joined_chats.push(payloadData.id)
    }
    if (payloadData.type == "channel_public" || payloadData.type == "channel_private") {
        document.querySelector('#sidebar-channel}}s').append(chat.html_button)
    } else {
        document.querySelector('#sidebar-dms').append(chat.html_button)
    }
    chat.Change(chat)
}

function userUpdate(payload) {
    if (payload.origin !== ylk.Self.user_id) {
        return
    }
    const index = ylk.Queued.indexOf(payload.event)
    if (index > -1) {
        ylk.Queued.splice(index, 1)
    }
}